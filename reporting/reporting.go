package reporting

import (
	"errors"
	"fmt"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/sirupsen/logrus"
)

//Reporter provides db and translation to functions
type Reporter struct {
	db     storage.Storage
	Config config.Config
}

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
func (r *Reporter) StandupReportByProject(channel model.Channel, dateFrom, dateTo time.Time) (string, error) {
	reportHead := fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting:setupDays failed: %v", err)
		return "", err
	}

	reportBody := ""
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		chanMembers, err := r.db.ListChannelMembers(channel.ChannelID)
		if err != nil || len(chanMembers) == 0 {
			continue
		}
		dayInfo := ""
		for _, member := range chanMembers {
			userIsNonReporter, err := r.db.IsNonReporter(member.UserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting:IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, member.UserID)
				continue
			}
			standup, err := r.db.SelectStandupsFiltered(member.UserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
				continue
			}
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandup, member.UserID)
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
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
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time) (string, error) {
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
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandupInChannel, channelName, slackUserID)
				continue
			}
			standup, err := r.db.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				continue
			}
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandupInChannel, channelName, slackUserID)
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
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
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channel model.Channel, slackUserID string, dateFrom, dateTo time.Time) (string, error) {
	reportHead := fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, slackUserID, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return "", err
	}
	reportBody := ""
	for day := 0; day <= numberOfDays; day++ {
		dayInfo := ""
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel.ChannelID, dateFrom, dateTo)
		if err != nil {
			continue
		}
		if userIsNonReporter {
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, slackUserID)
			dayInfo += "\n"
			continue
		}
		standup, err := r.db.SelectStandupsFiltered(slackUserID, channel.ChannelID, dateFrom, dateTo)
		if err != nil {
			continue
		}
		dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandup, slackUserID)
		dayInfo += fmt.Sprintf("%v \n", standup.Comment)
		if dayInfo != "" {
			reportBody += fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			reportBody += dayInfo
		}
	}

	if reportBody == "" {
		reportBody += r.Config.Translate.ReportNoData
	}
	report := reportHead + reportBody
	return report, nil
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
