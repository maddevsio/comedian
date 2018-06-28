package reporting

import (
	"errors"
	"fmt"
	"strings"
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
	report := fmt.Sprintf("Full Standup Report %v:\n\n", channelName)
	//log.Println("REPORT ENTRIES!!!", reportEntries)
	report += printNonReportersToString(reportEntries)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func StandupReportByUser(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) (string, error) {
	log.Infof("Making standup report for channel: %q, period: %s - %s",
		user, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	reportEntries, err := getReportEntriesForPeriod(db, user, dateFrom, dateTo)
	if err != nil {
		return "Error!!!", err
	}
	report := fmt.Sprintf("Full Standup Report for user <@%s>:\n\n", user.SlackName)
	//log.Println("REPORT ENTRIES!!!", reportEntries)
	report += ReportEntriesToString(reportEntries)
	return report, nil
}

func getReportEntriesForPeriod(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUser, err := db.FindStandupUser(user.SlackName)
		if err != nil {
			return nil, err
		}
		log.Println("Standup User", standupUser)
		createdStandups, err := db.SelectStandupsForPeriod(currentDateFrom, currentDateTo)
		if err != nil {
			return nil, err
		}
		log.Println("Created Standups", createdStandups)
		currentDayStandups := make([]model.Standup, 0, 1)
		currentDayNonReporter := make([]model.StandupUser, 0, 1)

		if standupUser.Created.After(currentDateTo) {
			log.Println("User is created after current date!")
			continue
		}
		log.Println("Entered standup users loop!")
		found := false
		for _, standup := range createdStandups {
			log.Println("User Slack Name: ", standupUser.SlackName)
			log.Println("Standup UserName: ", standup.Username)
			if standupUser.SlackName == standup.Username {
				found = true
				currentDayStandups = append(currentDayStandups, standup)
				break
			}
		}
		if !found {
			currentDayNonReporter = append(currentDayNonReporter, standupUser)
		}

		log.Println("Current Day Standups", currentDayStandups)
		log.Println("Current Day NON reporter", currentDayNonReporter)
		if len(currentDayNonReporter) > 0 || len(currentDayStandups) > 0 {
			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:     currentDateFrom,
					DateTo:       currentDateTo,
					Standups:     currentDayStandups,
					NonReporters: currentDayNonReporter})
		}
	}
	return reportEntries, nil
}

//ReportEntriesToString provides reporting entries in selected time period
func ReportEntriesToString(reportEntries []reportEntry) string {
	var report string
	if len(reportEntries) == 0 {
		return report + "No data for this period"
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		report += fmt.Sprintf("\n\nReport from %s to %s:\n", currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.Standups {
			report += fmt.Sprintf("\nOn project: <#%s>\n%s\n", standup.ChannelID, standup.Comment)
		}
		for _, user := range value.NonReporters {
			report += fmt.Sprintf("\n<@%s>: ignored standup\n", user.SlackName)
		}
	}

	return report
}

func getNonReportersForPeriod(db storage.Storage, channelName string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	channel := strings.Replace(channelName, "#", "", -1)
	log.Println("Number of days", numberOfDays)
	log.Println("ChannelName:", channel)

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := db.ListStandupUsersByChannelName(channel)
		if err != nil {
			return nil, err
		}
		log.Println("Standup Users", standupUsers)
		createdStandups, err := db.SelectStandupByChannelNameForPeriod(channel, currentDateFrom, currentDateTo)
		if err != nil {
			return nil, err
		}
		log.Println("Created Standups", createdStandups)
		currentDayStandups := make([]model.Standup, 0, len(standupUsers))
		currentDayNonReporters := make([]model.StandupUser, 0, len(standupUsers))
		for _, user := range standupUsers {
			if user.Created.After(currentDateTo) {
				log.Println("User is created after current date to!!!")
				continue
			}
			log.Println("Entered standup users loop!")
			found := false
			for _, standup := range createdStandups {
				log.Println("User Slack Name: ", user.SlackName)
				log.Println("Standup UserName: ", standup.Username)
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
		log.Println("Current Day Standups", currentDayStandups)
		log.Println("Current Day NON reporters", currentDayNonReporters)
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
	var report string
	if len(reportEntries) == 0 {
		return report + "No data for this period"
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		report += fmt.Sprintf("\n\nReport from %s to %s:\n", currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.Standups {
			report += fmt.Sprintf("\nStandup from <@%s>:\n%s\n", standup.Username, standup.Comment)
		}
		for _, user := range value.NonReporters {
			report += fmt.Sprintf("\n<@%s>: ignored standup\n", user.SlackName)
		}
	}

	return report
}

func setupDays(dateFrom, dateTo time.Time) (time.Time, int, error) {
	if dateTo.Before(dateFrom) {
		return time.Now(), 0, errors.New("starting date is bigger than end date")
	}
	if dateFrom.After(time.Now()) {
		return time.Now(), 0, errors.New("starting date can't be in the future")
	}
	if dateTo.After(time.Now()) {
		log.Info("Report end time was in the future, time range was truncated")
		dateTo = time.Now()
	}

	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)
	return dateFromRounded, numberOfDays, nil
}
