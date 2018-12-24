package notifier

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gitlab.com/team-monitoring/comedian/model"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
)

// Notifier struct is used to notify users about upcoming or skipped standups
type Notifier struct {
	bot *bot.Bot
}

// NewNotifier creates a new notifier
func NewNotifier(bot *bot.Bot) (Notifier, error) {
	notifier := Notifier{bot: bot}
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
	channels, err := n.bot.DB.GetChannels()
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
		warningTime := time.Unix(channel.StandupTime-n.bot.CP.ReminderTime*60, 0)
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
	tts, err := n.bot.DB.ListTimeTablesForDay(day)
	if err != nil {
		logrus.Errorf("ListTimeTablesForToday failed: %v", err)
		return
	}

	for _, tt := range tts {
		standupTime := time.Unix(tt.ShowDeadlineOn(day), 0)
		warningTime := time.Unix(tt.ShowDeadlineOn(day)-n.bot.CP.ReminderTime*60, 0)

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
		if !n.bot.DB.MemberHasTimeTable(u.ID) {
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
	localizer := i18n.NewLocalizer(n.bot.Bundle, n.bot.CP.Language)
	minutes := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "Minutes",
			Description: "Translate minutes differently",
			One:         "{{.time}} minute",
			Other:       "{{.time}} minutes",
		},
		PluralCount: n.bot.CP.ReminderTime,
		TemplateData: map[string]interface{}{
			"time": n.bot.CP.ReminderTime,
		},
	})

	warnNonReporters := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "WarnNonReporters",
			Description: "Warning message to those who did not submit standup",
			One:         "Hey, {{.user}}! {{.minutes}} to deadline and you are the only one who still did not submit standup! Brace yourselve!",
			Other:       "Hey, {{.users}}! {{.minutes}} to deadline and you people still did not submit standups! Go ahead!",
		},
		PluralCount: len(nonReporters),
		TemplateData: map[string]interface{}{
			"user":    nonReportersIDs[0],
			"users":   strings.Join(nonReportersIDs, ", "),
			"minutes": minutes,
		},
	})

	err = n.bot.SendMessage(channelID, warnNonReporters, nil)
	if err != nil {
		logrus.Errorf("notifier: n.bot.SendMessage failed: %v\n", err)
		return
	}
}

