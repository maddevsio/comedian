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
	"github.com/sirupsen/logrus"
)

// CollectorData used to parse data on user from Collector
type CollectorData struct {
	TotalCommits int `json:"total_commits"`
	Worklogs     int `json:"worklogs"`
}

// TeamMonitoring struct is used to get data from team monitoring servise
type TeamMonitoring struct {
	Chat   chat.Chat
	db     storage.Storage
	Config config.Config
}

// NewTeamMonitoring creates a new team monitoring
func NewTeamMonitoring(c config.Config, chat chat.Chat) (*TeamMonitoring, error) {
	if !c.TeamMonitoringEnabled {
		logrus.Info("Team Monitoring is disabled!")
		return &TeamMonitoring{}, errors.New("team monitoring is disabled")
	}
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	tm := &TeamMonitoring{Chat: chat, db: conn, Config: c}
	return tm, nil
}

// Start starts all team monitoring treads
func (tm *TeamMonitoring) Start() {
	gocron.Every(1).Day().At(tm.Config.ReportTime).Do(tm.reportRooks)
}

func (tm *TeamMonitoring) reportRooks() {
	finalText, err := tm.RevealRooks()
	if err != nil || finalText == "" {
		tm.Chat.SendMessage(tm.Config.ReportingChannel, "Report is currently unavailable due to unexpected error while processing data")
	}
	tm.Chat.SendMessage(tm.Config.ReportingChannel, finalText)
}

// RevealRooks displays data about rooks in channel general
func (tm *TeamMonitoring) RevealRooks() (string, error) {
	//check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return "", errors.New("day off")
	}
	timeFrom := time.Now().AddDate(0, 0, -1)
	// if today is monday, check 3 days of performance for user
	if int(time.Now().Weekday()) == 1 {
		timeFrom = time.Now().AddDate(0, 0, -3)
	}
	logrus.Infof("Time from: %v", timeFrom)
	allUsers, err := tm.db.ListAllChannelMembers()
	if err != nil {
		logrus.Errorf("team monitoring: tm.GetCurrentDayNonReporters failed: %v\n", err)
		return "", err
	}
	dateFrom := fmt.Sprintf("%d-%02d-%02d", timeFrom.Year(), timeFrom.Month(), timeFrom.Day())
	finalText := ""

	for _, user := range allUsers {
		// need to first identify if user should be tracked for this period
		if !tm.db.MemberShouldBeTracked(user.ID, timeFrom, time.Now()) {
			logrus.Infof("Member %v should not be tracked! Skipping", user.UserID)
			continue
		}

		project, err := tm.db.SelectChannel(user.ChannelID)
		if err != nil {
			logrus.Errorf("SelectChannel failed: %v", err)
			continue
		}
		// make request for info about user in project from collector to get commits and worklogs
		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)
		data, err := GetCollectorData(tm.Config, "user-in-project", userInProject, dateFrom, dateFrom)
		if err != nil {
			logrus.Errorf("team monitoring: getCollectorData failed: %v\n", err)
			continue
		}
		// need to identify if user submitted standup for this period
		isNonReporter, err := tm.db.IsNonReporter(user.UserID, user.ChannelID, timeFrom, time.Now())
		if err != nil {
			logrus.Errorf("team monitoring: IsNonReporter failed: %v\n", err)
			continue
		}

		// if logged less then 8 hours or did not commit or did not submit standup = include user in report
		if (data.Worklogs/3600 < 8) || (data.TotalCommits == 0) || (isNonReporter == true) {
			fails := ""
			if data.Worklogs/3600 < 8 {
				fails += fmt.Sprintf(tm.Config.Translate.NoWorklogs, utils.SecondsToHuman(data.Worklogs))
			} else {
				fails += fmt.Sprintf(tm.Config.Translate.HasWorklogs, utils.SecondsToHuman(data.Worklogs))
			}
			if data.TotalCommits == 0 {
				fails += tm.Config.Translate.NoCommits
			} else {
				fails += fmt.Sprintf(tm.Config.Translate.HasCommits, data.TotalCommits)
			}
			if isNonReporter == true {
				fails += tm.Config.Translate.NoStandup
			} else {
				fails += tm.Config.Translate.HasStandup
			}
			text := fmt.Sprintf(tm.Config.Translate.IsRook, user.UserID, project.ChannelName, fails)
			if int(time.Now().Weekday()) == 1 {
				text = fmt.Sprintf(tm.Config.Translate.IsRookMonday, user.UserID, project.ChannelName, fails)
			}
			text += "\n\n"
			finalText += text
		}
	}
	return finalText, nil
}

//GetCollectorData sends api request to collector servise and returns collector object
func GetCollectorData(conf config.Config, getDataOn, data, dateFrom, dateTo string) (CollectorData, error) {
	var collectorData CollectorData
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", conf.CollectorURL, conf.TeamDomain, getDataOn, data, dateFrom, dateTo)
	logrus.Infof("teammonitoring: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		logrus.Errorf("teammonitoring: http.NewRequest failed: %v\n", err)
		return collectorData, err
	}
	token := conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	res, err := http.DefaultClient.Do(req)
	logrus.Infof("RESPONSE: %v", res)
	if err != nil {
		logrus.Errorf("teammonitoring: http.DefaultClient.Do(req) failed: %v\n", err)
		return collectorData, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("teammonitoring: ioutil.ReadAll(res.Body) failed: %v\n", err)
		return collectorData, err
	}

	json.Unmarshal(body, &collectorData)

	return collectorData, nil
}
