package reporting

import (
	"fmt"
	"github.com/maddevsio/comedian/storage"
	"strings"
	"time"
)

func ManagerReportByProject(db storage.Storage, channelName string, dateFrom, dateTo time.Time) (string, error) {
	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)

	report := "Report:"

	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		report += fmt.Sprintf("\n\n%s to %s:\n", currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))

		standupUsersRaw, err := db.ListStandupUsersByChannelName(channelName)
		if err != nil {
			return "", nil
		}
		var standupUsersList []string
		for _, standupUser := range standupUsersRaw {
			user := standupUser.SlackName
			standupUsersList = append(standupUsersList, user)
		}

		userStandupRaw, err := db.SelectStandupByChannelIDForPeriod(channelName, currentDateFrom, currentDateTo)
		if err != nil {
			return "", nil
		}
		var usersWhoCreatedStandup []string
		for _, userStandup := range userStandupRaw {
			user := userStandup.Username
			usersWhoCreatedStandup = append(usersWhoCreatedStandup, user)
		}
		var nonReporters []string
		for _, user := range standupUsersList {
			found := false
			for _, standupCreator := range usersWhoCreatedStandup {
				if user == standupCreator {
					found = true
					break
				}
			}
			if !found {
				nonReporters = append(nonReporters, user)
			}
		}
		nonReportersCheck := len(nonReporters)
		if nonReportersCheck == 0 {
			report += "\tNO SLACKERS"
		} else {
			report += "\t" + strings.Join(nonReporters, "\n\t")
		}
	}
	return report, nil
}
