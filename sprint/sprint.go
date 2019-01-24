package sprint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
)

//Message used to parse response from Collector in cases of error
type Message struct {
	Message string `json:"message"`
}

//CollectorInfo used to parse sprint data from Collector
type CollectorInfo struct {
	SprintName  string `json:"sprint_name"`
	SprintURL   string `json:"link_to_sprint"`
	SprintStart string `json:"started"`
	SprintEnd   string `json:"end"`
	Tasks       []struct {
		Issue            string `json:"title"`
		Status           string `json:"status"`
		Assignee         string `json:"assignee"`
		AssigneeFullName string `json:"assignee_name"`
		Link             string `json:"link"`
	} `json:"issues"`
}

//GetSprintData sends api request to collector service and returns Info object
func GetSprintData(bot *bot.Bot, project string) (sprintInfo CollectorInfo, err error) {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.CP.Language)
	logrus.Infof("Get sprint data from collector. Project: %v", project)
	var sprintData CollectorInfo
	if bot.CP.CollectorEnabled == false {
		return sprintData, errors.New("Collector disabled")
	}
	var collectorURL string
	collectorURL = fmt.Sprintf("%v/rest/api/v1/projects/%v/%v/sprint/detail/", bot.Conf.CollectorURL, bot.TeamDomain, project)
	logrus.Info("collectorURL: ", collectorURL)
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
		if res.StatusCode == 404 {
			var message Message
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				logrus.Errorf("sprint: ioutil.ReadAll failed: %v", err)
			}
			err = json.Unmarshal(body, &message)
			if message.Message == "Project does not have the active sprints" {
				notActiveSprint := localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:          "NotActiveSprint",
						Description: "Displays message if project doesn't has active sprint",
						Other:       "Project has no active sprint yet",
					},
				})
				bot.SendMessage(bot.CP.SprintReportChannel, notActiveSprint, nil)
				return sprintInfo, errors.New("could not get data on this request")
			}
			if err != nil {
				logrus.Errorf("sprint_reporting: json.Unmarshal failed: %v", err)
				return sprintInfo, errors.New("Error unmarshaling json")
			}
		}
		logrus.Errorf("sprint: response status code - %v. Could not get data. project: %v", res.StatusCode, project)
		return sprintInfo, errors.New("could not get data on this request")
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("sprint: ioutil.ReadAll failed: %v", err)
	}
	json.Unmarshal(body, &sprintInfo)
	return sprintInfo, err
}
