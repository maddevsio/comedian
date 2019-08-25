package botuser

import (
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
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
		u = &slack.User{RealName: command.UserName}
	}

	ch, err := bot.slack.GetChannelInfo(command.ChannelID)
	if err != nil {
		ch = &slack.Channel{}
		ch.Name = command.ChannelName
	}

	_, err = bot.db.CreateStanduper(model.Standuper{
		Created:       time.Now().UTC(),
		TeamID:        command.TeamID,
		UserID:        command.UserID,
		ChannelID:     command.ChannelID,
		ChannelName:   ch.Name,
		RealName:      u.RealName,
		RoleInChannel: command.Text,
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
		channel, err = bot.db.CreateChannel(model.Channel{
			TeamID:           command.TeamID,
			ChannelID:        command.ChannelID,
			ChannelName:      ch.Name,
			StandupTime:      "",
			TZ:               "Asia/Bishkek",
			OnbordingMessage: "Hello and welcome to " + ch.Name,
			SubmissionDays:   "monday, tuesday, wednesday, thirsday, friday",
		})
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
	var deadline string
	channel, err := bot.db.SelectChannel(command.ChannelID)
	if err != nil {
		ch, err := bot.slack.GetChannelInfo(command.ChannelID)
		if err != nil {
			log.Error("Failed to GetChannelInfo in show command: ", err)
			ch = &slack.Channel{}
			ch.Name = command.ChannelName
		}

		channel, err = bot.db.CreateChannel(model.Channel{
			TeamID:           command.TeamID,
			ChannelID:        command.ChannelID,
			ChannelName:      ch.Name,
			StandupTime:      "",
			TZ:               "Asia/Bishkek",
			OnbordingMessage: "Hello and welcome to " + ch.Name,
			SubmissionDays:   "monday, tuesday, wednesday, thirsday, friday",
		})
		if err != nil {
			log.Error("Failed to create channel in show command: ", err)
		}
	}

	if channel.StandupTime == "" {
		showNoStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "showNoStandupTime",
				Other: "Standup deadline is not set",
			},
		})
		if err != nil {
			log.Error(err)
		}
		deadline = showNoStandupTime
	} else {
		showStandupTime, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "showStandupTime",
				Other: "Standup deadline is {{.Deadline}}",
			},
			TemplateData: map[string]interface{}{"Deadline": channel.StandupTime},
		})
		if err != nil {
			log.Error(err)
		}
		deadline = showStandupTime
	}

	members, err := bot.db.ListChannelStandupers(command.ChannelID)
	if err != nil || len(members) == 0 {
		listNoStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "listNoStandupers",
				Other: "No standupers in the team, /start to start standuping. ",
			},
		})
		if err != nil {
			log.Error(err)
		}
		return listNoStandupers + "\n" + deadline
	}

	var list []string

	for _, member := range members {
		var role string
		role = member.RoleInChannel

		if member.RoleInChannel == "" {
			role = "developer"
		}
		list = append(list, fmt.Sprintf("%s(%s)", member.RealName, role))
	}

	listStandupers, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "showStandupers",
			One:   "Only {{.Standupers}} submits standups in the team, '/start' to begin. ",
			Two:   "{{.Standupers}} submit standups in the team. ",
			Few:   "{{.Standupers}} submit standups in the team. ",
			Many:  "{{.Standupers}} submit standups in the team. ",
			Other: "{{.Standupers}} submit standups in the team. ",
		},
		PluralCount:  len(members),
		TemplateData: map[string]interface{}{"Standupers": strings.Join(list, ", ")},
	})
	if err != nil {
		log.Error(err)
	}

	return listStandupers + "\n" + deadline
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
