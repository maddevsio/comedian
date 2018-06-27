package reporting

import (
	"errors"
	"fmt"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
)

type reportEntry struct {
	DateFrom     time.Time
	DateTo       time.Time
	Standups     []model.Standup ``
	NonReporters []model.StandupUser
}

// StandupReportByProject creates a standup report for a specified period of time
func StandupReportByProject(db storage.Storage, channelName string, dateFrom, dateTo time.Time) (string, error) {
	log.Infof("Making standup report for channel: %q, period: %s - %s",
		channelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	reportEntries, err := getNonReportersForPeriod(db, channelName, dateFrom, dateTo)
	if err != nil {
		return "Error!!!", err
	}
	report := printNonReportersToString(reportEntries)
	return report, nil
}

func getNonReportersForPeriod(db storage.Storage, channelName string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	if dateTo.Before(dateFrom) {
		return nil, errors.New("starting date is bigger than end date")
	}
	if dateFrom.After(time.Now()) {
		return nil, errors.New("starting date can't be in the future")
	}
	if dateTo.After(time.Now()) {
		log.Info("Report end time was in the future, time range was truncated")
		dateTo = time.Now()
	}

	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := db.ListStandupUsersByChannelName(channelName)
		if err != nil {
			return nil, err
		}

		createdStandups, err := db.SelectStandupByChannelNameForPeriod(channelName, currentDateFrom, currentDateTo)
		if err != nil {
			return nil, err
		}
		currentDayStandups := make([]model.Standup, 0, len(standupUsers))
		currentDayNonReporters := make([]model.StandupUser, 0, len(standupUsers))
		for _, user := range standupUsers {
			if user.Created.After(currentDateTo) {
				continue
			}
			found := false
			for _, standup := range createdStandups {
				if user.SlackName == standup.Username {
					found = true
					currentDayStandups = append(currentDayStandups, standup)
					break
				}
			}
			if !found {
				currentDayNonReporters = append(currentDayNonReporters, user)
			}
		}
		if len(currentDayNonReporters) > 0 || len(currentDayStandups) > 0 {
			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:     currentDateFrom,
					DateTo:       currentDateTo,
					Standups:     currentDayStandups,
					NonReporters: currentDayNonReporters})
		}
	}
	return reportEntries, nil
}

func printNonReportersToString(reportEntries []reportEntry) string {
	report := "Report:\n\n"
	if len(reportEntries) == 0 {
		return report + "No one ignored standups in this period"
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		report += fmt.Sprintf("\n\n%s to %s:\n", currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.Standups {
			report += fmt.Sprintf("\n%s:\n%s\n", standup.Username, standup.Comment)
		}
		for _, user := range value.NonReporters {
			report += fmt.Sprintf("\n%s:\nIGNORED\n", user.SlackName)
		}
	}

	return report
}
