package api

import (
	"regexp"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) addTimeTable(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "AccessAtLeastPM",
		})
		return accessAtLeastPM
	}

	daysDivider := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "DaysDivider",
			Description: "Days divider",
			Other:       " on ",
		},
	})
	timeDivider := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "TimeDivider",
			Description: "Time divider",
			Other:       " at ",
		},
	})

	usersText, weekdays, time, err := utils.SplitTimeTalbeCommand(params, daysDivider, timeDivider)
	if err != nil {
		return DisplayHelpText("add_timetable")
	}
	users := strings.Split(usersText, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			wrongUsernameError := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "WrongUsernameError",
					Description: "Displays message when username is misspelled",
					Other:       "Seems like you misspelled username. Please, check and try command again!",
				},
			})
			totalString += wrongUsernameError
			continue

		}
		userID, userName := utils.SplitUser(u)

		m, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			m, err = ba.Bot.DB.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: channelID,
			})
			if err != nil {
				continue
			}
		}

		tt, err := ba.Bot.DB.SelectTimeTable(m.ID)
		if err != nil {
			logrus.Infof("Timetable for this standuper does not exist. Creating...")
			ttNew, err := ba.Bot.DB.CreateTimeTable(model.TimeTable{
				ChannelMemberID: m.ID,
			})
			ttNew = utils.PrepareTimeTable(ttNew, weekdays, time)
			ttNew, err = ba.Bot.DB.UpdateTimeTable(ttNew)
			if err != nil {
				canNotUpdateTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "CanNotUpdateTimetable",
						Description: "",
						Other:       "Could not update timetable for user {{.user}}: {{.error}}\n",
					},
					TemplateData: map[string]interface{}{
						"user":  userName,
						"error": err,
					},
				})
				totalString += canNotUpdateTimetable
				continue

			}
			logrus.Infof("Timetable created id:%v", ttNew.ID)
			timetableCreated := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TimetableCreated",
					Description: "Show message that timetable for user created",
					Other:       "Timetable for <@{{.user}}> created: {{.timetable}} \n",
				},
				TemplateData: map[string]interface{}{
					"user":      userID,
					"timetable": ttNew.Show(),
				},
			})
			totalString += timetableCreated
			continue
		}
		tt = utils.PrepareTimeTable(tt, weekdays, time)
		tt, err = ba.Bot.DB.UpdateTimeTable(tt)
		if err != nil {
			canNotUpdateTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "CanNotUpdateTimetable",
					Description: "",
					Other:       "Could not update timetable for user <@{{.user}}>: {{.error}}\n",
				},
				TemplateData: map[string]interface{}{
					"user":  userName,
					"error": err,
				},
			})
			totalString += canNotUpdateTimetable
			continue
		}
		logrus.Infof("Timetable updated id:%v", tt.ID)
		timetableUpdated := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "TimetableUpdated",
				Description: "",
				Other:       "Timetable for <@{{.user}}> updated: {{.timetable}} \n",
			},
			TemplateData: map[string]interface{}{
				"user":      userID,
				"timetable": tt.Show(),
			},
		})
		totalString += timetableUpdated
	}
	return totalString
}

func (ba *BotAPI) showTimeTable(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	var totalString string
	//add parsing of params
	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			wrongUsernameError := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "WrongUsernameError",
					Description: "Displays message when username is misspelled",
					Other:       "Seems like you misspelled username. Please, check and try command again!",
				},
			})
			totalString += wrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			notAStanduper := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "NotAStanduper",
					Description: "Display message when user not a standuper",
					Other:       "Seems like <@{{.user}}> is not even assigned as standuper in this channel!\n",
				},
				TemplateData: map[string]interface{}{
					"user": userName,
				},
			})
			totalString += notAStanduper
			continue

		}
		tt, err := ba.Bot.DB.SelectTimeTable(m.ID)
		if err != nil {
			noTimetableSet := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "NoTimetableSet",
					Description: "Display message when user doesn't have a timetable",
					Other:       "<@{{.user}}> does not have a timetable!\n",
				},
				TemplateData: map[string]interface{}{
					"user": userName,
				},
			})
			totalString += noTimetableSet
			continue

		}
		timetableShow := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "TimetableShow",
				Description: "",
				Other:       "Timetable for <@{{.user}}> is: {{.timetable}}\n",
			},
			TemplateData: map[string]interface{}{
				"user":      userName,
				"timetable": tt.Show(),
			},
		})
		totalString += timetableShow
	}
	return totalString
}

func (ba *BotAPI) removeTimeTable(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AccessAtLeastPM",
				Description: "Display warning that role must be at least pm",
				Other:       "Access Denied! You need to be at least PM in this project to use this command!",
			},
		})
		return accessAtLeastPM
	}

	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			wrongUsernameError := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "WrongUsernameError",
					Description: "Displays message when username is misspelled",
					Other:       "Seems like you misspelled username. Please, check and try command again!",
				},
			})
			totalString += wrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			notAStanduper := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "NotAStanduper",
					Description: "Display message when user not a standuper",
					Other:       "Seems like <@{{.user}}> is not even assigned as standuper in this channel!\n",
				},
				TemplateData: map[string]interface{}{
					"user": userName,
				},
			})
			totalString += notAStanduper
			continue

		}
		tt, err := ba.Bot.DB.SelectTimeTable(m.ID)
		if err != nil {
			noTimetableSet := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "NoTimetableSet",
					Description: "Display message when user doesn't have a timetable",
					Other:       "<@{{.user}}> does not have a timetable!\n",
				},
				TemplateData: map[string]interface{}{
					"user": userName,
				},
			})
			totalString += noTimetableSet
			continue
		}
		err = ba.Bot.DB.DeleteTimeTable(tt.ID)
		if err != nil {
			canNotDeleteTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "CanNotDeleteTimetable",
					Description: "Displays a message when a timetable deletion error occurs.",
					Other:       "Could not delete timetable for user <@{{.user}}>\n",
				},
				TemplateData: map[string]interface{}{
					"user": userName,
				},
			})
			totalString += canNotDeleteTimetable
			continue

		}
		timetableDeleted := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "TimetableDeleted",
				Description: "Displays message when timetable removed",
				Other:       "Timetable removed for <@{{.user}}>\n",
			},
			TemplateData: map[string]interface{}{
				"user": userName,
			},
		})
		totalString += timetableDeleted
	}
	return totalString
}
