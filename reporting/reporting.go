package reporting

import (
	"fmt"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/maddevsio/comedian/teammonitoring"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

//Reporter provides db and translation to functions
type Reporter struct {
	s    *chat.Slack
	db   storage.Storage
	conf config.Config
}

//Report used to generate report structure
type Report struct {
	ReportHead string
	ReportBody []ReportBodyContent
}

//ReportBodyContent used to generate report body content
type ReportBodyContent struct {
	Date time.Time
	Text string
}

// NewReporter creates a new reporter instance
func NewReporter(slack *chat.Slack) *Reporter {
	reporter := &Reporter{s: slack, db: slack.DB, conf: slack.Conf}
	return reporter
}

// Start starts all team monitoring treads
func (r *Reporter) Start() {
	gocron.Every(1).Day().At(r.conf.ReportTime).Do(r.displayYesterdayTeamReport)
	gocron.Every(1).Sunday().At(r.conf.ReportTime).Do(r.displayWeeklyTeamReport)

}

// teamReport generates report on users who submit standups
func (r *Reporter) displayYesterdayTeamReport() {
	var allReports []slack.Attachment

	channels, err := r.db.GetAllChannels()
	if err != nil {
		logrus.Errorf("GetAllChannels failed: %v", err)
		return
	}

	for _, channel := range channels {
		var attachments []slack.Attachment

		channelMembers, err := r.db.ListChannelMembers(channel.ChannelID)
		if err != nil {
			logrus.Errorf("ListChannelMembers failed for channel %v: %v", channel.ChannelName, err)
			continue
		}

		if len(channelMembers) == 0 {
			logrus.Infof("Skip %v channel", channel.ChannelID)
			continue
		}

		for _, member := range channelMembers {
			attachment := r.generateReportAttachment(member, channel)
			if len(attachment.Fields) == 0 {
				continue
			}
			attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)

			attachments = append(attachments, attachment)
		}

		r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeader, attachments)

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeader, allReports)
}

// teamReport generates report on users who submit standups
func (r *Reporter) displayWeeklyTeamReport() {
	var allReports []slack.Attachment

	channels, err := r.db.GetAllChannels()
	if err != nil {
		logrus.Errorf("GetAllChannels failed: %v", err)
		return
	}

	for _, channel := range channels {
		var attachments []slack.Attachment

		channelMembers, err := r.db.ListChannelMembers(channel.ChannelID)
		if err != nil {
			logrus.Errorf("ListChannelMembers failed for channel %v: %v", channel.ChannelName, err)
			continue
		}

		if len(channelMembers) == 0 {
			logrus.Infof("Skip %v channel", channel.ChannelID)
			continue
		}

		for _, member := range channelMembers {
			attachment := r.generateWeeklyReportAttachment(member, channel)
			if len(attachment.Fields) == 0 {
				continue
			}
			attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)

			attachments = append(attachments, attachment)
		}

		r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeaderWeekly, attachments)

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeaderWeekly, allReports)
}

func (r *Reporter) generateReportAttachment(member model.ChannelMember, project model.Channel) slack.Attachment {

	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, -1)

	dataOnUser, dataOnUserInProject, collectorError := r.GetCollectorDataOnMember(member, project, startDate, endDate)

	if collectorError != nil {
		userFull, _ := r.db.SelectUser(member.UserID)
		fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
		r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
	}

	//if report is being made for weekends, and user did not do anywork on these days, generate no report
	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		if dataOnUser.Worklogs == 0 && dataOnUser.TotalCommits == 0 {
			logrus.Infof("User %v in %v did not do anything yesterday. Skip!", member.UserID, project.ChannelName)
			return slack.Attachment{}
		}
	}

	startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

	isNonReporter, err := r.db.IsNonReporter(member.UserID, member.ChannelID, startDateTime, endDateTime)
	if err != nil {
		logrus.Infof("User is non reporter failed: %v", err)
	}

	fieldValue, points := r.PrepareAttachment(member, dataOnUser, dataOnUserInProject, isNonReporter, collectorError)

	return r.GenerateAttachment(fieldValue, points)
}

func (r *Reporter) generateWeeklyReportAttachment(member model.ChannelMember, project model.Channel) slack.Attachment {

	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now().AddDate(0, 0, -1)

	dataOnUser, dataOnUserInProject, collectorError := r.GetCollectorDataOnMember(member, project, startDate, endDate)

	if collectorError != nil {
		userFull, _ := r.db.SelectUser(member.UserID)
		fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
		r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
	}

	fieldValue, points := r.PrepareWeeklyAttachment(member, dataOnUser, dataOnUserInProject, collectorError)

	var attachment slack.Attachment
	var attachmentFields []slack.AttachmentField

	//if there is nothing to show, do not create attachment
	if fieldValue == "" {
		logrus.Infof("Noting to show on attachment on member: %v in %v. Skip!", member.UserID, member.ChannelID)
		return attachment
	}

	attachmentFields = append(attachmentFields, slack.AttachmentField{
		Value: fieldValue,
		Short: false,
	})

	attachment.Text = ""
	switch p := points; p {
	case 0:
		attachment.Color = "danger"
	case 1:
		attachment.Color = "warning"
	case 2:
		attachment.Color = "good"
	}

	attachment.Fields = attachmentFields
	return attachment
}

