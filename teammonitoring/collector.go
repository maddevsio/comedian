package teammonitoring

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
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
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	tm := &TeamMonitoring{Chat: chat, db: conn, Config: c}
	return tm, nil
}

// Start starts all team monitoring treads
func (tm *TeamMonitoring) Start() {
	gocron.Every(1).Day().At(tm.Config.ReportTime).Do(tm.RevealRooks)
}

// RevealRooks displays data about rooks in channel general
func (tm *TeamMonitoring) RevealRooks() {
	//check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	timeFrom := time.Now().AddDate(0, 0, -1)
	// if today is monday, check 3 days of performance for user
	if int(time.Now().Weekday()) == 1 {
		timeFrom = time.Now().AddDate(0, 0, -3)
	}
	allUsers, err := tm.db.ListAllChannelMembers()
	if err != nil {
		logrus.Errorf("team monitoring: tm.GetCurrentDayNonReporters failed: %v\n", err)
		return
	}

	dateFrom := fmt.Sprintf("%d-%02d-%02d", timeFrom.Year(), timeFrom.Month(), timeFrom.Day())

	for _, user := range allUsers {
		project, err := tm.db.SelectChannel(user.ChannelID)
		if err != nil {
			continue
		}
		userInProject := fmt.Sprintf("%v/%v", user.UserID, project.ChannelName)

		data, err := GetCollectorData(tm.Config, "user-in-project", userInProject, dateFrom, dateFrom)
		if err != nil {
			logrus.Errorf("team monitoring: getCollectorData failed: %v\n", err)
			continue
		}
		isNonReporter, err := tm.db.IsNonReporter(user.UserID, user.ChannelID, timeFrom.AddDate(0, 0, -1), time.Now())
		if err != nil {
			logrus.Errorf("team monitoring: IsNonReporter failed: %v\n", err)
			continue
		}

		if (data.Worklogs < 8) || (data.TotalCommits == 0) || (isNonReporter == true) {
			fails := ""
			if data.Worklogs < 8 {
				fails += fmt.Sprintf(tm.Config.Translate.NoWorklogs, data.Worklogs)
			} else {
				fails += fmt.Sprintf(tm.Config.Translate.HasWorklogs, data.Worklogs)
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
			tm.Chat.SendMessage(tm.Config.ReportingChannel, text)
		}
	}
}

//GetCollectorData sends api request to collector servise and returns collector object
func GetCollectorData(conf config.Config, getDataOn, data, dateFrom, dateTo string) (CollectorData, error) {
	var collectorData CollectorData
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/", conf.CollectorURL, getDataOn, data, dateFrom, dateTo)
	logrus.Infof("rest: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		logrus.Errorf("rest: http.NewRequest failed: %v\n", err)
		return collectorData, err
	}
	token := conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("rest: http.DefaultClient.Do(req) failed: %v\n", err)
		return collectorData, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("rest: ioutil.ReadAll(res.Body) failed: %v\n", err)
		return collectorData, err
	}

	json.Unmarshal(body, &collectorData)

	return collectorData, nil
}
