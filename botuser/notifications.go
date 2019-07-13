package botuser

import (
	"errors"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

//NotifierThread struct to manage notifier goroutines
type NotifierThread struct {
	channel model.Channel
	quit    chan struct{}
}

func (bot *Bot) warnChannels() {
	channels, err := bot.listTeamActiveChannels()
	if err != nil {
		log.Error(err)
		return
	}
	if len(channels) == 0 {
		return
	}

	for _, channel := range channels {
		bot.warnChannel(channel)
	}
}

func (bot *Bot) alarmChannels() {
	channels, err := bot.listTeamActiveChannels()
	if err != nil {
		return
	}

	for _, channel := range channels {
		bot.alarmChannel(channel)
	}
}

func (bot *Bot) warnChannel(channel model.Channel) string {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, _ := w.Parse(channel.StandupTime, time.Now())

	warningTime := time.Unix(r.Time.Unix()-bot.properties.ReminderTime*60, 0)

	if time.Now().Hour() != warningTime.Hour() || time.Now().Minute() != warningTime.Minute() {
		return "not warning time yet"
	}

	nonReporters, err := bot.findChannelNonReporters(channel)
	if err != nil {
		return "could not get non reporters"
	}

	message, err := bot.composeWarnMessage(nonReporters)
	if err != nil {
		return "could not warn non reporters"
	}

	bot.MessageChan <- Message{
		Type:    "message",
		Channel: channel.ChannelID,
		Text:    message,
	}

	return "message sent to channel"
}

func (bot *Bot) alarmChannel(channel model.Channel) string {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, _ := w.Parse(channel.StandupTime, time.Now())

	if time.Now().Hour() != r.Time.Hour() || time.Now().Minute() != r.Time.Minute() {
		return "not alarm time yet"
	}

	nt := &NotifierThread{channel: channel, quit: make(chan struct{})}

	bot.wg.Add(1)
	go func(nt *NotifierThread) {
		err := bot.sendAlarm(nt)
		if err != nil {
			log.Error(err)
		}
		bot.wg.Done()
	}(nt)
	bot.notifierThreads = append(bot.notifierThreads, nt)

	return "alarm begun"
}

func (bot *Bot) listTeamActiveChannels() ([]model.Channel, error) {
	var channels []model.Channel

	chs, err := bot.db.ListChannels()
	if err != nil {
		return channels, err
	}

	for _, channel := range chs {
		if channel.TeamID != bot.properties.TeamID {
			continue
		}

		if channel.StandupTime == "" {
			continue
		}

		w := when.New(nil)
		w.Add(en.All...)
		w.Add(ru.All...)

		r, err := w.Parse(channel.StandupTime, time.Now())
		if err != nil {
			continue
		}

		if r == nil {
			continue
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (bot *Bot) findChannelNonReporters(channel model.Channel) ([]string, error) {
	nonReporters := []string{}

	standupers, err := bot.db.ListChannelStandupers(channel.ChannelID)
	if err != nil {
		return nonReporters, err
	}
	for _, standuper := range standupers {
		if !bot.submittedStandupToday(standuper.UserID, standuper.ChannelID) {
			nonReporters = append(nonReporters, "<@"+standuper.UserID+">")
		}
	}

	return nonReporters, nil
}

func (bot *Bot) composeWarnMessage(nonReporters []string) (string, error) {
	if len(nonReporters) == 0 {
		return "", nil
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
		PluralCount:  int(bot.properties.ReminderTime),
		TemplateData: map[string]interface{}{"time": bot.properties.ReminderTime},
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

func (bot *Bot) sendAlarm(nt *NotifierThread) error {
	var repeats int

	alarm := func() error {
		select {
		case <-nt.quit:
			return nil
		default:
			err := bot.alarmRepeat(nt.channel, &repeats)
			if err != nil {
				return err
			}
			return nil
		}
	}

	b := backoff.NewConstantBackOff(time.Duration(bot.properties.NotifierInterval) * time.Minute)
	err := backoff.Retry(alarm, b)
	if err != nil {
		return errors.New("BackOff failed")
	}
	return nil
}

func (bot *Bot) alarmRepeat(channel model.Channel, repeats *int) error {

	if *repeats >= bot.properties.ReminderRepeatsMax {
		return nil
	}

	nonReporters, err := bot.findChannelNonReporters(channel)
	if err != nil {
		return nil
	}

	message, err := bot.composeAlarmMessage(nonReporters)

	bot.MessageChan <- Message{
		Type:    "message",
		Channel: channel.ChannelID,
		Text:    message,
	}

	*repeats++
	return errors.New("Continue backoff")
}
