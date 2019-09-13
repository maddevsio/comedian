package botuser

import (
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

func (bot *Bot) notifyChannels() error {
	channels, err := bot.listTeamActiveChannels()
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		return nil
	}

	for _, channel := range channels {
		err := bot.notify(channel)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}

func (bot *Bot) notify(channel model.Project) error {
	if !shouldSubmitStandupIn(&channel, time.Now()) {
		return nil
	}

	loc, err := time.LoadLocation(channel.TZ)
	if err != nil {
		return err
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	//the error is ommited here since to get to this stage the channel
	//needs to have proper standup time
	r, _ := w.Parse(channel.Deadline, time.Now())

	alarmtime := time.Unix(r.Time.Unix(), 0)
	warningTime := time.Unix(r.Time.Unix()-bot.workspace.ReminderOffset*60, 0)

	var message string

	switch {
	case time.Now().In(loc).Hour() == warningTime.Hour() && time.Now().In(loc).Minute() == warningTime.Minute():
		nonReporters, err := bot.findChannelNonReporters(channel)
		if err != nil {
			return fmt.Errorf("could not get non reporters: %v", err)
		}

		message, err = bot.composeWarnMessage(nonReporters)
		if err != nil {
			return fmt.Errorf("could not compose Warn Message: %v", err)
		}

	case time.Now().In(loc).Hour() == alarmtime.Hour() && time.Now().In(loc).Minute() == alarmtime.Minute():
		threadTime := time.Now().Unix() + bot.conf.NotificationTime*60

		nonReporters, err := bot.findChannelNonReporters(channel)

		if err != nil {
			return fmt.Errorf("could not get non reporters: %v", err)
		}

		if len(nonReporters) > 0 {
			usersNonReport := strings.Join(nonReporters, ",")

			_, err = bot.db.CreateNotificationThread(model.NotificationThread{
				ChannelID:        channel.ChannelID,
				UserIDs:          usersNonReport,
				NotificationTime: threadTime,
				ReminderCounter:  0,
			})
			if err != nil {
				log.Error("Error on executing CreateNotificationThread ", err, "ChannelID: ", channel.ChannelID)
				return err
			}
		}
		message, err = bot.composeAlarmMessage(nonReporters)
		if err != nil {
			return fmt.Errorf("could not compose Alarm Message: %v", err)
		}
	}

	bot.send(&Message{
		Type:    "message",
		Channel: channel.ChannelID,
		Text:    message,
	})

	thread, err := bot.db.SelectNotificationsThread(channel.ChannelID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Error("Error on executing SelectNotificationsThread! ", err, "ChannelID: ", channel.ChannelID, "ChannelName: ", channel.ChannelName)
		return err
	}

	var remindTime time.Time

	remindTime = time.Unix(thread.NotificationTime, 0)

	if time.Now().In(loc).Hour() != remindTime.Hour() || time.Now().In(loc).Minute() != remindTime.Minute() {
		return nil
	}

	if thread.ReminderCounter >= bot.workspace.MaxReminders {
		err = bot.db.DeleteNotificationThread(thread.ID)
		if err != nil {
			log.Error("Error on executing DeleteNotificationsThread! ", err, "Thread ID: ", thread.ID)
			return err
		}
	}

	stillNonReporters := strings.Split(thread.UserIDs, ",")

	var updatedNonReporters string

	for i, nonReport := range stillNonReporters {
		if bot.submittedStandupToday(nonReport, thread.ChannelID) {
			stillNonReporters = append(stillNonReporters[:i], stillNonReporters[i+1:]...)
		}
	}
	updatedNonReporters = strings.Join(stillNonReporters, ",")

	if len(updatedNonReporters) == 0 {
		err = bot.db.DeleteNotificationThread(thread.ID)
		if err != nil {
			log.Error("Error on executing DeleteNotificationsThread! ", err, "Thread ID: ", thread.ID)
			return err
		}
	}

	message, err = bot.composeRemindMessage(stillNonReporters)
	if err != nil {
		return fmt.Errorf("could not compose Remind Message: %v", err)
	}

	if message == "" {
		return nil
	}

	bot.send(&Message{
		Type:    "message",
		Channel: channel.ChannelID,
		Text:    message,
	})

	thread.NotificationTime = thread.NotificationTime + bot.conf.NotificationTime*60

	return bot.db.UpdateNotificationThread(thread.ID, thread.ChannelID, thread.NotificationTime, updatedNonReporters)
}

func (bot *Bot) listTeamActiveChannels() ([]model.Project, error) {
	var channels []model.Project

	chs, err := bot.db.ListWorkspaceProjects(bot.workspace.WorkspaceID)
	if err != nil {
		return channels, err
	}

	for _, channel := range chs {
		if channel.Deadline == "" {
			continue
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (bot *Bot) findChannelNonReporters(project model.Project) ([]string, error) {
	nonReporters := []string{}

	standupers, err := bot.db.ListProjectStandupers(project.ChannelID)
	if err != nil {
		return nonReporters, err
	}
	for _, standuper := range standupers {
		if !bot.submittedStandupToday(standuper.UserID, standuper.ChannelID) {
			nonReporters = append(nonReporters, standuper.UserID)
		}
	}

	return nonReporters, nil
}

func (bot *Bot) composeWarnMessage(nonReporters []string) (string, error) {
	if len(nonReporters) == 0 {
		return "", nil
	}

	for i, nr := range nonReporters {
		nonReporters[i] = "<@" + nr + ">"
	}

	minutes, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "minutes",
			One:   "{{.time}} minute",
			Two:   "{{.time}} minutes",
			Few:   "{{.time}} minutes",
			Many:  "{{.time}} minutes",
			Other: "{{.time}} minutes",
		},
		PluralCount:  int(bot.workspace.ReminderOffset),
		TemplateData: map[string]interface{}{"time": bot.workspace.ReminderOffset},
	})
	if err != nil {
		return "", err
	}

	warnNonReporters, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "warnNonReporters",
			One:   "{{.user}}, you are the only one to miss standup, in {{.minutes}}, hurry up!",
			Two:   "{{.users}} you may miss the deadline in {{.minutes}}",
			Few:   "{{.users}} you may miss the deadline in {{.minutes}}",
			Many:  "{{.users}} you may miss the deadline in {{.minutes}}",
			Other: "{{.users}} you may miss the deadline in {{.minutes}}",
		},
		PluralCount:  len(nonReporters),
		TemplateData: map[string]interface{}{"user": nonReporters[0], "users": strings.Join(nonReporters, ", "), "minutes": minutes},
	})
	if err != nil {
		return "", err
	}

	return warnNonReporters, nil
}

