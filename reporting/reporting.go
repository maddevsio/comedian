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

//AttachmentItem is needed to sort attachments
type AttachmentItem struct {
	attachment slack.Attachment
	points     int
}

// NewReporter creates a new reporter instance
func NewReporter(slack *chat.Slack) *Reporter {
	reporter := &Reporter{s: slack, db: slack.DB, conf: slack.Conf}
	return reporter
}

// Start starts all team monitoring treads
func (r *Reporter) Start() {
	gocron.Every(1).Day().At(r.conf.ReportTime).Do(r.CallDisplayYesterdayTeamReport)
	gocron.Every(1).Sunday().At(r.conf.ReportTime).Do(r.CallDisplayWeeklyTeamReport)

}

// CallDisplayYesterdayTeamReport calls displayYesterdayTeamReport
func (r *Reporter) CallDisplayYesterdayTeamReport() {
	_, err := r.displayYesterdayTeamReport()
	if err != nil {
		logrus.Error("Error in displayYesterdayTeamReport: ", err)
		r.s.SendUserMessage(r.conf.ManagerSlackUserID, fmt.Sprintf("Error sending yesterday report: %v", err))
	}
}

// CallDisplayWeeklyTeamReport calls displayWeeklyTeamReport
func (r *Reporter) CallDisplayWeeklyTeamReport() {
	_, err := r.displayWeeklyTeamReport()
	if err != nil {
		logrus.Error("Error in displayWeeklyTeamReport: ", err)
		r.s.SendUserMessage(r.conf.ManagerSlackUserID, fmt.Sprintf("Error sending weekly report: %v", err))
	}
}

// displayYesterdayTeamReport generates report on users who submit standups
func (r *Reporter) displayYesterdayTeamReport() (FinalReport string, err error) {
	var allReports []slack.Attachment

	channels, err := r.db.GetAllChannels()
	if err != nil {
		logrus.Errorf("GetAllChannels failed: %v", err)
		return FinalReport, err
	}

	for _, channel := range channels {
		var attachments []slack.Attachment
		var attachmentsPull []AttachmentItem

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

			UserInfo, err := r.db.SelectUser(member.UserID)
			if err != nil {
				logrus.Errorf("SelectUser failed for  user %v: %v", UserInfo.UserName, err)
				continue
			}

			dataOnUser, dataOnUserInProject, collectorError := r.GetCollectorDataOnMember(member, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))

			if collectorError == nil {
				worklogs, worklogsPoints = r.processWorklogs(dataOnUser.Worklogs, dataOnUserInProject.Worklogs)
				commits, commitsPoints = r.processCommits(dataOnUser.Commits, dataOnUserInProject.Commits)
			}

			if member.RoleInChannel == "pm" || member.RoleInChannel == "designer" {
				commits = ""
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

			//attachment text will be depend on worklogsPoints,commitsPoints and standupPoints
			if points >= 3 {
				attachment.Text = fmt.Sprintf(r.conf.Translate.NotTagStanduper, UserInfo.UserName, channel.ChannelName)
			} else {
				attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)
			}

			switch points {
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

			item := AttachmentItem{
				attachment: attachment,
				points:     dataOnUserInProject.Worklogs,
			}

			attachmentsPull = append(attachmentsPull, item)
		}

		if len(attachmentsPull) == 0 {
			continue
		}

		attachments = r.sortReportEntries(attachmentsPull)

		r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeader, attachments)

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeader, allReports)
	FinalReport = fmt.Sprintf(r.conf.Translate.ReportHeader, allReports)
	return FinalReport, nil
}

// displayWeeklyTeamReport generates report on users who submit standups
func (r *Reporter) displayWeeklyTeamReport() (FinalReport string, e error) {
	var allReports []slack.Attachment

	channels, err := r.db.GetAllChannels()
	if err != nil {
		logrus.Errorf("GetAllChannels failed: %v", err)
		return FinalReport, err
	}

	for _, channel := range channels {
		var attachmentsPull []AttachmentItem
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

			UserInfo, err := r.db.SelectUser(member.UserID)
			if err != nil {
				logrus.Errorf("SelectUser failed for  user %v: %v", UserInfo.UserName, err)
				continue
			}

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

			//attachment text will be depend on worklogsPoints and commitsPoints
			if points >= 2 {
				attachment.Text = fmt.Sprintf(r.conf.Translate.NotTagStanduper, UserInfo.UserName, channel.ChannelName)
			} else {
				attachment.Text = fmt.Sprintf(r.conf.Translate.IsRook, member.UserID, channel.ChannelName)
			}

			switch points {
			case 0:
				attachment.Color = "danger"
			case 1:
				attachment.Color = "warning"
			case 2:
				attachment.Color = "good"
			}

			attachment.Fields = attachmentFields

			item := AttachmentItem{
				attachment: attachment,
				points:     dataOnUserInProject.Worklogs,
			}

			attachmentsPull = append(attachmentsPull, item)
		}

		if len(attachmentsPull) == 0 {
			continue
		}

		attachments = r.sortReportEntries(attachmentsPull)

		r.s.SendMessage(channel.ChannelID, r.conf.Translate.ReportHeaderWeekly, attachments)

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	r.s.SendMessage(r.conf.ReportingChannel, r.conf.Translate.ReportHeaderWeekly, allReports)
	FinalReport = fmt.Sprintf(r.conf.Translate.ReportHeaderWeekly, allReports)
	return FinalReport, nil
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
		if projectWorklogs == 0 {
			return "", points
		}
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
		if projectCommits == 0 {
			return "", points
		}
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

func (r *Reporter) sortReportEntries(entries []AttachmentItem) []slack.Attachment {
	var attachments []slack.Attachment

	for i := 0; i < len(entries); i++ {
		if !sweep(entries, i) {
			break
		}
	}

	for _, item := range entries {
		attachments = append(attachments, item.attachment)
	}

	return attachments
}

func sweep(entries []AttachmentItem, prevPasses int) bool {
	var N = len(entries)
	var didSwap = false
	var firstIndex = 0
	var secondIndex = 1

	for secondIndex < (N - prevPasses) {

		var firstItem = entries[firstIndex]
		var secondItem = entries[secondIndex]
		if entries[firstIndex].points < entries[secondIndex].points {
			entries[firstIndex] = secondItem
			entries[secondIndex] = firstItem
			didSwap = true
		}
		firstIndex++
		secondIndex++
	}

	return didSwap
}