func (r *Reporter) GetCollectorDataOnMember(member model.ChannelMember, project model.Channel, startDate, endDate time.Time) (teammonitoring.CollectorData, teammonitoring.CollectorData, error) {
	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	dataOnUser, err := teammonitoring.GetCollectorData(r.conf, "users", member.UserID, dateFrom, dateTo)
	if err != nil {
		return teammonitoring.CollectorData{}, teammonitoring.CollectorData{}, err
	}

	userInProject := fmt.Sprintf("%v/%v", member.UserID, project.ChannelName)
	dataOnUserInProject, err := teammonitoring.GetCollectorData(r.conf, "user-in-project", userInProject, dateFrom, dateTo)
	if err != nil {
		return teammonitoring.CollectorData{}, teammonitoring.CollectorData{}, err
	}

	return dataOnUser, dataOnUserInProject, err
}

func (r *Reporter) PrepareAttachment(user model.ChannelMember, dataOnUser, dataOnUserInProject teammonitoring.CollectorData, isNonReporter bool, collectorError error) (string, int) {
	var worklogs, commits, standup, worklogsEmoji, worklogsTime string
	var points int

	//configure worklogs
	w := dataOnUser.Worklogs / 3600
	switch {
	case w < 3:
		worklogsEmoji = ":angry:"
	case w >= 3 && w < 7:
		worklogsEmoji = ":disappointed:"
	case w >= 7 && w < 9:
		worklogsEmoji = ":wink:"
		points++
	case w >= 9:
		worklogsEmoji = ":sunglasses:"
		points++
	}
	worklogsTime = utils.SecondsToHuman(dataOnUser.Worklogs)
	if dataOnUser.Worklogs != dataOnUserInProject.Worklogs {
		worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(dataOnUserInProject.Worklogs), utils.SecondsToHuman(dataOnUser.Worklogs))
	}
	worklogs = fmt.Sprintf(r.conf.Translate.Worklogs, worklogsTime, worklogsEmoji)

	//configure commits
	if dataOnUserInProject.TotalCommits == 0 {
		commits = fmt.Sprintf(r.conf.Translate.NoCommits, dataOnUserInProject.TotalCommits)
	} else {
		commits = fmt.Sprintf(r.conf.Translate.HasCommits, dataOnUserInProject.TotalCommits)
		points++
	}

	//configure standup
	if isNonReporter == true {
		standup = r.conf.Translate.NoStandup
	} else {
		standup = r.conf.Translate.HasStandup
		points++
	}

	// prepare field value and assign points
	fieldValue := fmt.Sprintf("%-16v|%-12v|%-10v|\n", worklogs, commits, standup)

	if user.RoleInChannel == "pm" || user.RoleInChannel == "designer" {
		fieldValue = fmt.Sprintf("%-16v|%-10v|\n", worklogs, standup)
		points++
	}

	if r.conf.TeamMonitoringEnabled == false || collectorError != nil {
		fieldValue = fmt.Sprintf("%-10v\n", standup)
	}

	return fieldValue, points
}

func (r *Reporter) PrepareWeeklyAttachment(user model.ChannelMember, dataOnUser, dataOnUserInProject teammonitoring.CollectorData, collectorError error) (string, int) {
	var worklogs, commits, worklogsEmoji, worklogsTime string
	var points int

	//configure worklogs
	w := dataOnUser.Worklogs / 3600
	switch {
	case w < 31:
		worklogsEmoji = ":disappointed:"
	case w >= 31 && w < 35:
		worklogsEmoji = ":wink:"
		points++
	case w >= 35:
		worklogsEmoji = ":sunglasses:"
		points++
	}
	worklogsTime = utils.SecondsToHuman(dataOnUser.Worklogs)
	if dataOnUser.Worklogs != dataOnUserInProject.Worklogs {
		worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(dataOnUserInProject.Worklogs), utils.SecondsToHuman(dataOnUser.Worklogs))
	}

	worklogs = fmt.Sprintf(r.conf.Translate.Worklogs, worklogsTime, worklogsEmoji)

	//configure commits
	if dataOnUserInProject.TotalCommits == 0 {
		commits = fmt.Sprintf(r.conf.Translate.NoCommits, dataOnUserInProject.TotalCommits)
	} else {
		commits = fmt.Sprintf(r.conf.Translate.HasCommits, dataOnUserInProject.TotalCommits)
		points++
	}

	// prepare field value and assign points
	fieldValue := fmt.Sprintf("%-16v|%-12v|\n", worklogs, commits)

	if user.RoleInChannel == "pm" || user.RoleInChannel == "designer" {
		fieldValue = fmt.Sprintf("%-16v|\n", worklogs)
		points++
	}

	if r.conf.TeamMonitoringEnabled == false || collectorError != nil {
		fieldValue = ""
	}

	return fieldValue, points
}

