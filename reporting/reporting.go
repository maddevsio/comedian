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
	report += printNonReportersToString(reportEntries)
	return report, nil
}

func getReportEntriesForPeriod(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	return nil, nil
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
		return report + "No one ignored standups in this period"
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
