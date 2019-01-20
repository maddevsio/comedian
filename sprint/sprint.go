package sprint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
)

//CollectorInfo used to parse sprint data from Collector
type CollectorInfo struct {
	SprintURL   string `json:"link_to_sprint"`
	SprintStart string `json:"started"`
	SprintEnd   string `json:"end"`
	Tasks       []struct {
		Issue    string `json:"title"`
		Status   string `json:"status"`
		Assignee string `json:"assignee"`
	} `json:"issues"`
}

//GetSprintData sends api request to collector service and returns Info object
func GetSprintData(bot *bot.Bot, project string) (sprintInfo CollectorInfo, err error) {
	logrus.Info("Get sprint data from collector")
	var sprintData CollectorInfo
	if bot.CP.CollectorEnabled == false {
		return sprintData, errors.New("Collector disabled")
	}
	var collectorURL string
	collectorURL = fmt.Sprintf("%v/rest/api/v1/projects/%v/%v/sprint/detail/", bot.Conf.CollectorURL, bot.TeamDomain, project)
	req, err := http.NewRequest("GET", collectorURL, nil)
	if err != nil {
		logrus.Errorf("sprint: NewRequest failed: %v", err)
		return sprintInfo, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", bot.Conf.CollectorToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sprintInfo, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logrus.Errorf("sprint: response status code - %v. Could not get data", res.StatusCode)
		return sprintInfo, errors.New("could not get data on this request")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("sprint: ioutil.ReadAll failed: %v", err)
	}
	json.Unmarshal(body, &sprintInfo)
	return sprintInfo, err
}