func (r *Reporter) GenerateAttachment(fieldValue string, points int) slack.Attachment {
	var attachment slack.Attachment
	var attachmentFields []slack.AttachmentField

	//if there is nothing to show, do not create attachment
	if fieldValue != "" {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Value: fieldValue,
			Short: false,
		})
	}

	attachment.Text = ""
	switch p := points; p {
	case 0:
		attachment.Color = "danger"
	case 1, 2:
		attachment.Color = "warning"
	case 3:
		attachment.Color = "good"
	}

	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		attachment.Color = "good"
	}

	attachment.Fields = attachmentFields
	return attachment
}

// StandupReportByProject creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProject(channel model.Channel, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.conf.Translate.ReportOnProjectHead, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("SetupDays failed: %v", err)
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		chanMembers, err := r.db.ListChannelMembers(channel.ChannelID)
		if err != nil || len(chanMembers) == 0 {
			continue
		}
		dayInfo := ""
		for _, member := range chanMembers {
			if !r.db.MemberShouldBeTracked(member.ID, dateFrom) {
				logrus.Infof("member should not be tracked: %v", member.UserID)
				continue
			}
			userIsNonReporter, err := r.db.IsNonReporter(member.UserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProject IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.conf.Translate.UserDidNotStandup, member.UserID)
			} else {
				standup, err := r.db.SelectStandupsFiltered(member.UserID, channel.ChannelID, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
					continue
				}
				dayInfo += fmt.Sprintf(r.conf.Translate.UserDidStandup, member.UserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.conf.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}

	}
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.conf.Translate.ReportOnUserHead, slackUserID, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		channels, err := r.db.GetUserChannels(slackUserID)
		if err != nil || len(channels) == 0 {
			continue
		}
		dayInfo := ""
		for _, channel := range channels {
			channelName, err := r.db.GetChannelName(channel)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser GetChannelName failed: %v", err)
				continue
			}
			member, err := r.db.FindChannelMemberByUserID(slackUserID, channel)
			if err != nil {
				logrus.Infof("FindChannelMemberByUserID failed: %v", err)
				continue
			}
			if !r.db.MemberShouldBeTracked(member.ID, dateFrom) {
				logrus.Infof("member should not be tracked: %v", slackUserID)
				continue
			}
			userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.conf.Translate.UserDidNotStandupInChannel, channelName, slackUserID)
			} else {
				standup, err := r.db.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting.go reportByUser SelectStandupsFiltered failed: %v", err)
				}
				dayInfo += fmt.Sprintf(r.conf.Translate.UserDidStandupInChannel, channelName, slackUserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
			dayInfo += "================================================\n"
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.conf.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channel model.Channel, slackUserID string, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.conf.Translate.ReportOnProjectAndUserHead, slackUserID, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := utils.SetupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dayInfo := ""
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		logrus.Infof("reportByProjectAndUser: dateFrom: '%v', dateTo: '%v'", dateFrom, dateTo)
		member, err := r.db.FindChannelMemberByUserID(slackUserID, channel.ChannelID)
		if err != nil {
			logrus.Infof("FindChannelMemberByUserID failed: %v", err)
			continue
		}
		if !r.db.MemberShouldBeTracked(member.ID, dateFrom) {
			logrus.Infof("member should not be tracked: %v", slackUserID)
			continue
		}
		userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel.ChannelID, dateFrom, dateTo)
		if err != nil {
			logrus.Errorf("reporting.go reportByProjectAndUser IsNonReporter failed: %v", err)
			continue
		}
		if userIsNonReporter {
			dayInfo += fmt.Sprintf(r.conf.Translate.UserDidNotStandup, slackUserID)
			dayInfo += "\n"
		} else {
			standup, err := r.db.SelectStandupsFiltered(slackUserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProjectAndUser SelectStandupsFiltered failed: %v", err)
				continue
			}
			dayInfo += fmt.Sprintf(r.conf.Translate.UserDidStandup, slackUserID)
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.conf.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}
