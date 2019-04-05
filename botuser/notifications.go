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
	channels, err := bot.db.ListChannels()
	if err != nil {
		log.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
		return
	}

	// For each standup time, if standup time is now, start reminder
	for _, channel := range channels {
		if channel.TeamID != bot.properties.TeamID {
			continue
		}

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
			bot.wg.Add(1)
			go func() {
				err := bot.SendChannelNotification(channel.ChannelID)
				if err != nil {
					log.Error(err)
				}
				bot.wg.Done()
			}()
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (bot *Bot) SendWarning(channelID string) error {
	log.Info("SendWarning for channel: ", channelID)
	standupers, err := bot.db.ListStandupers()
	if err != nil {
		return err
	}

	if len(standupers) == 0 {
		log.Info("No standupers to warn for channel: ", channelID)
		return nil
	}

	nonReportersIDs := []string{}
	for _, standuper := range standupers {
		if standuper.ChannelID == channelID && !standuper.SubmittedStandupToday {
			nonReportersIDs = append(nonReportersIDs, "<@"+standuper.UserID+">")
		}
	}

	if len(nonReportersIDs) == 0 {
		log.Info("No non reporters to warn for channel: ", channelID)
		return nil
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

	payload = translation.Payload{bot.bundle, bot.properties.Language, "WarnNonReporters", len(nonReportersIDs), map[string]interface{}{"user": nonReportersIDs[0], "users": strings.Join(nonReportersIDs, ", "), "minutes": minutes}}
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
		log.Error(err)
		return errors.New("Could not post message to a channel")
	}

	return nil
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (bot *Bot) SendChannelNotification(channelID string) error {
	log.Info("SendChannelNotification for channel: ", channelID)
	standupers, err := bot.db.ListChannelStandupers(channelID)
	if err != nil {
		return err
	}

	if len(standupers) == 0 {
		log.Info("No standupers in channel: ", channelID)
		return nil
	}

	nonReporters := []model.Standuper{}
	for _, standuper := range standupers {
		if standuper.ChannelID == channelID && !standuper.SubmittedStandupToday {
			nonReporters = append(nonReporters, standuper)
		}
	}

	if len(nonReporters) == 0 {
		log.Info("len(nonReporters) == 0")
		return nil
	}

	var repeats int

	notifyNotAll := func() error {
		err := bot.notifyNotAll(channelID, nonReporters, &repeats)
		if err != nil {
			return err
		}
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(bot.properties.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		log.Errorf("notifier: backoff.Retry failed: %v\n", err)
		return errors.New("BackOff failed")
	}
	return nil
}

func (bot *Bot) notifyNotAll(channelID string, nonReporters []model.Standuper, repeats *int) error {

	if *repeats > bot.properties.ReminderRepeatsMax || len(nonReporters) < 1 {
		log.Info("Finish Backoff")
		return nil
	}

	roundNonReporters := []string{}
	for _, st := range nonReporters {
		standuper, err := bot.db.GetStanduper(st.ID)
		if err != nil {
			log.Error("notifyNotAll failed at GetStanduper: ", err)
			continue
		}
		if !standuper.SubmittedStandupToday {
			roundNonReporters = append(roundNonReporters, fmt.Sprintf("<@%v>", standuper.UserID))
		}
	}

	if len(roundNonReporters) == 0 {
		log.Info("No non reporters in notifyNotAll")
		return nil
	}

	payload := translation.Payload{bot.bundle, bot.properties.Language, "TagNonReporters", len(roundNonReporters), map[string]interface{}{"user": roundNonReporters[0], "users": strings.Join(roundNonReporters, ", ")}}
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

	err = bot.SendMessage(channelID, tagNonReporters, nil)
	if err != nil {
		log.Error("SendMessage in notify not all failed: ", err)
	}
	*repeats++
	err = errors.New("Continue backoff")
	return err

}