func (bot *Bot) composeAlarmMessage(nonReporters []string) (string, error) {
	if len(nonReporters) == 0 {
		return "", nil
	}

	for i, nr := range nonReporters {
		nonReporters[i] = "<@" + nr + ">"
	}

	alarmNonReporters, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "tagNonReporters",
			One:   "{{.user}}, you are the only one missed standup, shame!",
			Two:   "{{.users}} you have missed standup deadlines, shame!",
			Few:   "{{.users}} you have missed standup deadlines, shame!",
			Many:  "{{.users}} you have missed standup deadlines, shame!",
			Other: "{{.users}} you have missed standup deadlines, shame!",
		},
		PluralCount:  len(nonReporters),
		TemplateData: map[string]interface{}{"user": nonReporters[0], "users": strings.Join(nonReporters, ", ")},
	})
	if err != nil {
		return "", err
	}

	return alarmNonReporters, nil
}

func (bot *Bot) composeRemindMessage(nonReporters []string) (string, error) {
	if len(nonReporters) == 0 {
		return "", nil
	}

	for i, nr := range nonReporters {
		nonReporters[i] = "<@" + nr + ">"
	}

	remindNonReporters, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "tagStillNonReporters",
			One:   "{{.user}}, you still haven't written a standup! Write a standup!",
			Two:   "{{.users}} you still haven't written a standup! Write a standup!",
			Few:   "{{.users}} you still haven't written a standup! Write a standup!",
			Many:  "{{.users}} you still haven't written a standup! Write a standup!",
			Other: "{{.users}} you still haven't written a standup! Write a standup!",
		},
		PluralCount:  len(nonReporters),
		TemplateData: map[string]interface{}{"user": nonReporters[0], "users": strings.Join(nonReporters, ",")},
	})
	if err != nil {
		return "", err
	}

	return remindNonReporters, nil
}

func shouldSubmitStandupIn(channel *model.Project, t time.Time) bool {
	// TODO need to think of how to include translated versions
	if strings.Contains(channel.SubmissionDays, strings.ToLower(t.Weekday().String())) {
		return true
	}
	return false
}
