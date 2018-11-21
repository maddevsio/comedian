package reporting

import (
	"fmt"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/chat"
	"gitlab.com/team-monitoring/comedian/collector"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"gitlab.com/team-monitoring/comedian/utils"
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
			var attachment slack.Attachment
			var attachmentFields []slack.AttachmentField
			var worklogs, commits, standup string
			var worklogsPoints, commitsPoints, standupPoints int

			attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)
			dataOnUser, dataOnUserInProject, collectorError := r.GetCollectorDataOnMember(member, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))

			if collectorError == nil {
				worklogs, worklogsPoints = r.processWorklogs(dataOnUser.Worklogs, dataOnUserInProject.Worklogs)
				commits, commitsPoints = r.processCommits(dataOnUser.Commits, dataOnUserInProject.Commits)
			}

			if member.RoleInChannel == "pm" || member.RoleInChannel == "designer" {
				commits = ""
				commitsPoints++
			}

			if r.conf.CollectorEnabled == false || collectorError != nil {
				worklogs = ""
				worklogsPoints++
				commits = ""
				commitsPoints++
			}

			standup, standupPoints = r.processStandup(member)

			fieldValue := worklogs + commits + standup

			//if there is nothing to show, do not create attachment
			if fieldValue == "" {
				continue
			}

			attachmentFields = append(attachmentFields, slack.AttachmentField{
				Value: fieldValue,
				Short: false,
			})

			points := worklogsPoints + commitsPoints + standupPoints

			switch points {
			case 0:
				attachment.Color = "danger"
			case 1:
				attachment.Color = "warning"
			case 2, 3:
				attachment.Color = "good"
			}

			if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
				attachment.Color = "good"
			}

			attachment.Fields = attachmentFields

			attachments = append(attachments, attachment)
		}

		if len(attachments) == 0 {
			continue
		}

		if channel.ChannelID == "G5ZEM8X7E" || channel.ChannelID == "G6H5YVB3Q" {
			r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeader, attachments)
		}

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeader, allReports)
}

func (r *Reporter) processWorklogs(totalWorklogs, projectWorklogs int) (string, int) {
	points := 0
	worklogsEmoji := ""

	w := totalWorklogs / 3600
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
	worklogsTime := utils.SecondsToHuman(totalWorklogs)

	if totalWorklogs != projectWorklogs {
		worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(projectWorklogs), utils.SecondsToHuman(totalWorklogs))
	}

	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		worklogsEmoji = ""
	}

	worklogs := fmt.Sprintf(r.conf.Translate.Worklogs, worklogsTime, worklogsEmoji)
	return worklogs, points
}

func (r *Reporter) processWeeklyWorklogs(totalWorklogs, projectWorklogs int) (string, int) {
	points := 0
	worklogsEmoji := ""

	w := totalWorklogs / 3600
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
	worklogsTime := utils.SecondsToHuman(totalWorklogs)

	if totalWorklogs != projectWorklogs {
		worklogsTime = fmt.Sprintf(r.conf.Translate.WorklogsTime, utils.SecondsToHuman(projectWorklogs), utils.SecondsToHuman(totalWorklogs))
	}

	worklogs := fmt.Sprintf(r.conf.Translate.Worklogs, worklogsTime, worklogsEmoji)
	return worklogs, points
}

func (r *Reporter) processCommits(totalCommits, projectCommits int) (string, int) {
	points := 0
	commitsEmoji := ""

	c := projectCommits
	switch {
	case c == 0:
		commitsEmoji = ":shit:"
	case c > 0:
		commitsEmoji = ":wink:"
		points++
	}

	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		commitsEmoji = ""
	}

	commits := fmt.Sprintf(r.conf.Translate.Commits, projectCommits, commitsEmoji)
	return commits, points
}

func (r *Reporter) processStandup(member model.ChannelMember) (string, int) {
	points := 0
	standup := ""
	t := time.Now().AddDate(0, 0, -1)

	shouldBeTracked := r.db.MemberShouldBeTracked(member.ID, t)
	if !shouldBeTracked {
		points++
		return "", points
	}

	timeFrom := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	timeTo := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)

	isNonReporter, err := r.db.IsNonReporter(member.UserID, member.ChannelID, timeFrom, timeTo)

	if err != nil {
		points++
		return "", points
	}

	if isNonReporter == true {
		standup = r.conf.Translate.NoStandup
	} else {
		standup = r.conf.Translate.HasStandup
		points++
	}

	return standup, points
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
			var attachment slack.Attachment
			var attachmentFields []slack.AttachmentField
			var worklogs, commits string
			var worklogsPoints, commitsPoints int

			attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)
			dataOnUser, dataOnUserInProject, collectorError := r.GetCollectorDataOnMember(member, time.Now().AddDate(0, 0, -7), time.Now().AddDate(0, 0, -1))

			if collectorError == nil {
				worklogs, worklogsPoints = r.processWeeklyWorklogs(dataOnUser.Worklogs, dataOnUserInProject.Worklogs)
				commits, commitsPoints = r.processCommits(dataOnUser.Commits, dataOnUserInProject.Commits)
			}

			if member.RoleInChannel == "pm" || member.RoleInChannel == "designer" {
				commits = ""
				commitsPoints++
			}

			if r.conf.CollectorEnabled == false || collectorError != nil {
				worklogs = ""
				worklogsPoints++
				commits = ""
				commitsPoints++
			}

			fieldValue := worklogs + commits

			//if there is nothing to show, do not create attachment
			if fieldValue == "" {
				continue
			}

			attachmentFields = append(attachmentFields, slack.AttachmentField{
				Value: fieldValue,
				Short: false,
			})

			points := worklogsPoints + commitsPoints

			switch points {
			case 0:
				attachment.Color = "danger"
			case 1:
				attachment.Color = "warning"
			case 2:
				attachment.Color = "good"
			}

			attachment.Fields = attachmentFields

			attachments = append(attachments, attachment)
		}

		if len(attachments) == 0 {
			continue
		}

		if channel.ChannelID == "G5ZEM8X7E" || channel.ChannelID == "G6H5YVB3Q" {
			r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeader, attachments)
		}

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeader, allReports)
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

//GetCollectorDataOnMember sends API request to Collector endpoint and returns CollectorData type
func (r *Reporter) GetCollectorDataOnMember(member model.ChannelMember, startDate, endDate time.Time) (collector.CollectorData, collector.CollectorData, error) {
	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	project, err := r.db.GetChannelName(member.ChannelID)
	if err != nil {
		return collector.CollectorData{}, collector.CollectorData{}, err
	}

	dataOnUser, err := collector.GetCollectorData(r.conf, "users", member.UserID, dateFrom, dateTo)
	if err != nil {
		return collector.CollectorData{}, collector.CollectorData{}, err
	}

	userInProject := fmt.Sprintf("%v/%v", member.UserID, project)
	dataOnUserInProject, err := collector.GetCollectorData(r.conf, "user-in-project", userInProject, dateFrom, dateTo)
	if err != nil {
		return collector.CollectorData{}, collector.CollectorData{}, err
	}

	return dataOnUser, dataOnUserInProject, err
}
