package reporting

import (
	"fmt"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"time"
)

func StandupReportByProject(db storage.Storage, channelName string, dateFrom, dateTo time.Time) (string, error) {
	nonReporters, err := getNonReportersForPeriod(db, channelName, dateFrom, dateTo)
	if err != nil {
		return "", err
	}
	report := printNonReportersToString(nonReporters)
	return report, nil
}

func getNonReportersForPeriod(db storage.Storage, channelName string, dateFrom, dateTo time.Time) (map[time.Time][]model.StandupUser, error) {
	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)

	nonReporters := make(map[time.Time][]model.StandupUser)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := db.ListStandupUsersByChannelName(channelName)
		if err != nil {
			return nil, err
		}

		usersWhoCreatedStandups, err := db.SelectStandupByChannelIDForPeriod(channelName, currentDateFrom, currentDateTo)
		if err != nil {
			return nil, err
		}
		currentDayNonReporters := make([]model.StandupUser, 0, len(standupUsers))
		for _, user := range standupUsers {
			if user.Created.After(currentDateTo) {
				continue
			}
			found := false
			for _, standupCreator := range usersWhoCreatedStandups {
				if user.SlackName == standupCreator.Username {
					found = true
					break
				}
			}
			if !found {
				currentDayNonReporters = append(currentDayNonReporters, user)
			}
		}
		if len(currentDayNonReporters) > 0 {
			nonReporters[currentDateFrom] = currentDayNonReporters
		}
	}
	return nonReporters, nil
}

func printNonReportersToString(nonReporters map[time.Time][]model.StandupUser) string {
	report := "Report:\n\n"
	if len(nonReporters) == 0 {
		return report + "No one ignored standups in this period"
	}

	for key, value := range nonReporters {
		currentDateFrom := key
		currentDateTo := currentDateFrom.Add(time.Hour * 24)
		report += fmt.Sprintf("\n\n%s to %s:\n", currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, user := range value {
			report += user.SlackName + "\n"
		}
	}

	return report
}
