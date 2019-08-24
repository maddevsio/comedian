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

	if command.Text == "" {
		return bot.removeDeadline(command)
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	wrongDeadlineFormat, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "wrongDeadlineFormat",
			Other: "Could not recognize deadline time. Use 1pm or 13:00 formats",
		},
	})
	if err != nil {
		log.Error(err)
	}

	r, err := w.Parse(command.Text, time.Now())
	if err != nil {
		return wrongDeadlineFormat
	}
	if r == nil {
		return wrongDeadlineFormat
	}

	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "deadlineNotSet",
				Other: "Could not change channel deadline",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	channel.StandupTime = r.Text

	_, err = bot.db.UpdateChannel(channel)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "deadlineNotSet",
				Other: "Could not change channel deadline",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	addStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "addStandupTime",
			Other: "Updated standup deadline to {{.Deadline}}",
		},
		TemplateData: map[string]interface{}{"Deadline": command.Text},
	})
	if err != nil {
		log.Error(err)
	}
	return addStandupTime
}

func (bot *Bot) removeDeadline(command slack.SlashCommand) string {
	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "deadlineNotSet",
				Other: "Could not change channel deadline",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	channel.StandupTime = ""

	_, err = bot.db.UpdateChannel(channel)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "deadlineNotSet",
				Other: "Could not change channel deadline",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	removeStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "removeStandupTime",
			Other: "Standup deadline removed",
		},
	})
	if err != nil {
		log.Error(err)
	}
	return removeStandupTime
}
