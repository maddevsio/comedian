package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/collector"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (r *REST) generateReportOnProject(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 3 {
		return r.conf.Translate.WrongNArgs
	}
	var channelName string
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	//if <#channelname>
	if rg.MatchString(commandParams[0]) {
		_, channelName = utils.SplitChannel(commandParams[0])
	} else {
		//#channelname
		channelName = strings.Replace(commandParams[0], "#", "", -1)
	}
	channelID, err := r.db.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return "Неверное название проекта!"
	}

	channel, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return err.Error()
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}

	report, err := r.report.StandupReportByProject(channel, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProject: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.CollectorEnabled {
			cd, err := collector.GetCollectorData(r.conf, "projects", channel.ChannelName, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportOnProjectCollectorData, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}

func (r *REST) generateReportOnUser(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 3 {
		return r.conf.Translate.WrongNArgs
	}
	var username string
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	//if <@userid|username>
	if rg.MatchString(commandParams[0]) {
		_, username = utils.SplitUser(commandParams[0])
	} else {
		//@username
		username = strings.Replace(commandParams[0], "@", "", -1)
	}
	user, err := r.db.SelectUserByUserName(username)
	if err != nil {
		return "User does not exist!"
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}

	report, err := r.report.StandupReportByUser(user.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByUser failed: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.CollectorEnabled {
			cd, err := collector.GetCollectorData(r.conf, "users", user.UserID, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportCollectorDataUser, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}

func (r *REST) generateReportOnUserInProject(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 4 {
		return r.conf.Translate.WrongNArgs
	}
	var channelName string
	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	if rg.MatchString(commandParams[0]) {
		//<#channelid|channelname>
		_, channelName = utils.SplitChannel(commandParams[0])
	} else {
		//#channelname
		channelName = strings.Replace(commandParams[0], "#", "", -1)
	}
	channelID, err := r.db.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return r.conf.Translate.WrongProjectName
	}

	channel, err := r.db.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return err.Error()
	}
	var username string
	if rg.MatchString(commandParams[1]) {
		//<@userid|username>
		_, username = utils.SplitUser(commandParams[1])
	} else {
		//@username
		username = strings.Replace(commandParams[1], "@", "", -1)
	}

	user, err := r.db.SelectUserByUserName(username)
	if err != nil {
		return r.conf.Translate.NoSuchUserInWorkspace
	}
	member, err := r.db.FindChannelMemberByUserName(user.UserName, channelID)
	if err != nil {
		return fmt.Sprintf(r.conf.Translate.CanNotFindMember, user.UserID)
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[3])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return err.Error()
	}

	report, err := r.report.StandupReportByProjectAndUser(channel, member.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProjectAndUser failed: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += r.conf.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if r.conf.CollectorEnabled {
			data := fmt.Sprintf("%v/%v", member.UserID, channel.ChannelName)
			cd, err := collector.GetCollectorData(r.conf, "user-in-project", data, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(r.conf.Translate.ReportCollectorDataUser, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}
