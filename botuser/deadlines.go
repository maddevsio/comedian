package botuser

import (
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

func (bot *Bot) addDeadline(command slack.SlashCommand) string {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, err := w.Parse(command.Text, time.Now())
	if err != nil {
		log.Error("Failed to parse params", err)
		return "Unable to recognize time for a deadline"
	}
	if r == nil {
		log.Error("r is nil. No matches found", err)
		return "Unable to recognize time for a deadline"
	}

	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		log.Error("failed to select channel", err)
		return "could not recognize channel, please add me to the channel and try again"
	}

	channel.StandupTime = r.Text

	if nt, exist := bot.FindNotifierThread(channel); exist {
		go bot.StopNotifierThread(nt)
		bot.DeleteNotifierThreadFromList(channel)
	}

	_, err = bot.db.UpdateChannel(channel)
	if err != nil {
		log.Error("failed to update channel", err)
		return "could not set channel deadline"
	}

	standupers, err := bot.db.ListChannelStandupers(command.ChannelID)
	if err != nil {
		log.Errorf("BotAPI: ListChannelStandupers failed: %v\n", err)
	}

	if len(standupers) == 0 {
		addStandupTimeNoUsers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "addStandupTimeNoUsers",
				Other: "",
			},
			TemplateData: map[string]interface{}{"timeInt": r.Time.Unix()},
		})
		if err != nil {
			log.Error(err)
		}
		return addStandupTimeNoUsers
	}

	addStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "addStandupTime",
			Other: "",
		},
		TemplateData: map[string]interface{}{"timeInt": r.Time.Unix()},
	})
	if err != nil {
		log.Error(err)
	}
	return addStandupTime
}

func (bot *Bot) removeDeadline(command slack.SlashCommand) string {
	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		return "could not recognize channel, please add me to the channel and try again"
	}

	channel.StandupTime = ""

	if nt, exist := bot.FindNotifierThread(channel); exist {
		go bot.StopNotifierThread(nt)
		bot.DeleteNotifierThreadFromList(channel)
	}

	_, err = bot.db.UpdateChannel(channel)

	if err != nil {
		return "could not remove channel deadline"
	}
	st, err := bot.db.ListChannelStandupers(command.ChannelID)
	if len(st) != 0 {
		removeStandupTimeWithUsers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "removeStandupTimeWithUsers",
				Other: "",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return removeStandupTimeWithUsers
	}

	removeStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "removeStandupTime",
			Other: "",
		},
	})
	if err != nil {
		log.Error(err)
	}
	return removeStandupTime
}

func (bot *Bot) showDeadline(command slack.SlashCommand) string {
	channel, err := bot.db.SelectChannel(command.ChannelID)
	// need to check error first, because it is misleading!
	if err != nil || channel.StandupTime == "" {
		showNoStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "showNoStandupTime",
				Other: "",
			},
			TemplateData: map[string]interface{}{"standuptime": channel.StandupTime},
		})
		if err != nil {
			log.Error(err)
		}
		return showNoStandupTime
	}

	showStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "showStandupTime",
			Other: "",
		},
		TemplateData: map[string]interface{}{"standuptime": channel.StandupTime},
	})
	if err != nil {
		log.Error(err)
	}
	return showStandupTime
}
