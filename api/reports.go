package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/collector"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) generateReportOnProject(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 3 {
		return DisplayHelpText("report_on_project")
	}
	channelName, err := GetChannelNameFromString(commandParams[0])
	if err != nil {
		DisplayHelpText("report_on_project")
	}
	channelID, err := ba.Bot.DB.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return "Неверное название проекта!"
	}

	channel, err := ba.Bot.DB.SelectChannel(channelID)
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

	report, err := ba.StandupReportByProject(channel, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProject: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += ba.Bot.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if ba.Bot.CP.CollectorEnabled {
			cd, err := collector.GetCollectorData(ba.Bot, "projects", channel.ChannelName, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(ba.Bot.Translate.ReportOnProjectCollectorData, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}

func (ba *BotAPI) generateReportOnUser(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 3 {
		return DisplayHelpText("report_on_user")
	}
	username, err := GetUserNameFromString(commandParams[0])
	if err != nil {
		DisplayHelpText("report_on_user")
	}
	user, err := ba.Bot.DB.SelectUserByUserName(username)
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

	report, err := ba.StandupReportByUser(user.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByUser failed: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += ba.Bot.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if ba.Bot.CP.CollectorEnabled {
			cd, err := collector.GetCollectorData(ba.Bot, "users", user.UserID, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(ba.Bot.Translate.ReportCollectorDataUser, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}

func (ba *BotAPI) generateReportOnUserInProject(accessLevel int, params string) string {
	commandParams := strings.Fields(params)
	if len(commandParams) != 4 {
		return DisplayHelpText("report_on_user_in_project")
	}
	channelName, err := GetChannelNameFromString(commandParams[1])
	if err != nil {
		DisplayHelpText("report_on_user_in_project")
	}
	channelID, err := ba.Bot.DB.GetChannelID(channelName)
	if err != nil {
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return ba.Bot.Translate.WrongProjectName
	}

	channel, err := ba.Bot.DB.SelectChannel(channelID)
	if err != nil {
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return "Could not select channlel! \n" + DisplayHelpText("report_on_user_in_project")
	}
	username, err := GetUserNameFromString(commandParams[0])
	if err != nil {
		return "Could not get user! \n" + DisplayHelpText("report_on_user_in_project")
	}

	user, err := ba.Bot.DB.SelectUserByUserName(username)
	if err != nil {
		return ba.Bot.Translate.NoSuchUserInWorkspace
	}
	member, err := ba.Bot.DB.FindChannelMemberByUserName(user.UserName, channelID)
	if err != nil {
		return fmt.Sprintf(ba.Bot.Translate.CanNotFindMember, user.UserID)
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return "Could not parse date from! \n" + DisplayHelpText("report_on_user_in_project")
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[3])
	if err != nil {
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return "Could not parse date to! \n" + DisplayHelpText("report_on_user_in_project")
	}

	report, err := ba.StandupReportByProjectAndUser(channel, member.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProjectAndUser failed: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		text += ba.Bot.Translate.ReportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if ba.Bot.CP.CollectorEnabled {
			data := fmt.Sprintf("%v/%v", member.UserID, channel.ChannelName)
			cd, err := collector.GetCollectorData(ba.Bot, "user-in-project", data, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}
			text += fmt.Sprintf(ba.Bot.Translate.ReportCollectorDataUser, cd.Commits, utils.SecondsToHuman(cd.Worklogs))
		}
	}
	return text
}

//GetChannelNameFromString return channel name
func GetChannelNameFromString(channel string) (string, error) {
	var channelName string
	rg, err := regexp.Compile("<#([a-z0-9]+)|([a-z0-9]+)>")
	if err != nil {
		logrus.Error("Error in regexp.Compile")
	}
	//if <#channelname>
	if rg.MatchString(channel) {
		_, channelName = utils.SplitChannel(channel)
	} else {
		//#channelname
		channelName = strings.Replace(channel, "#", "", -1)
	}
	return channelName, err
}

//GetUserNameFromString return username
func GetUserNameFromString(user string) (string, error) {
	var username string
	rg, err := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")
	if err != nil {
		logrus.Error("Error in regexp.Compile")
	}
	//if <@userid|username>
	if rg.MatchString(user) {
		_, username = utils.SplitUser(user)
	} else {
		//@username
		username = strings.Replace(user, "@", "", -1)
	}
	return username, err
}

// StandupReportByProject creates a standup report for a specified period of time
func (ba *BotAPI) StandupReportByProject(channel model.Channel, dateFrom, dateTo time.Time) (model.Report, error) {
	report := model.Report{}
	report.ReportHead = fmt.Sprintf(ba.Bot.Translate.ReportOnProjectHead, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("SetupDays failed: %v", err)
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		chanMembers, err := ba.Bot.DB.ListChannelMembers(channel.ChannelID)
		if err != nil || len(chanMembers) == 0 {
			continue
		}
		dayInfo := ""
		for _, member := range chanMembers {
			if !ba.Bot.DB.MemberShouldBeTracked(member.ID, dateFrom) {
				logrus.Infof("member should not be tracked: %v", member.UserID)
				continue
			}
			userIsNonReporter, err := ba.Bot.DB.IsNonReporter(member.UserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProject IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidNotStandup, member.UserID)
			} else {
				standup, err := ba.Bot.DB.SelectStandupsFiltered(member.UserID, channel.ChannelID, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
					continue
				}
				dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidStandup, member.UserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			text := fmt.Sprintf(ba.Bot.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}

	}
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (ba *BotAPI) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time) (model.Report, error) {
	report := model.Report{}
	report.ReportHead = fmt.Sprintf(ba.Bot.Translate.ReportOnUserHead, slackUserID, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		channels, err := ba.Bot.DB.GetUserChannels(slackUserID)
		if err != nil || len(channels) == 0 {
			continue
		}
		dayInfo := ""
		for _, channel := range channels {
			channelName, err := ba.Bot.DB.GetChannelName(channel)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser GetChannelName failed: %v", err)
				continue
			}
			member, err := ba.Bot.DB.FindChannelMemberByUserID(slackUserID, channel)
			if err != nil {
				logrus.Infof("FindChannelMemberByUserID failed: %v", err)
				continue
			}
			if !ba.Bot.DB.MemberShouldBeTracked(member.ID, dateFrom) {
				logrus.Infof("member should not be tracked: %v", slackUserID)
				continue
			}
			userIsNonReporter, err := ba.Bot.DB.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidNotStandupInChannel, channelName, slackUserID)
			} else {
				standup, err := ba.Bot.DB.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting.go reportByUser SelectStandupsFiltered failed: %v", err)
				}
				dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidStandupInChannel, channelName, slackUserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			text := fmt.Sprintf(ba.Bot.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (ba *BotAPI) StandupReportByProjectAndUser(channel model.Channel, slackUserID string, dateFrom, dateTo time.Time) (model.Report, error) {
	report := model.Report{}
	report.ReportHead = fmt.Sprintf(ba.Bot.Translate.ReportOnProjectAndUserHead, slackUserID, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dayInfo := ""
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		logrus.Infof("reportByProjectAndUser: dateFrom: '%v', dateTo: '%v'", dateFrom, dateTo)
		member, err := ba.Bot.DB.FindChannelMemberByUserID(slackUserID, channel.ChannelID)
		if err != nil {
			logrus.Infof("FindChannelMemberByUserID failed: %v", err)
			continue
		}
		if !ba.Bot.DB.MemberShouldBeTracked(member.ID, dateFrom) {
			logrus.Infof("member should not be tracked: %v", slackUserID)
			continue
		}
		userIsNonReporter, err := ba.Bot.DB.IsNonReporter(slackUserID, channel.ChannelID, dateFrom, dateTo)
		if err != nil {
			logrus.Errorf("reporting.go reportByProjectAndUser IsNonReporter failed: %v", err)
			continue
		}
		if userIsNonReporter {
			dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidNotStandup, slackUserID)
			dayInfo += "\n"
		} else {
			standup, err := ba.Bot.DB.SelectStandupsFiltered(slackUserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProjectAndUser SelectStandupsFiltered failed: %v", err)
				continue
			}
			dayInfo += fmt.Sprintf(ba.Bot.Translate.UserDidStandup, slackUserID)
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
		}
		if dayInfo != "" {
			text := fmt.Sprintf(ba.Bot.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}
