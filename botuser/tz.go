package botuser

import (
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func (bot *Bot) modifyTZ(command slack.SlashCommand) string {
	tz := command.Text

	if strings.TrimSpace(tz) == "" {
		tz = "Asia/Bishkek"
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedRecognizeTZ",
				Other: "Failed to recognize new TZ you entered, double check the tz name and try again",
			},
		})

		if err != nil {
			log.Error(err)
		}
		return msg
	}

	channel, err := bot.db.SelectProject(command.ChannelID)
	if err != nil {
		failed, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "tzNotSet",
				Other: "Could not change channel time zone",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return failed
	}

	channel.TZ = tz

	_, err = bot.db.UpdateProject(channel)
	if err != nil {
		log.Error(err)
		msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedUpdateTZ",
				Other: "Failed to update Timezone",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return msg
	}

	msg, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "updateTZ",
			Other: "Channel timezone is updated, new TZ is {{.TZ}}",
		},
		TemplateData: map[string]interface{}{
			"TZ": tz,
		},
	})
	if err != nil {
		log.Error(err)
	}
	return msg
}
