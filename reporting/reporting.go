package reporting

import (
	"errors"
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
	gocron.Every(1).Day().At(r.conf.ReportTime).Do(r.teamReport)
}

// teamReport generates report on users who did not fullfill their working duties
func (r *Reporter) teamReport() {
	attachments := []slack.Attachment{}
	var err error

	day := int(time.Now().Weekday())
	logrus.Infof("Generate report on day %v", day)
	switch day {
	case 1:
		attachments, err = r.generateReportForSunday()
		if err != nil {
			logrus.Errorf("Failed to generate report for Sunday! Error: %v", err)
			fail := fmt.Sprintf(":warning: Failed to generate report for Sunday! Error: %v", err)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
			return
		}
		if len(attachments) == 0 {
			r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.EmptyReportForSunday, nil)
			return
		}
		r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeaderMonday, attachments)
	case 2, 3, 4, 5, 6:
		attachments, err = r.generateReportForYesterday()
		if err != nil {
			logrus.Errorf("Failed to generate report for yesterday! Error: %v", err)
			fail := fmt.Sprintf(":warning: Failed to generate report for yesterday! Error: %v", err)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
			return
		}
		r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeader, attachments)
	case 0:
		attachments, err = r.generateReportForWeek()
		if err != nil {
			logrus.Errorf("Failed to generate weekly report! Error: %v", err)
			fail := fmt.Sprintf(":warning: Failed to generate weekly report! Error: %v", err)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
			return
		}
		r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeaderWeekly, attachments)
	}
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

func (r *Reporter) generateReportForYesterday() ([]slack.Attachment, error) {
	attachments := []slack.Attachment{}
	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, -1)
	startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

	allUsers, err := r.db.ListAllChannelMembers()
	if err != nil {
		return attachments, err
	}

	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	for _, user := range allUsers {
		var attachment slack.Attachment
		var attachmentFields []slack.AttachmentField
		var worklogs, commits, standup, worklogsEmoji, worklogsTime string
		var points int

		// seems it does not work correctly. There might be different cases
		if !r.db.MemberShouldBeTracked(user.ID, startDate) {
			logrus.Infof("Member Should not be tracked: %v", user.ID)
			continue
		}

		project, err := r.db.SelectChannel(user.ChannelID)
		if err != nil {
			logrus.Infof("Select channel failed: %v", err)
			continue
		}

		dataOnUser, collectorErrorOnUser := teammonitoring.GetCollectorData(r.conf, "users", user.UserID, dateFrom, dateTo)
		if collectorErrorOnUser != nil {
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)
		dataOnUserInProject, collectorErrorOnUserInProject := teammonitoring.GetCollectorData(r.conf, "user-in-project", userInProject, dateFrom, dateTo)
		if collectorErrorOnUserInProject != nil {
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		isNonReporter, err := r.db.IsNonReporter(user.UserID, user.ChannelID, startDateTime, endDateTime)
		if err != nil {
			logrus.Infof("User is non reporter failed: %v", err)
		}

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

		if dataOnUserInProject.TotalCommits == 0 {
			commits = fmt.Sprintf(r.conf.Translate.NoCommits, dataOnUserInProject.TotalCommits)
		} else {
			commits = fmt.Sprintf(r.conf.Translate.HasCommits, dataOnUserInProject.TotalCommits)
			points++
		}
		if isNonReporter == true {
			standup = r.conf.Translate.NoStandup
		} else {
			standup = r.conf.Translate.HasStandup
			points++
		}

		whoAndWhere := fmt.Sprintf(r.conf.Translate.IsRook, user.UserID, project.ChannelName)
		fieldValue := fmt.Sprintf("%-16v|%-12v|%-10v|\n", worklogs, commits, standup)
		if r.conf.TeamMonitoringEnabled == false || collectorErrorOnUser != nil || collectorErrorOnUserInProject != nil {
			fieldValue = fmt.Sprintf("%-10v\n", standup)
		}
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Value: fieldValue,
			Short: false,
		})

		attachment.Text = whoAndWhere
		switch p := points; p {
		case 0:
			attachment.Color = "danger"
		case 1, 2:
			attachment.Color = "warning"
		case 3:
			attachment.Color = "good"
		}
		attachment.Fields = attachmentFields
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

