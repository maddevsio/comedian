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
	Chat   chat.Chat
	db     storage.Storage
	Config config.Config
}

// NewNotifier creates a new notifier
func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	notifier := &Notifier{Chat: chat, db: conn, Config: c}
	return notifier, nil
}

// Start starts all notifier treads
func (n *Notifier) Start() error {
	notificationChan := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-notificationChan:
			n.NotifyChannels()
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
		warningTime := time.Unix(channel.StandupTime-n.Config.ReminderTime*60, 0)
		if time.Now().Hour() == warningTime.Hour() && time.Now().Minute() == warningTime.Minute() {
			n.SendWarning(channel.ChannelID)
		}
		if time.Now().Hour() == standupTime.Hour() && time.Now().Minute() == standupTime.Minute() {
			go n.SendChannelNotification(channel.ChannelID)
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (n *Notifier) SendWarning(channelID string) {
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	if len(nonReporters) == 0 {
		return
	}

	nonReportersIDs := []string{}
	for _, user := range nonReporters {
		nonReportersIDs = append(nonReportersIDs, "<@"+user.UserID+">")
	}
	err = n.Chat.SendMessage(channelID, fmt.Sprintf(n.Config.Translate.NotifyUsersWarning, strings.Join(nonReportersIDs, ", "), n.Config.ReminderTime))
	if err != nil {
		logrus.Errorf("notifier: n.Chat.SendMessage failed: %v\n", err)
		return
	}

}

//SendChannelNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendChannelNotification(channelID string) {
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	// if everyone wrote their standups display all done message!
	if len(nonReporters) == 0 {
		err := n.Chat.SendMessage(channelID, n.Config.Translate.NotifyAllDone)
		if err != nil {
			logrus.Errorf("notifier: SendMessage failed: %v\n", err)
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
		err := n.Chat.SendUserMessage(nonReporter.UserID, fmt.Sprintf(n.Config.Translate.NotifyDirectMessage, nonReporter.UserID, channel.ChannelID, channel.ChannelName))
		if err != nil {
			logrus.Errorf("notifier: SendMessage failed: %v\n", err)
		}
	}

	repeats := 0

	notifyNotAll := func() error {
		nonReporters, err := n.getCurrentDayNonReporters(channelID)
		if err != nil {
			logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
			return err
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.UserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		if repeats < n.Config.ReminderRepeatsMax && len(nonReporters) > 0 {
			n.Chat.SendMessage(channelID, fmt.Sprintf(n.Config.Translate.NotifyNotAll, strings.Join(nonReportersSlackIDs, ", ")))
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		//n.notifyAdminsAboutNonReporters(channelID, nonReportersSlackIDs)
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(n.Config.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
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
