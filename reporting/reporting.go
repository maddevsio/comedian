package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/sirupsen/logrus"
)

//Reporter provides db and translation to functions
type (
	Reporter struct {
		db     storage.Storage
		Config config.Config
	}

	// CollectorData used to parse data on user from Collector
	CollectorData struct {
		TotalCommits int `json:"total_commits"`
		TotalMerges  int `json:"total_merges"`
		Worklogs     int `json:"worklogs"`
	}
)

//NewReporter creates new reporter instanse
func NewReporter(c config.Config) (*Reporter, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	r := &Reporter{db: conn, Config: c}
	return r, nil
}

// StandupReportByProject creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProject(channelID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channelName, err := r.db.GetChannelName(channelID)
	if err != nil {
		logrus.Errorf("reporting:GetChannelName failed: %v", err)
		return "", err
	}
	reportHead := fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))

	reportBody := ""
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting:setupDays failed: %v", err)
		return "", err
	}

	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		standupers, err := r.db.ListChannelMembers(channelID)
		logrus.Infof("reporting: standupers: %v, err: %v", standupers, err)
		if err != nil || len(standupers) == 0 {
			reportBody += r.Config.Translate.ReportNoData
			continue
		}
		dayInfo := ""
		for _, user := range standupers {
			userIsNonReporter, err := r.db.IsNonReporter(user.UserID, channelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting:IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, user.UserID)
				continue
			}
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandup, user.UserID)
			standups, err := r.db.SelectStandupsFiltered(user.UserID, channelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
				continue
			}
			dayInfo += fmt.Sprintf("%v \n", standups[0].Comment)
		}
		if dayInfo != "" {
			reportBody += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			reportBody += dayInfo
		}
	}
	report := reportHead + reportBody
	//reportBody += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	reportHead := fmt.Sprintf(r.Config.Translate.ReportOnUserHead, slackUserID, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	reportBody := ""

	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		// начало
		channels, err := r.db.GetUserChannels(slackUserID)
		if err != nil || len(channels) == 0 {
			reportBody += r.Config.Translate.ReportNoData
			continue
		}
		dayInfo := ""
		for _, channel := range channels {
			channelName, err := r.db.GetChannelName(channel)
			if err != nil {
				continue
			}
			userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
			if err != nil && err.Error() == "sql: no rows in result set" {
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandupInChannel, channel, channelName, slackUserID)
				continue
			}
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandupInChannel, channel, channelName, slackUserID)
			standups, err := r.db.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			dayInfo += fmt.Sprintf("%v \n", standups[0].Comment)
		}
		if dayInfo != "" {
			reportBody += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			reportBody += dayInfo
		}
	}

	if reportBody == "" {
		reportBody += r.Config.Translate.ReportNoData
	}
	report := reportHead + reportBody
	//reportBody += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channelID, slackUserID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	channelName, err := r.db.GetChannelName(channel)
	if err != nil {
		return "", err
	}
	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, slackUserID, channelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
		if err != nil {
			continue
		}
		report += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
		if userIsNonReporter {
			report += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, slackUserID)
			report += "\n"
			continue
		}
		report += fmt.Sprintf(r.Config.Translate.UserDidStandup, slackUserID)
		standups, err := r.db.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
		if err != nil {
			fmt.Println(err)
			continue
		}
		report += fmt.Sprintf("%v \n", standups[0].Comment)
	}

	//report += r.fetchCollectorData(collectorData)
	return report, nil
}

func (r *Reporter) fetchCollectorData(data []byte) string {
	var cd CollectorData
	err := json.Unmarshal(data, &cd)
	if err != nil {
		return ""
	}
	if cd.Worklogs != 0 {
		return fmt.Sprintf(r.Config.Translate.ReportCollectorDataUser, cd.TotalCommits, cd.TotalMerges, cd.Worklogs/3600)
	}
	return fmt.Sprintf(r.Config.Translate.ReportOnProjectCollectorData, cd.TotalCommits, cd.TotalMerges)
}

//setupDays gets dates and returns their differense in days
func (r *Reporter) setupDays(dateFrom, dateTo time.Time) (time.Time, int, error) {
	if dateTo.Before(dateFrom) {
		return time.Now(), 0, errors.New(r.Config.Translate.DateError1)
	}
	if dateTo.After(time.Now()) {
		return time.Now(), 0, errors.New(r.Config.Translate.DateError2)
	}
	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	numberOfDays := int(dateToRounded.Sub(dateFromRounded).Hours() / 24)
	return dateFromRounded, numberOfDays, nil
}
