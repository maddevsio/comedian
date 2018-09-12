package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
)

//Reporter provides db and translation to functions
type (
	Reporter struct {
		DB     storage.Storage
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
	r := &Reporter{DB: conn, Config: c}
	return r, nil
}

// StandupReportByProject creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProject(channelID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channel)

	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		report += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
		standupers, err := r.DB.ListStandupUsersByChannelID(channel)
		if err != nil || len(standupers) == 0 {
			report += r.Config.Translate.ReportNoData
			continue
		}
		for _, user := range standupers {
			userIsNonReporter, err := r.DB.IsNonReporter(user.SlackUserID, channel, dateFrom, dateTo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if userIsNonReporter {
				report += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, user.SlackUserID)
				continue
			}
			report += fmt.Sprintf(r.Config.Translate.UserDidStandup, user.SlackUserID)
			standups, err := r.DB.SelectStandupsFiltered(user.SlackUserID, channel, dateFrom, dateTo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			report += fmt.Sprintf("%v \n", standups[0].Comment)
		}
		report += "\n"
	}

	report += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	report := fmt.Sprintf(r.Config.Translate.ReportOnUserHead, slackUserID)

	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		report += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
		channels, err := r.DB.GetUserChannels(slackUserID)
		if err != nil || len(channels) == 0 {
			report += r.Config.Translate.ReportNoData
			continue
		}
		for _, channel := range channels {
			userIsNonReporter, err := r.DB.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if userIsNonReporter {
				report += fmt.Sprintf(r.Config.Translate.UserDidNotStandupInChannel, channel, slackUserID)
				continue
			}
			report += fmt.Sprintf(r.Config.Translate.UserDidStandupInChannel, channel, slackUserID)
			standups, err := r.DB.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			report += fmt.Sprintf("%v \n", standups[0].Comment)
		}
		report += "\n"
	}

	report += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channelID string, user model.StandupUser, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, channel, user.SlackUserID)

	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}

	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		report += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
		userIsNonReporter, err := r.DB.IsNonReporter(user.SlackUserID, channel, dateFrom, dateTo)
		if err != nil {
			report += r.Config.Translate.ReportNoData
			continue
		}
		if userIsNonReporter {
			report += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, user.SlackUserID)
			continue
		}
		report += fmt.Sprintf(r.Config.Translate.UserDidStandup, user.SlackUserID)
		standups, err := r.DB.SelectStandupsFiltered(user.SlackUserID, channel, dateFrom, dateTo)
		if err != nil {
			fmt.Println(err)
			continue
		}
		report += fmt.Sprintf("%v \n", standups[0].Comment)
	}

	report += r.fetchCollectorData(collectorData)
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
