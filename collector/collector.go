package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/chat"
)

//Data used to parse data on user from Collector
type Data struct {
	Commits  int `json:"total_commits"`
	Worklogs int `json:"worklogs"`
}

//GetCollectorData sends api request to collector servise and returns collector object
func GetCollectorData(slack chat.Slack, getDataOn, data, dateFrom, dateTo string) (Data, error) {
	var collectorData Data
	if slack.CP.CollectorEnabled == false {
		return collectorData, nil
	}
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", slack.Conf.CollectorURL, slack.TeamDomain, getDataOn, data, dateFrom, dateTo)
	logrus.Infof("teammonitoring: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		return collectorData, err
	}
	token := slack.Conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return collectorData, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		logrus.Errorf("teammonitoring: res status code - %v. Could not get data", res.StatusCode)
		return collectorData, errors.New("could not get data on this request")
	}
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &collectorData)
	return collectorData, nil
}
