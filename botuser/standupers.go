package botuser

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (bot *Bot) joinCommand(command slack.SlashCommand) string {
	return ""
}

func (bot *Bot) showCommand(command slack.SlashCommand) string {
	members, err := bot.db.ListChannelStandupers(command.ChannelID)
	if err != nil || len(members) == 0 {
		listNoStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "listNoStandupers",
				Other: "",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return listNoStandupers
	}

	var list []string

	for _, member := range members {
		list = append(list, member.RealName+"-"+member.RoleInChannel)
	}

	listStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "listStandupers",
			Other: "",
		},
		PluralCount:  len(members),
		TemplateData: map[string]interface{}{"members": strings.Join(list, ", ")},
	})
	if err != nil {
		log.Error(err)
	}

	return listStandupers
}

func (bot *Bot) quitCommand(command slack.SlashCommand) string {
	return ""
}
