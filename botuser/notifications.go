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

//NotifierThread struct to manage notifier goroutines
type NotifierThread struct {
	channel model.Channel
	quit    chan struct{}
}

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
			nt := &NotifierThread{channel: channel, quit: make(chan struct{})}

			bot.wg.Add(1)
			go func(nt *NotifierThread) {
				err := bot.SendChannelNotification(nt)
				if err != nil {
					log.Error(err)
				}
				bot.wg.Done()
			}(nt)
			bot.AddNewNotifierThread(nt)
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
		if standuper.ChannelID == channelID && !bot.submittedStandupToday(standuper.UserID, standuper.ChannelID) && standuper.RoleInChannel != "pm" {
			nonReportersIDs = append(nonReportersIDs, "<@"+standuper.UserID+">")
		}
	}

	if len(nonReportersIDs) == 0 {
		log.Info("No non reporters to warn for channel: ", channelID)
		return nil
	}

	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "Minutes", int(bot.properties.ReminderTime), map[string]interface{}{"time": bot.properties.ReminderTime}}
	minutes := translation.Translate(payload)

	payload = translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "WarnNonReporters", len(nonReportersIDs), map[string]interface{}{"user": nonReportersIDs[0], "users": strings.Join(nonReportersIDs, ", "), "minutes": minutes}}
	warnNonReporters := translation.Translate(payload)

	err = bot.SendMessage(channelID, warnNonReporters, nil)
	if err != nil {
		log.Error(err)
		return nil
	}

	return nil
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (bot *Bot) SendChannelNotification(nt *NotifierThread) error {
	log.Info("SendChannelNotification for channel: ", nt.channel.ChannelID)
	standupers, err := bot.db.ListChannelStandupers(nt.channel.ChannelID)
	if err != nil {
		return err
	}

	if len(standupers) == 0 {
		log.Info("No standupers in channel: ", nt.channel.ChannelID)
		return nil
	}

	nonReporters := []model.Standuper{}
	for _, standuper := range standupers {
		if standuper.ChannelID == nt.channel.ChannelID && !bot.submittedStandupToday(standuper.UserID, standuper.ChannelID) && standuper.RoleInChannel != "pm" {
			nonReporters = append(nonReporters, standuper)
		}
	}

	if len(nonReporters) == 0 {
		log.Info("len(nonReporters) == 0")
		return nil
	}

	var repeats int

	notifyNotAll := func() error {
		select {
		case <-nt.quit:
			return nil
		default:
			err := bot.notifyNotAll(nt.channel.ChannelID, nonReporters, &repeats)
			if err != nil {
				return err
			}
			return nil
		}
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

	channel, err := bot.db.SelectChannel(channelID)
	if err != nil {
		log.Errorf("SelectChannel in notify not all failed for channel [%v]: [%v]", channelID, err)
		return err
	}

	if channel.StandupTime == 0 {
		log.Info("Channel standup time 0. Finish Backoff")
		return nil
	}

	if *repeats > bot.properties.ReminderRepeatsMax || len(nonReporters) < 1 {
		log.Info("Finish Backoff")
		return nil
	}

	roundNonReporters := []string{}
	for _, st := range nonReporters {
		if !bot.submittedStandupToday(st.UserID, st.ChannelID) && st.RoleInChannel != "pm" {
			roundNonReporters = append(roundNonReporters, fmt.Sprintf("<@%v>", st.UserID))
		}
	}

	if len(roundNonReporters) == 0 {
		log.Info("No non reporters in notifyNotAll")
		return nil
	}

	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "TagNonReporters", len(roundNonReporters), map[string]interface{}{"user": roundNonReporters[0], "users": strings.Join(roundNonReporters, ", ")}}
	tagNonReporters := translation.Translate(payload)

	err = bot.SendMessage(channelID, tagNonReporters, nil)
	if err != nil {
		log.Error("SendMessage in notify not all failed: ", err)
	}
	*repeats++
	return errors.New("Continue backoff")

}
