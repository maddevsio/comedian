package botuser

import (
	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (bot *Bot) joinCommand(command slack.SlashCommand) string {
	_, err := bot.db.FindStansuperByUserID(command.UserID, command.ChannelID)
	if err == nil {
		youAlreadyStandup, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "youAlreadyStandup",
				Other: "You are already a part of standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return youAlreadyStandup
	}

	u, err := bot.slack.GetUserInfo(command.UserID)
	if err != nil {
		log.Error(err)
	}

	var realName string

	if u == nil {
		realName = "S.T.A.L.K.E.R."
	} else {
		realName = u.RealName
	}

	_, err = bot.db.CreateStanduper(model.Standuper{
		TeamID:                command.TeamID,
		UserID:                command.UserID,
		ChannelID:             command.ChannelID,
		ChannelName:           command.ChannelName,
		SubmittedStandupToday: true,
		RealName:              realName,
		RoleInChannel:         command.Text,
	})
	if err != nil {
		createStanduperFailed, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "createStanduperFailed",
				Other: "Could not add you to standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		log.Error("CreateStanduper failed: ", err)
		return createStanduperFailed
	}

	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		selectChannelFailed, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "selectChannelFailed",
				Other: "Could not add you to standup team, try kick and re-add me to the channel",
			},
		})
		if err != nil {
			log.Error(err)
		}
		log.Error("SelectChannel failed: ", err)
		return selectChannelFailed
	}

	if channel.StandupTime == "" {
		welcomeWithNoDeadline, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "welcomeNoDedline",
				Other: "Welcome to the standup team, no standup deadline has been setup yet",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return welcomeWithNoDeadline
	}

	welcomeWithDeadline, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "welcomeWithDedline",
			Other: "Welcome to the standup team, please, submit your standups no later than {{.Deadline}}",
		},
		TemplateData: map[string]interface{}{
			"Deadline": channel.StandupTime,
		},
	})
	if err != nil {
		log.Error(err)
	}
	return welcomeWithDeadline
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
		var role string
		role = member.RoleInChannel

		if member.RoleInChannel == "" {
			role = "developer"
		}
		list = append(list, member.RealName+" - "+role)
	}

	listStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "showStandupers",
			One:   "Only {{.Standupers}} standups in the team, /join to start standuping",
			Two:   "{{.Standupers}} submit standups in the team",
			Few:   "{{.Standupers}} submit standups in the team",
			Many:  "{{.Standupers}} submit standups in the team",
			Other: "{{.Standupers}} submit standups in the team",
		},
		PluralCount:  len(members),
		TemplateData: map[string]interface{}{"Standupers": strings.Join(list, ", ")},
	})
	if err != nil {
		log.Error(err)
	}

	return listStandupers
}

func (bot *Bot) quitCommand(command slack.SlashCommand) string {
	standuper, err := bot.db.FindStansuperByUserID(command.UserID, command.ChannelID)
	if err != nil {
		notStanduper, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "notStanduper",
				Other: "You do not standup yet",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return notStanduper
	}

	err = bot.db.DeleteStanduper(standuper.ID)
	if err != nil {
		log.Error("DeleteStanduper failed: ", err)
		failedLeaveStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "failedLeaveStandupers",
				Other: "Could not remove you from standup team",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return failedLeaveStandupers
	}

	leaveStanupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "leaveStanupers",
			Other: "You no longer have to submit standups, thanks for all your standups and messages",
		},
	})
	if err != nil {
		log.Error(err)
	}
	return leaveStanupers
}
