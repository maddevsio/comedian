package teammonitoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// CollectorData used to parse data on user from Collector
type CollectorData struct {
	TotalCommits int `json:"total_commits"`
	Worklogs     int `json:"worklogs"`
}

// TeamMonitoring struct is used to get data from team monitoring servise
type TeamMonitoring struct {
	s    *chat.Slack
	db   storage.Storage
	conf config.Config
}

// NewTeamMonitoring creates a new team monitoring
func NewTeamMonitoring(slack *chat.Slack) (*TeamMonitoring, error) {
	if !slack.Conf.TeamMonitoringEnabled {
		logrus.Info("Team Monitoring is disabled!")
		return &TeamMonitoring{}, errors.New("team monitoring is disabled")
	}
	tm := &TeamMonitoring{s: slack, db: slack.DB, conf: slack.Conf}
	return tm, nil
}

// Start starts all team monitoring treads
func (tm *TeamMonitoring) Start() {
	gocron.Every(1).Day().At(tm.conf.ReportTime).Do(tm.reportRooks)
}

func (tm *TeamMonitoring) reportRooks() {
	attachments, err := tm.RevealRooks()
	if err != nil {
		tm.s.SendMessage(tm.conf.ReportingChannel, err.Error(), nil)
	}
	if len(attachments) == 0 {
		logrus.Info("Empty Report")
		return
	}
	if int(time.Now().Weekday()) == 1 {
		tm.s.SendMessage(tm.conf.ReportingChannel, tm.conf.Translate.ReportHeaderMonday, attachments)
		return
	}
	tm.s.SendMessage(tm.conf.ReportingChannel, tm.conf.Translate.ReportHeader, attachments)
}

// RevealRooks displays data about rooks in channel general
func (tm *TeamMonitoring) RevealRooks() ([]slack.Attachment, error) {
	attachments := []slack.Attachment{}
	//check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 0 {
		return attachments, errors.New(tm.conf.Translate.ErrorRooksReportWeekend)
	}

	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, -1)
	// if today is monday, check 3 days of performance for user
	if int(time.Now().Weekday()) == 1 {
		startDate = time.Now().AddDate(0, 0, -3)
	}

	startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

	allUsers, err := tm.db.ListAllChannelMembers()

	if err != nil {
		logrus.Errorf("team monitoring: tm.GetCurrentDayNonReporters failed: %v\n", err)
		return attachments, err
	}
	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	for _, user := range allUsers {
		var attachment slack.Attachment
		var attachmentFields []slack.AttachmentField
		// need to first identify if user should be tracked for this period
		if !tm.db.MemberShouldBeTracked(user.ID, startDate) {
			logrus.Infof("Member %v should not be tracked! Skipping", user.UserID)
			continue
		}
		project, err := tm.db.SelectChannel(user.ChannelID)
		if err != nil {
			logrus.Errorf("SelectChannel failed: %v", err)
			continue
		}
		dataOnUser, err := GetCollectorData(tm.conf, "users", user.UserID, dateFrom, dateTo)
		if err != nil {
			logrus.Errorf("team monitoring: getCollectorData on users failed: %v\n", err)
			userFull, _ := tm.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning::warning::warning: GetCollectorData on Users failed on user %v|%v in %v! Please, add this user to Collector service :bangbang:", userFull.UserName, userFull.UserID, project.ChannelName)
			tm.s.SendUserMessage(tm.conf.ManagerSlackUserID, fail)
			continue
		}
		// make request for info about user in project from collector to get commits and worklogs
		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)
		dataOnUserInProject, err := GetCollectorData(tm.conf, "user-in-project", userInProject, dateFrom, dateTo)
		if err != nil {
			logrus.Errorf("team monitoring: getCollectorData failed: %v\n", err)
			userFull, _ := tm.db.SelectUser(user.UserID)
			fail := fmt.Sprintf(":warning: Failed to get data on %v|%v in %v! Check Collector servise!", userFull.UserName, userFull.UserID, project.ChannelName)
			tm.s.SendUserMessage(tm.conf.ManagerSlackUserID, fail)
			continue
		}
		// need to identify if user submitted standup for this period
		isNonReporter, err := tm.db.IsNonReporter(user.UserID, user.ChannelID, startDateTime, endDateTime)
		if err != nil {
			logrus.Errorf("team monitoring: IsNonReporter failed: %v\n", err)
			continue
		}

		var worklogs, commits, standup string
		var worklogsEmoji, worklogsTime string
		var points int

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
			worklogsTime = fmt.Sprintf(tm.conf.Translate.WorklogsTime, utils.SecondsToHuman(dataOnUserInProject.Worklogs), utils.SecondsToHuman(dataOnUser.Worklogs))
		}
		worklogs = fmt.Sprintf(tm.conf.Translate.Worklogs, worklogsTime, worklogsEmoji)

		if dataOnUserInProject.TotalCommits == 0 {
			commits = fmt.Sprintf(tm.conf.Translate.NoCommits, dataOnUserInProject.TotalCommits)
		} else {
			commits = fmt.Sprintf(tm.conf.Translate.HasCommits, dataOnUserInProject.TotalCommits)
			points++
		}
		if isNonReporter == true {
			standup = tm.conf.Translate.NoStandup
		} else {
			standup = tm.conf.Translate.HasStandup
			points++
		}
		whoAndWhere := fmt.Sprintf(tm.conf.Translate.IsRook, user.UserID, project.ChannelName)
		fieldValue := fmt.Sprintf("%-16v|%-12v|%-10v|\n", worklogs, commits, standup)
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Value: fieldValue,
			Short: false,
		})

		logrus.Infof("POINTS: %v", points)
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

//GetCollectorData sends api request to collector servise and returns collector object
func GetCollectorData(conf config.Config, getDataOn, data, dateFrom, dateTo string) (CollectorData, error) {
	var collectorData CollectorData
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", conf.CollectorURL, conf.TeamDomain, getDataOn, data, dateFrom, dateTo)
	logrus.Infof("teammonitoring: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		return collectorData, err
	}
	token := conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return collectorData, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return collectorData, errors.New("could not get data on this request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return collectorData, err
	}
	json.Unmarshal(body, &collectorData)

	return collectorData, nil
}
