package notifier

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/maddevsio/comedian/model"

	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/sirupsen/logrus"
)

// Notifier struct is used to notify users about upcoming or skipped standups
type Notifier struct {
	s    *chat.Slack
	db   storage.Storage
	conf config.Config
}

// NewNotifier creates a new notifier
func NewNotifier(slack *chat.Slack) (*Notifier, error) {
	notifier := &Notifier{s: slack, db: slack.DB, conf: slack.Conf}
	return notifier, nil
}

// Start starts all notifier treads
func (n *Notifier) Start() error {
	notificationForChannels := time.NewTicker(time.Second * 60).C
	notificationForTimeTable := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-notificationForChannels:
			n.NotifyChannels()
		case <-notificationForTimeTable:
			n.NotifyIndividuals()
		}
	}
}

// NotifyChannels reminds users of channels about upcoming or missing standups
func (n *Notifier) NotifyChannels() {
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	channels, err := n.db.GetChannels()
	if err != nil {
		logrus.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
		return
	}
	// For each standup time, if standup time is now, start reminder
	for _, channel := range channels {
		if channel.StandupTime == 0 {
			continue
		}
		standupTime := time.Unix(channel.StandupTime, 0)
		warningTime := time.Unix(channel.StandupTime-n.conf.ReminderTime*60, 0)
		if time.Now().Hour() == warningTime.Hour() && time.Now().Minute() == warningTime.Minute() {
			n.SendWarning(channel.ChannelID)
		}
		if time.Now().Hour() == standupTime.Hour() && time.Now().Minute() == standupTime.Minute() {
			go n.SendChannelNotification(channel.ChannelID)
		}
	}
}

