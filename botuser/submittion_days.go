package botuser

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func (bot *Bot) modifySubmittionDays(command slack.SlashCommand) string {
	submittionDays := command.Text

	channel, err := bot.db.SelectProject(command.ChannelID)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "submittionDaysNotSet",
				Other: "Could not change channel submittion days",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	channel.SubmissionDays = submittionDays

	_, err = bot.db.UpdateProject(channel)
	if err != nil {
		log.Error(err)
		msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateSumittionDays",
				Other: "Failed to update Sumittion Days",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return msg
	}

	msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateSubmittionDays",
			Other: "Channel submittion days are updated, new schedule is {{.SD}}",
		},
		TemplateData: map[string]interface{}{
			"SD": submittionDays,
		},
	})
	if err != nil {
		log.Error(err)
	}
	return msg
}
