package botuser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/translation"
)

// NotifyChannels reminds users of channels about upcoming or missing standups
func (bot *Bot) NotifyChannels(t time.Time) {
	if int(t.Weekday()) == 6 || int(t.Weekday()) == 0 {
		return
	}
	channels, err := bot.db.GetTeamChannels(bot.properties.TeamID)
	if err != nil {
		log.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
		return
	}

	// For each standup time, if standup time is now, start reminder
	for _, channel := range channels {
		if channel.StandupTime == 0 {
			continue
		}
		standupTime := time.Unix(channel.StandupTime, 0)
		warningTime := time.Unix(channel.StandupTime-bot.properties.ReminderTime*60, 0)
		if t.Hour() == warningTime.Hour() && t.Minute() == warningTime.Minute() {
			err := bot.SendWarning(channel.ChannelID)
			if err != nil {
				log.Error(err)
			}
		}

		if t.Hour() == standupTime.Hour() && t.Minute() == standupTime.Minute() {
			go bot.SendChannelNotification(channel.ChannelID)
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (bot *Bot) SendWarning(channelID string) error {
	nonReporters, err := bot.getCurrentDayNonReporters(channelID)
	if err != nil {
		return err
	}

	if len(nonReporters) == 0 {
		return nil
	}

	nonReportersIDs := []string{}
	for _, user := range nonReporters {
		nonReportersIDs = append(nonReportersIDs, "<@"+user.UserID+">")
	}

	payload := translation.Payload{bot.bundle, bot.properties.Language, "Minutes", int(bot.properties.ReminderTime), map[string]interface{}{"time": bot.properties.ReminderTime}}
	minutes, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}

	payload = translation.Payload{bot.bundle, bot.properties.Language, "WarnNonReporters", len(nonReporters), map[string]interface{}{"user": nonReportersIDs[0], "users": strings.Join(nonReportersIDs, ", "), "minutes": minutes}}
	warnNonReporters, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}

	err = bot.SendMessage(channelID, warnNonReporters, nil)
	if err != nil {
		return err
	}

	return nil
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (bot *Bot) SendChannelNotification(channelID string) {
	members, err := bot.db.ListChannelMembers(channelID)
	if err != nil {
		log.Errorf("notifier: bot.db.ListChannelMembers failed: %v\n", err)
		return
	}
	if len(members) == 0 {
		log.Info("No standupers in this channel\n")
		return
	}
	nonReporters, err := bot.getCurrentDayNonReporters(channelID)
	if err != nil {
		log.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	if len(nonReporters) == 0 {
		log.Info("len(nonReporters) == 0")
		return
	}

	channel, err := bot.db.SelectChannel(channelID)
	if err != nil {
		log.Errorf("notifier: SelectChannel failed: %v\n", err)
		return
	}

	var repeats int

	notifyNotAll := func() error {
		err := bot.notifyNotAll(channel, &repeats)
		if err != nil {
			return err
		}
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(bot.properties.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		log.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

func (bot *Bot) notifyNotAll(channel model.Channel, repeats *int) error {

	nonReporters, err := bot.getCurrentDayNonReporters(channel.ChannelID)
	if err != nil {
		log.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return err
	}

	nonReportersSlackIDs := []string{}
	for _, nonReporter := range nonReporters {
		nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.UserID))
	}
	log.Infof("notifier: Notifier non reporters: %v", nonReporters)

	if *repeats < bot.properties.ReminderRepeatsMax && len(nonReporters) > 0 {

		payload := translation.Payload{bot.bundle, bot.properties.Language, "TagNonReporters", len(nonReporters), map[string]interface{}{"user": nonReportersSlackIDs[0], "users": strings.Join(nonReportersSlackIDs, ", ")}}
		tagNonReporters, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}

		bot.SendMessage(channel.ChannelID, tagNonReporters, nil)
		*repeats++
		err = errors.New("Continue backoff")
		return err
	}
	// othervise Direct Message non reporters
	for _, nonReporter := range nonReporters {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "DirectMessage", len(nonReporters), map[string]interface{}{"user": nonReporter.UserID, "channelID": channel.ChannelID, "channelName": channel.ChannelName}}
		directMessage, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}

		err = bot.SendUserMessage(nonReporter.UserID, directMessage)
		if err != nil {
			log.Errorf("notifier: s.SendMessage failed: %v\n", err)
		}
	}
	return nil
}

// getNonReporters returns a list of standupers that did not write standups
func (bot *Bot) getCurrentDayNonReporters(channelID string) ([]model.ChannelMember, error) {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	nonReporters, err := bot.db.GetNonReporters(channelID, timeFrom, time.Now())
	if err != nil && err != errors.New("no rows in result set") {
		log.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}