// NotifyIndividuals reminds users of channels about upcoming or missing standups
func (n *Notifier) NotifyIndividuals() {
	day := strings.ToLower(time.Now().Weekday().String())
	tts, err := n.db.ListTimeTablesForDay(day)
	if err != nil {
		logrus.Errorf("ListTimeTablesForToday failed: %v", err)
		return
	}

	for _, tt := range tts {
		standupTime := time.Unix(tt.ShowDeadlineOn(day), 0)
		warningTime := time.Unix(tt.ShowDeadlineOn(day)-n.conf.ReminderTime*60, 0)

		if time.Now().Hour() == warningTime.Hour() && time.Now().Minute() == warningTime.Minute() {
			n.SendIndividualWarning(tt.ChannelMemberID)
		}
		if time.Now().Hour() == standupTime.Hour() && time.Now().Minute() == standupTime.Minute() {
			go n.SendIndividualNotification(tt.ChannelMemberID)
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (n *Notifier) SendWarning(channelID string) {
	allNonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	nonReporters := []model.ChannelMember{}
	for _, u := range allNonReporters {
		if !n.db.MemberHasTimeTable(u.ID) {
			nonReporters = append(nonReporters, u)
		}
	}
	if len(nonReporters) == 0 {
		return
	}
	nonReportersIDs := []string{}
	for _, user := range nonReporters {
		nonReportersIDs = append(nonReportersIDs, "<@"+user.UserID+">")
	}
	err = n.s.SendMessage(channelID, fmt.Sprintf(n.conf.Translate.NotifyUsersWarning, strings.Join(nonReportersIDs, ", "), n.conf.ReminderTime))
	if err != nil {
		logrus.Errorf("notifier: n.s.SendMessage failed: %v\n", err)
		return
	}
}

// SendIndividualWarning reminds users in chat about upcoming standups
func (n *Notifier) SendIndividualWarning(channelMemberID int64) {
	chm, err := n.db.SelectChannelMember(channelMemberID)
	if err != nil {
		logrus.Errorf("SelectChannelMember failed: %v", err)
		return
	}
	submittedStandup := n.db.SubmittedStandupToday(chm.UserID, chm.ChannelID)
	if !submittedStandup {
		err = n.s.SendMessage(chm.ChannelID, fmt.Sprintf(n.conf.Translate.IndividualStandupersWarning, chm.UserID, n.conf.ReminderTime))
		if err != nil {
			logrus.Errorf("notifier: n.s.SendMessage failed: %v\n", err)
			return
		}
		return
	}
	logrus.Infof("%v is not non reporter", chm.UserID)
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendChannelNotification(channelID string) {
	members, err := n.db.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.db.ListChannelMembers failed: %v\n", err)
		return
	}
	if len(members) == 0 {
		logrus.Info("No standupers in this channel\n")
		return
	}
	allNonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	nonReporters := []model.ChannelMember{}

	for _, u := range allNonReporters {
		if !n.db.MemberHasTimeTable(u.ID) {
			nonReporters = append(nonReporters, u)
		}
	}
	logrus.Infof("NON REPORTERS: %v", nonReporters)
	// if everyone wrote their standups display all done message!
	if len(nonReporters) == 0 {
		err := n.s.SendMessage(channelID, n.conf.Translate.NotifyAllDone)
		if err != nil {
			logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
		}
		return
	}

	channel, err := n.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
	}

	// othervise Direct Message non reporters
	for _, nonReporter := range nonReporters {
		err := n.s.SendUserMessage(nonReporter.UserID, fmt.Sprintf(n.conf.Translate.NotifyDirectMessage, nonReporter.UserID, channel.ChannelID, channel.ChannelName))
		if err != nil {
			logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
		}
	}

	repeats := 0

	notifyNotAll := func() error {
		allNonReporters, err := n.getCurrentDayNonReporters(channelID)
		if err != nil {
			logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
			return err
		}
		nonReporters := []model.ChannelMember{}

		for _, u := range allNonReporters {
			if !n.db.MemberHasTimeTable(u.ID) {
				nonReporters = append(nonReporters, u)
			}
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.UserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		if repeats < n.conf.ReminderRepeatsMax && len(nonReporters) > 0 {
			n.s.SendMessage(channelID, fmt.Sprintf(n.conf.Translate.NotifyNotAll, strings.Join(nonReportersSlackIDs, ", ")))
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		//n.notifyAdminsAboutNonReporters(channelID, nonReportersSlackIDs)
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(n.conf.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

//SendIndividualNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendIndividualNotification(channelMemberID int64) {
	chm, err := n.db.SelectChannelMember(channelMemberID)
	if err != nil {
		logrus.Errorf("SelectChannelMember failed: %v", err)
		return
	}
	channel, err := n.db.SelectChannel(chm.ChannelID)
	if err != nil {
		logrus.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
	}
	submittedStandup := n.db.SubmittedStandupToday(chm.UserID, chm.ChannelID)
	if !submittedStandup {
		err := n.s.SendUserMessage(chm.UserID, fmt.Sprintf(n.conf.Translate.NotifyDirectMessage, chm.UserID, channel.ChannelID, channel.ChannelName))
		if err != nil {
			logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
		}
	}
	repeats := 0
	notify := func() error {
		submittedStandup := n.db.SubmittedStandupToday(chm.UserID, chm.ChannelID)
		if repeats < n.conf.ReminderRepeatsMax && !submittedStandup {
			n.s.SendMessage(channel.ChannelID, fmt.Sprintf(n.conf.Translate.IndividualStandupersLate, chm.UserID))
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		logrus.Infof("User %v submitted standup!", chm.UserID)
		return nil
	}
	b := backoff.NewConstantBackOff(time.Duration(n.conf.NotifierInterval) * time.Minute)
	err = backoff.Retry(notify, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

// getNonReporters returns a list of standupers that did not write standups
func (n *Notifier) getCurrentDayNonReporters(channelID string) ([]model.ChannelMember, error) {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	nonReporters, err := n.db.GetNonReporters(channelID, timeFrom, time.Now())
	if err != nil && err != errors.New("no rows in result set") {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}
