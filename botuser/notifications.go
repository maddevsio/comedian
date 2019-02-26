package botuser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
)

// NotifyChannels reminds users of channels about upcoming or missing standups
func (bot *Bot) NotifyChannels() {
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	channels, err := bot.db.GetTeamChannels(bot.Properties.TeamID)
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
		warningTime := time.Unix(channel.StandupTime-bot.Properties.ReminderTime*60, 0)
		if time.Now().Hour() == warningTime.Hour() && time.Now().Minute() == warningTime.Minute() {
			bot.SendWarning(channel.ChannelID)
		}

		if time.Now().Hour() == standupTime.Hour() && time.Now().Minute() == standupTime.Minute() {
			go bot.SendChannelNotification(channel.ChannelID)
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (bot *Bot) SendWarning(channelID string) {
	nonReporters, err := bot.getCurrentDayNonReporters(channelID)
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
	localizer := i18n.NewLocalizer(bot.bundle, bot.Properties.Language)
	minutes := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "Minutes",
			Description: "Translate minutes differently",
			One:         "{{.time}} minute",
			Other:       "{{.time}} minutes",
		},
		PluralCount: bot.Properties.ReminderTime,
		TemplateData: map[string]interface{}{
			"time": bot.Properties.ReminderTime,
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

	err = bot.SendMessage(channelID, warnNonReporters, nil)
	if err != nil {
		logrus.Errorf("notifier: bot.SendMessage failed: %v\n", err)
		return
	}
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (bot *Bot) SendChannelNotification(channelID string) {
	localizer := i18n.NewLocalizer(bot.bundle, bot.Properties.Language)

	members, err := bot.db.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("notifier: bot.db.ListChannelMembers failed: %v\n", err)
		return
	}
	if len(members) == 0 {
		logrus.Info("No standupers in this channel\n")
		return
	}
	nonReporters, err := bot.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	if len(nonReporters) == 0 {
		logrus.Info("len(nonReporters) == 0")
		return
	}

	channel, err := bot.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
	}

	repeats := 0

	notifyNotAll := func() error {
		nonReporters, err := bot.getCurrentDayNonReporters(channelID)
		if err != nil {
			logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
			return err
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.UserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		if repeats < bot.Properties.ReminderRepeatsMax && len(nonReporters) > 0 {

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

			bot.SendMessage(channelID, tagNonReporters, nil)
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
				TemplateData: map[string]interface{}{
					"user":        nonReporter.UserID,
					"channelID":   channel.ChannelID,
					"channelName": channel.ChannelName,
				},
			})

			err := bot.SendUserMessage(nonReporter.UserID, directMessage)
			if err != nil {
				logrus.Errorf("notifier: s.SendMessage failed: %v\n", err)
			}
		}
		//n.notifyAdminsAboutNonReporters(channelID, nonReportersSlackIDs)
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(bot.Properties.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

// getNonReporters returns a list of standupers that did not write standups
func (bot *Bot) getCurrentDayNonReporters(channelID string) ([]model.ChannelMember, error) {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	nonReporters, err := bot.db.GetNonReporters(channelID, timeFrom, time.Now())
	if err != nil && err != errors.New("no rows in result set") {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}
