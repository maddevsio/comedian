package botuser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

func (bot *Bot) addTime(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
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

	timeInt, err := bot.ParseTimeTextToInt(params)
	if err != nil {
		return err.Error()
	}
	err = bot.DB.CreateStandupTime(timeInt, channelID)
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
	channelMembers, err := bot.DB.ListChannelMembers(channelID)
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

func (bot *Bot) removeTime(accessLevel int, channelID string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
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
	err := bot.DB.DeleteStandupTime(channelID)
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
	st, err := bot.DB.ListChannelMembers(channelID)
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

func (bot *Bot) showTime(channelID string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	standupTime, err := bot.DB.GetChannelStandupTime(channelID)
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

//ParseTimeTextToInt parse time text to int
func (bot *Bot) ParseTimeTextToInt(timeText string) (int64, error) {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	if timeText == "0" {
		return int64(0), nil
	}
	matchHourMinuteFormat, _ := regexp.MatchString("[0-9][0-9]:[0-9][0-9]", timeText)
	matchAMPMFormat, _ := regexp.MatchString("[0-9][0-9][a-z]", timeText)

	if matchHourMinuteFormat {
		t := strings.Split(timeText, ":")
		hours, _ := strconv.Atoi(t[0])
		munites, _ := strconv.Atoi(t[1])

		if hours > 23 || munites > 59 {
			wrongTimeFormat := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "WrongTimeFormat",
					Description: "Displays message if a time format is wrong",
					Other:       "Wrong time! Please, check the time format and try again!",
				},
			})
			return int64(0), errors.New(wrongTimeFormat)
		}
		return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hours, munites, 0, 0, time.Local).Unix(), nil
	}

	if matchAMPMFormat {
		shortTimeFormat := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "shortTimeFormat",
				Description: "Displays message if a time format is short",
				Other:       "Seems like you used short time format, please, use 24:00 hour format instead!",
			},
		})
		return int64(0), errors.New(shortTimeFormat)
	}

	unknownTimeFormat := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "unknownTimeFormat",
			Description: "Displays message if time format is unknown",
			Other:       "Could not understand how you mention time. Please, use 24:00 hour format and try again!",
		},
	})
	return int64(0), errors.New(unknownTimeFormat)

}
