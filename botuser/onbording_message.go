package botuser

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func (bot *Bot) modifyOnbordingMessage(command slack.SlashCommand) string {
	onbordingMessage := command.Text

	channel, err := bot.db.SelectProject(command.ChannelID)
	if err != nil {
		deadlineNotSet, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "onbordingMessageNotSet",
				Other: "Could not change channel onbording message",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return deadlineNotSet
	}

	channel.OnbordingMessage = onbordingMessage

	_, err = bot.db.UpdateProject(channel)
	if err != nil {
		log.Error(err)
		msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateOnbordingMessage",
				Other: "Failed to update onbording message",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return msg
	}

	msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateOnbordingMessage",
			Other: "Channel onbording message is updated, new message is {{.OM}}",
		},
		TemplateData: map[string]interface{}{
			"OM": onbordingMessage,
		},
	})
	if err != nil {
		log.Error(err)
	}
	return msg
}
