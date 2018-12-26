package api

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) addTime(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)
	if accessLevel > 3 {
		accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AccessAtLeastPM",
				Description: "Displays warning that role must be at least pm",
				Other:       "Access Denied! You need to be at least PM in this project to use this command!",
			},
		})
		return accessAtLeastPM
	}

	timeInt, err := utils.ParseTimeTextToInt(params)
	if err != nil {
		return err.Error()
	}
	err = ba.Bot.DB.CreateStandupTime(timeInt, channelID)
	if err != nil {
		somethingWentWrong := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "SomethingWentWrong",
				Description: "Displays message when occure unexpected errors",
				Other:       "Something went wrong. Please, try again later or report the problem to chatbot support!",
			},
		})
		logrus.Errorf("rest: CreateStandupTime failed: %v\n", err)
		return somethingWentWrong

	}
	channelMembers, err := ba.Bot.DB.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("BotAPI: ListChannelMembers failed: %v\n", err)
	}
	if len(channelMembers) == 0 {
		addStandupTimeNoUsers := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddStandupTimeNoUsers",
				Description: "Displays a message when a standup time is added for a channel, but there are no standupers in the channel.",
				Other:       "<!date^{{.timeInt}}^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>",
			},
			TemplateData: map[string]interface{}{
				"timeInt": timeInt,
			},
		})
		return addStandupTimeNoUsers

	}
	addStandupTime := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AddStandupTime",
			Description: "Displays a message when a standup time is successfully added for a channel",
			Other:       "<!date^{{.timeInt}}^Standup time set at {time}|Standup time set at 12:00>",
		},
		TemplateData: map[string]interface{}{
			"timeInt": timeInt,
		},
	})
	return addStandupTime

}

func (ba *BotAPI) removeTime(accessLevel int, channelID string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)
	if accessLevel > 3 {
		accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AccessAtLeastPM",
				Description: "Displays warning that role must be at least pm",
				Other:       "Access Denied! You need to be at least PM in this project to use this command!",
			},
		})
		return accessAtLeastPM

	}
	err := ba.Bot.DB.DeleteStandupTime(channelID)
	if err != nil {
		somethingWentWrong := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "SomethingWentWrong",
				Description: "Displays message when occure unexpected errors",
				Other:       "Something went wrong. Please, try again later or report the problem to chatbot support!",
			},
		})
		logrus.Errorf("rest: DeleteStandupTime failed: %v\n", err)
		return somethingWentWrong

	}
	st, err := ba.Bot.DB.ListChannelMembers(channelID)
	if len(st) != 0 {
		removeStandupTimeWithUsers := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "RemoveStandupTimeWithUsers",
				Description: "Displays message when a channels's standup time removed, but there are standupers in the channel",
				Other:       "standup time for this channel removed, but there are people marked as a standuper.",
			},
		})
		return removeStandupTimeWithUsers

	}
	removeStandupTime := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RemoveStandupTime",
			Description: "Displays message when a standuptime for a channel successfully removed",
			Other:       "standup time for channel deleted",
		},
	})
	return removeStandupTime

}

func (ba *BotAPI) showTime(channelID string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	standupTime, err := ba.Bot.DB.GetChannelStandupTime(channelID)
	if err != nil || standupTime == int64(0) {
		showNoStandupTime := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ShowNoStandupTime",
				Description: "Displays message when a standup time doesn't set for channel",
				Other:       "No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!",
			},
		})
		logrus.Errorf("GetChannelStandupTime failed: %v", err)
		return showNoStandupTime

	}
	showStandupTime := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ShowStandupTime",
			Description: "Shows a standup time of channel",
			Other:       "<!date^{{.standuptime}}^Standup time is {time}|Standup time set at 12:00>",
		},
		TemplateData: map[string]interface{}{
			"standuptime": standupTime,
		},
	})
	return showStandupTime

}
