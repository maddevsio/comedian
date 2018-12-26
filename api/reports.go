package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/collector"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) generateReportOnProject(accessLevel int, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		wrongProject := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "WrongProject",
				Description: "Displays message when project name is not found",
				Other:       "Invalid project name!",
			},
		})
		logrus.Errorf("rest: GetChannelID failed: %v\n", err)
		return wrongProject
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
		reportNoData := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ReportNoData",
				Description: "Displays message when there is no standup data for this period",
				Other:       "No standup data for this period",
			},
		})
		reportNoData += "\n"
		text += reportNoData
		return text

	}
	for _, t := range report.ReportBody {
		text += t.Text
		if ba.Bot.CP.CollectorEnabled {
			cd, err := collector.GetCollectorData(ba.Bot, "projects", channel.ChannelName, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}

			totalCommits := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TotalCommits",
					Description: "Shows a total commits",
					Other:       "Total commits for period: {{.commits}}",
				},
				TemplateData: map[string]interface{}{
					"commits": cd.Commits,
				},
			})
			totalCommits += "\n"
			totalWorklogs := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "totalWorklogs",
					Description: "Shows a total worklogs",
					Other:       "Total worklogs for period: {{.worklogs}}",
				},
				TemplateData: map[string]interface{}{
					"worklogs": utils.SecondsToHuman(cd.Worklogs),
				},
			})
			totalWorklogs += "\n"
			//"\nTotal commits for period: {{.commits}}\nTotal worklogs for period: {{.worklogs}}"
			reportOnProjectCollectorData := "\n" + totalCommits + totalWorklogs
			text += reportOnProjectCollectorData

		}
	}
	return text
}

func (ba *BotAPI) generateReportOnUser(accessLevel int, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		userDoesNotExist := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "userDoesNotExist",
				Description: "Displays message if user doesn't exist",
				Other:       "User does not exist!",
			},
		})
		return userDoesNotExist

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
		reportNoData := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ReportNoData",
				Description: "Displays message when there is no standup data for this period",
				Other:       "No standup data for this period",
			},
		})
		text += reportNoData
		return text
	}
	for _, t := range report.ReportBody {
		text += t.Text
		if ba.Bot.CP.CollectorEnabled {
			cd, err := collector.GetCollectorData(ba.Bot, "users", user.UserID, t.Date.Format("2006-01-02"), t.Date.Format("2006-01-02"))
			if err != nil {
				continue
			}

			totalCommits := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TotalCommits",
					Description: "Shows a total commits",
					Other:       "Total commits for period: {{.commits}}",
				},
				TemplateData: map[string]interface{}{
					"commits": cd.Commits,
				},
			})
			totalCommits += "\n"
			loggedHours := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "LoggedHours",
					Description: "Shows a logged hours",
					Other:       "Logged Hours: {{.worklogs}}",
				},
				TemplateData: map[string]interface{}{
					"worklogs": utils.SecondsToHuman(cd.Worklogs),
				},
			})
			loggedHours += "\n\n"
			//"\nTotal commits for period: %v\nLogged Hours: %v\n\n"
			reportCollectorDataUser := "\n" + totalCommits + loggedHours
			text += reportCollectorDataUser
		}
	}
	return text
}

