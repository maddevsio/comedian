package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (r *REST) addTimeTable(accessLevel int, channelID, params string) string {
	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		return r.slack.Translate.AccessAtLeastPM
	}

	usersText, weekdays, time, err := utils.SplitTimeTalbeCommand(params, r.slack.Translate.DaysDivider, r.slack.Translate.TimeDivider)
	if err != nil {
		return r.displayHelpText("add_timetable")
	}
	users := strings.Split(usersText, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += r.slack.Translate.WrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			m, err = r.db.CreateChannelMember(model.ChannelMember{
				UserID:    userID,
				ChannelID: channelID,
			})
			if err != nil {
				continue
			}
		}

		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			logrus.Infof("Timetable for this standuper does not exist. Creating...")
			ttNew, err := r.db.CreateTimeTable(model.TimeTable{
				ChannelMemberID: m.ID,
			})
			ttNew = utils.PrepareTimeTable(ttNew, weekdays, time)
			ttNew, err = r.db.UpdateTimeTable(ttNew)
			if err != nil {
				totalString += fmt.Sprintf(r.slack.Translate.CanNotUpdateTimetable, userName, err)
				continue
			}
			logrus.Infof("Timetable created id:%v", ttNew.ID)
			totalString += fmt.Sprintf(r.slack.Translate.TimetableCreated, userID, ttNew.Show())
			continue
		}
		tt = utils.PrepareTimeTable(tt, weekdays, time)
		tt, err = r.db.UpdateTimeTable(tt)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.CanNotUpdateTimetable, userName, err)
			continue
		}
		logrus.Infof("Timetable updated id:%v", tt.ID)
		totalString += fmt.Sprintf(r.slack.Translate.TimetableUpdated, userID, tt.Show())
	}
	return totalString
}

func (r *REST) showTimeTable(accessLevel int, channelID, params string) string {
	var totalString string
	//add parsing of params
	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += r.slack.Translate.WrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.NotAStanduper, userName)
			continue
		}
		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.NoTimetableSet, userName)
			continue
		}
		totalString += fmt.Sprintf(r.slack.Translate.TimetableShow, userName, tt.Show())
	}
	return totalString
}

func (r *REST) removeTimeTable(accessLevel int, channelID, params string) string {
	//add parsing of params
	var totalString string
	if accessLevel > 3 {
		return r.slack.Translate.AccessAtLeastPM
	}

	users := strings.Split(params, " ")
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	for _, u := range users {
		if !rg.MatchString(u) {
			totalString += r.slack.Translate.WrongUsernameError
			continue
		}
		userID, userName := utils.SplitUser(u)

		m, err := r.db.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.NotAStanduper, userName)
			continue
		}
		tt, err := r.db.SelectTimeTable(m.ID)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.NoTimetableSet, userName)
			continue
		}
		err = r.db.DeleteTimeTable(tt.ID)
		if err != nil {
			totalString += fmt.Sprintf(r.slack.Translate.CanNotDeleteTimetable, userName)
			continue
		}
		totalString += fmt.Sprintf(r.slack.Translate.TimetableDeleted, userName)
	}
	return totalString
}
