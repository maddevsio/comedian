package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) addTimeTable(accessLevel int, channelID, params string) string {
	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		return ba.Bot.Translate.AccessAtLeastPM
	}

	usersText, weekdays, time, err := utils.SplitTimeTalbeCommand(params, ba.Bot.Translate.DaysDivider, ba.Bot.Translate.TimeDivider)
	if err != nil {
		return DisplayHelpText("add_timetable")
	}
	users := strings.Split(usersText, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += ba.Bot.Translate.WrongUsernameError
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
				totalString += fmt.Sprintf(ba.Bot.Translate.CanNotUpdateTimetable, userName, err)
				continue
			}
			logrus.Infof("Timetable created id:%v", ttNew.ID)
			totalString += fmt.Sprintf(ba.Bot.Translate.TimetableCreated, userID, ttNew.Show())
			continue
		}
		tt = utils.PrepareTimeTable(tt, weekdays, time)
		tt, err = ba.Bot.DB.UpdateTimeTable(tt)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.CanNotUpdateTimetable, userName, err)
			continue
		}
		logrus.Infof("Timetable updated id:%v", tt.ID)
		totalString += fmt.Sprintf(ba.Bot.Translate.TimetableUpdated, userID, tt.Show())
	}
	return totalString
}

func (ba *BotAPI) showTimeTable(accessLevel int, channelID, params string) string {
	var totalString string
	//add parsing of params
	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += ba.Bot.Translate.WrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.NotAStanduper, userName)
			continue
		}
		tt, err := ba.Bot.DB.SelectTimeTable(m.ID)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.NoTimetableSet, userName)
			continue
		}
		totalString += fmt.Sprintf(ba.Bot.Translate.TimetableShow, userName, tt.Show())
	}
	return totalString
}

func (ba *BotAPI) removeTimeTable(accessLevel int, channelID, params string) string {
	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		return ba.Bot.Translate.AccessAtLeastPM
	}

	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += ba.Bot.Translate.WrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.NotAStanduper, userName)
			continue
		}
		tt, err := ba.Bot.DB.SelectTimeTable(m.ID)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.NoTimetableSet, userName)
			continue
		}
		err = ba.Bot.DB.DeleteTimeTable(tt.ID)
		if err != nil {
			totalString += fmt.Sprintf(ba.Bot.Translate.CanNotDeleteTimetable, userName)
			continue
		}
		totalString += fmt.Sprintf(ba.Bot.Translate.TimetableDeleted, userName)
	}
	return totalString
}
