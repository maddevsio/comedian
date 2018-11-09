package teammonitoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/maddevsio/comedian/config"
	"github.com/sirupsen/logrus"
)

// CollectorData used to parse data on user from Collector
type CollectorData struct {
	TotalCommits int `json:"total_commits"`
	Worklogs     int `json:"worklogs"`
}

//GetCollectorData sends api request to collector servise and returns collector object
func GetCollectorData(conf config.Config, getDataOn, data, dateFrom, dateTo string) (CollectorData, error) {
	var collectorData CollectorData
	if conf.TeamMonitoringEnabled == false {
		return collectorData, nil
	}
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
		logrus.Errorf("teammonitoring: res status code - %v. Could not get data", res.StatusCode)
		return collectorData, errors.New("could not get data on this request")
	}
	body, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &collectorData)
	return collectorData, nil
}