func (ba *BotAPI) generateReportOnUserInProject(accessLevel int, params string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		wrongProjectName := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "WrongProjectName",
				Description: "Displays message if project doesn't exist",
				Other:       "Wrong project title!",
			},
		})
		return wrongProjectName
	}

	channel, err := ba.Bot.DB.SelectChannel(channelID)
	if err != nil {
		cantSelectChannelInReport := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "cantSelectChannelInReport",
				Description: "Displays message if you cannot find a channel for a report.",
				Other:       "Could not select channel!",
			},
		})
		cantSelectChannelInReport += "\n"
		cantSelectChannelInReport += DisplayHelpText("report_on_user_in_project")
		logrus.Errorf("rest: SelectChannel failed: %v\n", err)
		return cantSelectChannelInReport

	}
	username, err := GetUserNameFromString(commandParams[0])
	if err != nil {
		wrongName := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "WrongName",
				Description: "Displays if the username cannot be obtained from the parameters",
				Other:       "Could not get user!",
			},
		})
		wrongName += "\n"
		wrongName += DisplayHelpText("report_on_user_in_project")
		return wrongName
	}

	user, err := ba.Bot.DB.SelectUserByUserName(username)
	if err != nil {
		noSuchUserInWorkspace := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "NoSuchUserInWorkspace",
				Description: "Displays message if user is not in slack workspace",
				Other:       "No such user in your slack!",
			},
		})
		return noSuchUserInWorkspace
	}
	member, err := ba.Bot.DB.FindChannelMemberByUserName(user.UserName, channelID)
	if err != nil {
		canNotFindMember := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "CanNotFindMember",
				Description: "Displays message if user doesn't have any role in the channel",
				Other:       "<@{{.user}}> does not have any role in this channel",
			},
			TemplateData: map[string]interface{}{
				"user": user.UserID,
			},
		})
		canNotFindMember += "\n"
		return canNotFindMember
	}

	dateFrom, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		errorParsingFromDate := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ErrorParsingFromDate",
				Description: "Displays message when occurs error parsing date for report",
				Other:       "Could not parse date from!",
			},
		})
		errorParsingFromDate += "\n"
		errorParsingFromDate += DisplayHelpText("report_on_user_in_project")
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return errorParsingFromDate

	}
	dateTo, err := time.Parse("2006-01-02", commandParams[3])
	if err != nil {
		errorParsingToDate := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ErrorParsingToDate",
				Description: "Displays message when occurs error parsing date for report",
				Other:       "Could not parse date from!",
			},
		})
		errorParsingToDate += "\n"
		errorParsingToDate += DisplayHelpText("report_on_user_in_project")
		logrus.Errorf("rest: time.Parse failed: %v\n", err)
		return errorParsingToDate

	}

	report, err := ba.StandupReportByProjectAndUser(channel, member.UserID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("rest: StandupReportByProjectAndUser failed: %v\n", err)
		return err.Error()
	}

	text := ""
	text += report.ReportHead
	if len(report.ReportBody) == 0 {
		reportNoData := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ReportNoData",
				Description: "Displays message when there is no standup data for this period",
				Other:       "No standup data for this period",
			},
		})
		text += reportNoData + "\n"
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

			totalCommits := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "TotalCommits",
					Description: "Shows a total commits",
					Other:       "Total commits for period: {{.commits}}",
				},
				TemplateData: map[string]interface{}{
					"commits": cd.Commits,
				},
			})
			totalCommits += "\n"
			loggedHours := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "LoggedHours",
					Description: "Shows a logged hours",
					Other:       "Logged Hours: {{.worklogs}}",
				},
				TemplateData: map[string]interface{}{
					"worklogs": utils.SecondsToHuman(cd.Worklogs),
				},
			})
			loggedHours += "\n\n"
			//"\nTotal commits for period: %v\nLogged Hours: %v\n\n"
			reportCollectorDataUser := "\n" + totalCommits + loggedHours

			text += reportCollectorDataUser
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
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	report := model.Report{}
	reportOnProjectHead := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnProjectHead",
			Description: "Displays report on project in date",
			Other:       "Full Report on project #{{.channel}} from {{.dateFrom}} to {{.dateTo}}:",
		},
		TemplateData: map[string]interface{}{
			"channel":  channel.ChannelName,
			"dateFrom": dateFrom.Format("2006-01-02"),
			"dateTo":   dateTo.Format("2006-01-02"),
		},
	})
	reportOnProjectHead += "\n\n"
	report.ReportHead = reportOnProjectHead
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
				userDidNotStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "UserDidNotStandup",
						Description: "Displays message when user didn't submit standup",
						Other:       "<@{{.user}}> did not submit standup!",
					},
					TemplateData: map[string]interface{}{
						"user": member.UserID,
					},
				})
				userDidNotStandup += "\n"
				dayInfo += userDidNotStandup
			} else {
				standup, err := ba.Bot.DB.SelectStandupsFiltered(member.UserID, channel.ChannelID, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
					continue
				}
				userDidStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "UserDidStandup",
						Description: "Displays message if user successfully submitted standup",
						Other:       "<@{{.user}}> submitted standup: ",
					},
					TemplateData: map[string]interface{}{
						"user": member.UserID,
					},
				})
				dayInfo += userDidStandup
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			reportDate := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "ReportDate",
					Description: "Displays report for date",
					Other:       "Report for: {{.dateFrom}}",
				},
				TemplateData: map[string]interface{}{
					"dateFrom": dateFrom.Format("2006-01-02"),
				},
			})
			reportDate += "\n"
			text := reportDate
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}

	}
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (ba *BotAPI) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time) (model.Report, error) {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	report := model.Report{}
	reportOnUserHead := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnUserHead",
			Description: "Displays head of report on user",
			Other:       "Full Report on user <@{{.user}}> from {{.dateFrom}} to {{.dateTo}}:",
		},
		TemplateData: map[string]interface{}{
			"user":     slackUserID,
			"dateFrom": dateFrom.Format("2006-01-02"),
			"dateTo":   dateTo.Format("2006-01-02"),
		},
	})
	reportOnUserHead += "\n\n"
	report.ReportHead = reportOnUserHead
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
				userDidNotStandupInChannel := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "UserDidNotStandupInChannel",
						Description: "Displays message when user didn't submit standup in the channel",
						Other:       "In #{{.channel}} <@{{.user}}> did not submit standup!",
					},
					TemplateData: map[string]interface{}{
						"channel": channelName,
						"user":    slackUserID,
					},
				})
				userDidNotStandupInChannel += "\n"
				dayInfo += userDidNotStandupInChannel
			} else {
				standup, err := ba.Bot.DB.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting.go reportByUser SelectStandupsFiltered failed: %v", err)
				}
				userDidStandupInChannel := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "UserDidStandupInChannel",
						Description: "Displays message when user submitted standup in the channel",
						Other:       "In #{{.channel}} <@{{.user}}> submitted standup: ",
					},
					TemplateData: map[string]interface{}{
						"channel": channelName,
						"user":    slackUserID,
					},
				})
				dayInfo += userDidStandupInChannel
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			reportDate := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "ReportDate",
					Description: "Displays report for date",
					Other:       "Report for: {{.dateFrom}}",
				},
				TemplateData: map[string]interface{}{
					"dateFrom": dateFrom.Format("2006-01-02"),
				},
			})
			reportDate += "\n"
			text := reportDate
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (ba *BotAPI) StandupReportByProjectAndUser(channel model.Channel, slackUserID string, dateFrom, dateTo time.Time) (model.Report, error) {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	report := model.Report{}
	reportOnProjectAndUserHead := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnProjectAndUserHead",
			Description: "Displays report on user in project",
			Other:       "Report on user <@{{.user}}> in project #{{.channel}} from {{.dateFrom}} to {{.dateTo}}",
		},
		TemplateData: map[string]interface{}{
			"user":     slackUserID,
			"channel":  channel.ChannelName,
			"dateFrom": dateFrom.Format("2006-01-02"),
			"dateTo":   dateTo.Format("2006-01-02"),
		},
	})
	reportOnProjectAndUserHead += "\n\n"
	report.ReportHead = reportOnProjectAndUserHead
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
			userDidNotStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "UserDidNotStandup",
					Description: "Displays message when user didn't submit standup",
					Other:       "<@{{.user}}> did not submit standup!",
				},
				TemplateData: map[string]interface{}{
					"user": slackUserID,
				},
			})
			userDidNotStandup += "\n"
			dayInfo += userDidNotStandup
			dayInfo += "\n"
		} else {
			standup, err := ba.Bot.DB.SelectStandupsFiltered(slackUserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProjectAndUser SelectStandupsFiltered failed: %v", err)
				continue
			}
			userDidStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "UserDidStandup",
					Description: "Displays message if user successfully submitted standup",
					Other:       "<@{{.user}}> submitted standup: ",
				},
				TemplateData: map[string]interface{}{
					"user": slackUserID,
				},
			})
			dayInfo += userDidStandup
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
		}
		if dayInfo != "" {
			reportDate := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "ReportDate",
					Description: "Displays report for date",
					Other:       "Report for: {{.dateFrom}}",
				},
				TemplateData: map[string]interface{}{
					"dateFrom": dateFrom.Format("2006-01-02"),
				},
			})
			reportDate += "\n"
			text := reportDate
			text += dayInfo
			rbc := model.ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}
