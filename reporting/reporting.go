package reporting

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

type reportEntry struct {
	DateFrom     time.Time
	DateTo       time.Time
	Standups     []model.Standup
	NonReporters []model.StandupUser
}

var localizer *i18n.Localizer

func initLocalizer() *i18n.Localizer {
	localizer, err := config.GetLocalizer()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return localizer
}

// StandupReportByProject creates a standup report for a specified period of time
func StandupReportByProject(db storage.Storage, channelID string, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	reportEntries, err := getReportEntriesForPeriodByChannel(db, channelID, dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return "Error!", err
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnProjectHead"})
	if err != nil {
		log.Fatal(err)
	}
	report := fmt.Sprintf(text, channelID)
	report += ReportEntriesForPeriodByChannelToString(reportEntries)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func StandupReportByUser(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	reportEntries, err := getReportEntriesForPeriodbyUser(db, user, dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return "Error!", err
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnUserHead"})
	if err != nil {
		log.Fatal(err)
	}
	report := fmt.Sprintf(text, user.SlackName)
	report += ReportEntriesByUserToString(reportEntries)
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func StandupReportByProjectAndUser(db storage.Storage, channelID string, user model.StandupUser, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	reportEntries, err := getReportEntriesForPeriodByChannelAndUser(db, channelID, user, dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return "Error!", err
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnProjectAndUserHead"})
	if err != nil {
		log.Fatal(err)
	}
	report := fmt.Sprintf(text, channelID, user.SlackName)
	report += ReportEntriesForPeriodByChannelToString(reportEntries)
	return report, nil
}

//getReportEntriesForPeriodByChannel returns report entries by channel
func getReportEntriesForPeriodByChannel(db storage.Storage, channelID string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return nil, err
	}
	channel := strings.Replace(channelID, "#", "", -1)

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := db.ListStandupUsersByChannelName(channel)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
			return nil, err
		}
		createdStandups, err := db.SelectStandupByChannelNameForPeriod(channel, currentDateFrom, currentDateTo)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
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

//ReportEntriesForPeriodByChannelToString returns report entries by channel in text
func ReportEntriesForPeriodByChannelToString(reportEntries []reportEntry) string {
	localizer = initLocalizer()
	var report string
	if len(reportEntries) == 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportNoData"})
		if err != nil {
			log.Fatal(err)
		}
		return report + text
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportPeriod"})
		if err != nil {
			log.Fatal(err)
		}
		report += fmt.Sprintf(text, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.Standups {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportStandupFromUser"})
			if err != nil {
				log.Fatal(err)
			}
			report += fmt.Sprintf(text, standup.Username, standup.Comment)
		}
		for _, user := range value.NonReporters {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportIgnoredStandup"})
			if err != nil {
				log.Fatal(err)
			}
			report += fmt.Sprintf(text, user.SlackName)
		}
	}

	return report
}

func getReportEntriesForPeriodbyUser(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUser, err := db.FindStandupUser(user.SlackName)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
			return nil, err
		}
		createdStandups, err := db.SelectStandupsForPeriod(currentDateFrom, currentDateTo)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
			return nil, err
		}
		currentDayStandups := make([]model.Standup, 0, 1)
		currentDayNonReporter := make([]model.StandupUser, 0, 1)

		if standupUser.Created.After(currentDateTo) {
			continue
		}
		found := false
		for _, standup := range createdStandups {
			if standupUser.SlackName == standup.Username {
				found = true
				currentDayStandups = append(currentDayStandups, standup)
				break
			}
		}
		if !found {
			currentDayNonReporter = append(currentDayNonReporter, standupUser)
		}
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

//ReportEntriesByUserToString provides reporting entries in selected time period
func ReportEntriesByUserToString(reportEntries []reportEntry) string {
	localizer = initLocalizer()
	var report string
	if len(reportEntries) == 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportNoData"})
		if err != nil {
			log.Fatal(err)
		}
		return report + text
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo

		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportPeriod"})
		if err != nil {
			log.Fatal(err)
		}
		report += fmt.Sprintf(text, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.Standups {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportStandupsFromProject"})
			if err != nil {
				log.Fatal(err)
			}
			report += fmt.Sprintf(text, standup.ChannelID, standup.Comment)
		}
		for _, user := range value.NonReporters {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportIgnoredStandup"})
			if err != nil {
				log.Fatal(err)
			}
			report += fmt.Sprintf(text, user.SlackName)
		}
	}
	return report
}

//getReportEntriesForPeriodByChannelAndUser returns report entries by channel
func getReportEntriesForPeriodByChannelAndUser(db storage.Storage, channelID string, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)
		standupUser, err := db.FindStandupUserInChannel(user.SlackName, channelID)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
			return nil, err
		}
		createdStandups, err := db.SelectStandupsByChannelIDForPeriod(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
			return nil, err
		}
		currentDayStandups := make([]model.Standup, 0, 1)
		currentDayNonReporters := make([]model.StandupUser, 0, 1)

		if standupUser.Created.After(currentDateTo) {
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
			currentDayNonReporters = append(currentDayNonReporters, standupUser)
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

//setupDays gets dates and returns their differense in days
func setupDays(dateFrom, dateTo time.Time) (time.Time, int, error) {
	if dateTo.Before(dateFrom) {
		return time.Now(), 0, errors.New("starting date is bigger than end date")
	}
	if dateTo.After(time.Now()) {
		return time.Now(), 0, errors.New("Report end time was in the future, time range was truncated")
	}

	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)
	return dateFromRounded, numberOfDays, nil
}