// SendIndividualWarning reminds users in chat about upcoming standups
func (n *Notifier) SendIndividualWarning(channelMemberID int64) {
	chm, err := n.bot.DB.SelectChannelMember(channelMemberID)
	if err != nil {
		logrus.Errorf("SelectChannelMember failed: %v", err)
		return
	}
	submittedStandup := n.bot.DB.SubmittedStandupToday(chm.UserID, chm.ChannelID)
	if !submittedStandup {
		localizer := i18n.NewLocalizer(n.bot.Bundle, n.bot.CP.Language)

		minutes := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "Minutes",
				Description: "Translate minutes differently",
				One:         "{{.time}} minute",
				Other:       "{{.time}} minutes",
			},
			PluralCount: n.bot.CP.ReminderTime,
			TemplateData: map[string]interface{}{
				"time": n.bot.CP.ReminderTime,
			},
		})

		warnIndividualNonReporters := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "WarnIndividualNonReporters",
				Description: "Warning message to those who did not submit standup",
				Other:       "Hey, {{.users}}! {{.minutes}} to deadline and you still did not submit standup! Hurry up!",
			},
			TemplateData: map[string]interface{}{
				"user":    chm.UserID,
				"minutes": minutes,
			},
		})

		err = n.bot.SendMessage(chm.ChannelID, warnIndividualNonReporters, nil)
		if err != nil {
			logrus.Errorf("notifier: n.bot.SendMessage failed: %v\n", err)
			return
		}
		return
	}
	logrus.Infof("%v is not non reporter", chm.UserID)
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendChannelNotification(channelID string) {
	localizer := i18n.NewLocalizer(n.bot.Bundle, n.bot.CP.Language)

	members, err := n.bot.DB.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.bot.DB.ListChannelMembers failed: %v\n", err)
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
		if !n.bot.DB.MemberHasTimeTable(u.ID) {
			nonReporters = append(nonReporters, u)
		}
	}
	if len(nonReporters) == 0 {
		return
	}

	channel, err := n.bot.DB.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
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
			if !n.bot.DB.MemberHasTimeTable(u.ID) {
				nonReporters = append(nonReporters, u)
			}
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.UserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		if repeats < n.bot.CP.ReminderRepeatsMax && len(nonReporters) > 0 {

			tagNonReporters := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TagNonReporters",
					Description: "Display message about those who did not submit standup",
					One:         "Hey, {{.user}}! You missed deadline and you are the only one who still did not submit standup! Get it done!",
					Other:       "Hey, {{.users}}! You all missed deadline and still did not submit standups! Time management problems detected!",
				},
				PluralCount: len(nonReporters),
				TemplateData: map[string]interface{}{
					"user":  nonReportersSlackIDs[0],
					"users": strings.Join(nonReportersSlackIDs, ", "),
				},
			})

			n.bot.SendMessage(channelID, tagNonReporters, nil)
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		// othervise Direct Message non reporters
		for _, nonReporter := range nonReporters {
			directMessage := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "DirectMessage",
					Description: "DM Warning message to those who did not submit standup",
					Other:       "Hey, <@{{.user}}>! you failed to submit standup in <#{{.channelID}}|{{.channelName}}> on time! Do it ASAP!",
				},
				PluralCount: len(nonReporters),
				TemplateData: map[string]interface{}{
					"user":        nonReporter.UserID,
					"channelID":   channel.ChannelID,
					"channelName": channel.ChannelName,
				},
			})

			err := n.bot.SendUserMessage(nonReporter.UserID, directMessage)
			if err != nil {
				logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
			}
		}
		//n.notifyAdminsAboutNonReporters(channelID, nonReportersSlackIDs)
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(n.bot.CP.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

//SendIndividualNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendIndividualNotification(channelMemberID int64) {
	localizer := i18n.NewLocalizer(n.bot.Bundle, n.bot.CP.Language)

	chm, err := n.bot.DB.SelectChannelMember(channelMemberID)
	if err != nil {
		logrus.Errorf("SelectChannelMember failed: %v", err)
		return
	}
	channel, err := n.bot.DB.SelectChannel(chm.ChannelID)
	if err != nil {
		logrus.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
	}
	submittedStandup := n.bot.DB.SubmittedStandupToday(chm.UserID, chm.ChannelID)
	if submittedStandup {
		return
	}
	repeats := 0
	notify := func() error {
		submittedStandup := n.bot.DB.SubmittedStandupToday(chm.UserID, chm.ChannelID)
		if repeats < n.bot.CP.ReminderRepeatsMax && !submittedStandup {
			individualStandupersLate := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "IndividualStandupersLate",
					Description: "Display message about those who missed submit standup",
					Other:       "{{.user}}, you missed standup deadline! Submit standup ASAP!",
				},
				TemplateData: map[string]interface{}{
					"user": chm.UserID,
				},
			})

			n.bot.SendMessage(channel.ChannelID, individualStandupersLate, nil)
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		if !submittedStandup {
			tagIndividualNonReporters := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TagIndividualNonReporters",
					Description: "Display message about those who did not submit standup with individual schedules",
					Other:       "Hey, {{.user}}! You failed to submit standup in time! Get it done ASAP!",
				},
				TemplateData: map[string]interface{}{
					"user": chm.UserID,
				},
			})

			err := n.bot.SendUserMessage(chm.UserID, tagIndividualNonReporters)
			if err != nil {
				logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
			}
		}
		logrus.Infof("User %v submitted standup!", chm.UserID)
		return nil
	}
	b := backoff.NewConstantBackOff(time.Duration(n.bot.CP.NotifierInterval) * time.Minute)
	err = backoff.Retry(notify, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

// getNonReporters returns a list of standupers that did not write standups
func (n *Notifier) getCurrentDayNonReporters(channelID string) ([]model.ChannelMember, error) {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	nonReporters, err := n.bot.DB.GetNonReporters(channelID, timeFrom, time.Now())
	if err != nil && err != errors.New("no rows in result set") {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}