func (r *Reporter) generateReportForWeek() ([]slack.Attachment, error) {
	if r.conf.TeamMonitoringEnabled == false {
		return nil, errors.New(r.conf.Translate.ErrorRooksReportWeekend)
	}
	attachments := []slack.Attachment{}
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now().AddDate(0, 0, -1)

	allUsers, err := r.db.ListAllChannelMembers()
	if err != nil {
		return attachments, err
	}

	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	for _, user := range allUsers {
		var attachment slack.Attachment
		var attachmentFields []slack.AttachmentField
		var worklogs, commits, worklogsTime string

		project, err := r.db.SelectChannel(user.ChannelID)
		if err != nil {
			logrus.Infof("Select channel failed: %v", err)
			continue
		}

		dataOnUser, collectorErrorOnUser := teammonitoring.GetCollectorData(r.conf, "users", user.UserID, dateFrom, dateTo)
		if collectorErrorOnUser != nil {
			logrus.Errorf("collectorErrorOnUser failed: %v", collectorErrorOnUser)
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)
		dataOnUserInProject, collectorErrorOnUserInProject := teammonitoring.GetCollectorData(r.conf, "user-in-project", userInProject, dateFrom, dateTo)
		if collectorErrorOnUserInProject != nil {
			logrus.Errorf("collectorErrorOnUserInProject failed: %v", collectorErrorOnUserInProject)
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		worklogsTime = utils.SecondsToHuman(dataOnUser.Worklogs)
		if dataOnUser.Worklogs != dataOnUserInProject.Worklogs {
			worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(dataOnUserInProject.Worklogs), utils.SecondsToHuman(dataOnUser.Worklogs))
		}
		worklogs = fmt.Sprintf(" worklogs: %v ", worklogsTime)
		commits = fmt.Sprintf(" commits: %v ", dataOnUserInProject.TotalCommits)

		whoAndWhere := fmt.Sprintf(r.conf.Translate.IsRook, user.UserID, project.ChannelName)

		fieldValue := fmt.Sprintf("%-16v|%-12v|\n", worklogs, commits)

		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Value: fieldValue,
			Short: false,
		})

		attachment.Text = whoAndWhere
		attachment.Color = "good"
		attachment.Fields = attachmentFields
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

func (r *Reporter) generateReportForSunday() ([]slack.Attachment, error) {
	attachments := []slack.Attachment{}
	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, -1)
	startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

	allUsers, err := r.db.ListAllChannelMembers()
	if err != nil {
		return attachments, err
	}

	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	for _, user := range allUsers {
		var attachment slack.Attachment
		var attachmentFields []slack.AttachmentField
		var worklogsTime, fieldValue string

		project, err := r.db.SelectChannel(user.ChannelID)
		if err != nil {
			logrus.Infof("Select channel failed: %v", err)
			continue
		}

		dataOnUser, collectorErrorOnUser := teammonitoring.GetCollectorData(r.conf, "users", user.UserID, dateFrom, dateTo)
		if collectorErrorOnUser != nil {
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)
		dataOnUserInProject, collectorErrorOnUserInProject := teammonitoring.GetCollectorData(r.conf, "user-in-project", userInProject, dateFrom, dateTo)
		if collectorErrorOnUserInProject != nil {
			userFull, _ := r.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			r.s.SendUserMessage(r.conf.ManagerSlackUserID, fail)
		}

		isNonReporter, err := r.db.IsNonReporter(user.UserID, user.ChannelID, startDateTime, endDateTime)
		if dataOnUserInProject.Worklogs == 0 && dataOnUser.TotalCommits == 0 && err != nil {
			logrus.Infof("worklogs, commits, standup for user %v in channel %v are empty. Skip!", user.UserID, user.ChannelID)
			continue
		}

		worklogsTime = utils.SecondsToHuman(dataOnUser.Worklogs)
		if dataOnUser.Worklogs != dataOnUserInProject.Worklogs {
			worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(dataOnUserInProject.Worklogs), utils.SecondsToHuman(dataOnUser.Worklogs))
		}

		if dataOnUserInProject.Worklogs != 0 {
			fieldValue += fmt.Sprintf(" worklogs: %v ", worklogsTime)
			fieldValue += "|"
		}

		if dataOnUserInProject.TotalCommits != 0 {
			fieldValue += fmt.Sprintf(" commits: %v ", dataOnUserInProject.TotalCommits)
			fieldValue += "|"
		}

		if isNonReporter != true && err == nil {
			fieldValue += r.conf.Translate.HasStandup
			fieldValue += "|\n"
		}

		whoAndWhere := fmt.Sprintf(r.conf.Translate.IsRook, user.UserID, project.ChannelName)

		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Value: fieldValue,
			Short: false,
		})

		attachment.Text = whoAndWhere
		attachment.Color = "good"
		attachment.Fields = attachmentFields
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}
