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
	"github.com/sirupsen/logrus"
)

type reportContent struct {
	Channel      string
	Standups     []model.Standup
	NonReporters []model.StandupUser
}

type reportEntry struct {
	DateFrom       time.Time
	DateTo         time.Time
	ReportContents []reportContent
}

var localizer *i18n.Localizer

func initLocalizer() *i18n.Localizer {
	localizer, err := config.GetLocalizer()
	if err != nil {
		logrus.Errorf("reporting: GetLocalizer failed: %v\n", err)
		return nil
	}
	return localizer
}

// StandupReportByProject creates a standup report for a specified period of time
func StandupReportByProject(db storage.Storage, channelID string, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	channel := strings.Replace(channelID, "#", "", -1)
	reportEntries, err := getReportEntriesForPeriodByChannel(db, channel, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: getReportEntriesForPeriodByChannel failed: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("report entries: %#v\n", reportEntries)
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnProjectHead"})
	if err != nil {
		logrus.Errorf("reporting: Localize failed: %v\n", err)
	}
	report := fmt.Sprintf(text, channel)
	report += ReportEntriesForPeriodByChannelToString(reportEntries)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func StandupReportByUser(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	reportEntries, err := getReportEntriesForPeriodbyUser(db, user, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: getReportEntriesForPeriodbyUser failed: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("reporting: report entries: %#v\n", reportEntries)
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnUserHead"})
	if err != nil {
		logrus.Errorf("reporting: Localize failed: %v\n", err)
	}
	report := fmt.Sprintf(text, user.SlackName)
	report += ReportEntriesByUserToString(reportEntries)
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func StandupReportByProjectAndUser(db storage.Storage, channelID string, user model.StandupUser, dateFrom, dateTo time.Time) (string, error) {
	localizer = initLocalizer()
	channel := strings.Replace(channelID, "#", "", -1)
	reportEntries, err := getReportEntriesForPeriodByChannelAndUser(db, channel, user, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: getReportEntriesForPeriodByChannelAndUser: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("reporting: report entries: %#v\n", reportEntries)

	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportOnProjectAndUserHead"})
	if err != nil {
		logrus.Errorf("reporting: Localize failed: %v\n", err)

	}
	report := fmt.Sprintf(text, channel, user.SlackName)
	report += ReportEntriesForPeriodByChannelToString(reportEntries)
	return report, nil
}

//getReportEntriesForPeriodByChannel returns report entries by channel
func getReportEntriesForPeriodByChannel(db storage.Storage, channelID string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: setupDays failed: %v\n", err)
		return nil, err
	}
	logrus.Infof("reporting: chanReport, channel: <#%v>", channelID)

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		currentDayStandups, err := db.SelectStandupsByChannelIDForPeriod(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v", err)
			return nil, err
		}
		currentDayNonReporters, err := db.ListNonReportersByTimeAndChannelID(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: ListNonReportersByTimeAndChannelID failed: %v", err)
			return nil, err
		}
		logrus.Infof("reporting: chanReport, current day standups: %v\n", currentDayStandups)
		logrus.Infof("reporting: chanReport, current day non reporters: %v\n", currentDayNonReporters)
		if len(currentDayNonReporters) > 0 || len(currentDayStandups) > 0 {
			reportContents := make([]reportContent, 0, 1)
			reportContents = append(reportContents,
				reportContent{
					Standups:     currentDayStandups,
					NonReporters: currentDayNonReporters})

			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:       currentDateFrom,
					DateTo:         currentDateTo,
					ReportContents: reportContents})
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
			logrus.Errorf("reporting: Localize failed: %v\n", err)

		}
		return report + text
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportPeriod"})
		if err != nil {
			logrus.Errorf("reporting: Localize failed: %v\n", err)

		}
		report += fmt.Sprintf(text, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.ReportContents[0].Standups {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportStandupFromUser"})
			if err != nil {
				logrus.Errorf("reporting: Localize failed: %v\n", err)

			}
			report += fmt.Sprintf(text, standup.Username, standup.Comment)
		}
		for _, user := range value.ReportContents[0].NonReporters {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportIgnoredStandup"})
			if err != nil {
				logrus.Errorf("reporting: Localize failed: %v\n", err)

			}
			report += fmt.Sprintf(text, user.SlackName)
		}
	}

	return report
}

func getReportEntriesForPeriodbyUser(db storage.Storage, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: setupDays failed: %v\n", err)
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := db.FindStandupUsers(user.SlackName)
		if err != nil {
			logrus.Errorf("reporting: FindStandupUser failed: %v\n", err)
			return nil, err
		}
		reportContents := make([]reportContent, 0, len(standupUsers))
		for _, standupUser := range standupUsers {
			channel := standupUser.ChannelID
			logrus.Infof("reporting: userReport, user: %#v\n", standupUser)
			createdStandups, err := db.SelectStandupsForPeriod(currentDateFrom, currentDateTo)
			if err != nil {
				logrus.Errorf("select standup for period: %v\n", err)
				return nil, err
			}
			logrus.Infof("reporting: userReport, created standups: %#v\n", createdStandups)
			currentDayStandups := make([]model.Standup, 0, 1)
			currentDayNonReporter := make([]model.StandupUser, 0, 1)

			if standupUser.Created.After(currentDateTo) {
				continue
			}
			found := false
			for _, standup := range createdStandups {
				if standupUser.SlackUserID == standup.UsernameID {
					found = true
					currentDayStandups = append(currentDayStandups, standup)
					break
				}
			}
			if !found {
				currentDayNonReporter = append(currentDayNonReporter, standupUser)
			}
			logrus.Infof("reporting: userReport, current day standups: %#v\n", currentDayStandups)
			logrus.Infof("reporting: userReport, current day non reporters: %#v\n", currentDayNonReporter)

			if len(currentDayNonReporter) > 0 || len(currentDayStandups) > 0 {
				reportContents = append(reportContents,
					reportContent{
						Channel:      channel,
						Standups:     currentDayStandups,
						NonReporters: currentDayNonReporter})
			}
		}
		reportEntries = append(reportEntries,
			reportEntry{
				DateFrom:       currentDateFrom,
				DateTo:         currentDateTo,
				ReportContents: reportContents})
	}
	logrus.Infof("reporting: userReport, final report entries: %#v\n", reportEntries)
	return reportEntries, nil
}

//ReportEntriesByUserToString provides reporting entries in selected time period
func ReportEntriesByUserToString(reportEntries []reportEntry) string {
	localizer = initLocalizer()
	var report string
	emptyReport := true
	for _, reportEntry := range reportEntries {
		if len(reportEntry.ReportContents) != 0 {
			emptyReport = false
		}
	}
	if emptyReport {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportNoData"})
		if err != nil {
			logrus.Errorf("reporting: Localize failed: %v\n", err)

		}
		return report + text
	}

	for _, value := range reportEntries {
		if len(value.ReportContents) == 0 {
			continue
		}
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo

		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportPeriod"})
		if err != nil {
			logrus.Errorf("reporting: Localize failed: %v\n", err)

		}
		report += fmt.Sprintf(text, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, reportContent := range value.ReportContents {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportShowChannel"})
			if err != nil {
				logrus.Errorf("reporting: Localize failed: %v\n", err)
			}
			report += fmt.Sprintf(text, reportContent.Channel)
			for _, standup := range reportContent.Standups {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportStandupsFromProject"})
				if err != nil {
					logrus.Errorf("reporting: Localize failed: %v\n", err)
				}
				report += fmt.Sprintf(text, standup.ChannelID, standup.Comment)
			}
			for _, user := range reportContent.NonReporters {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportIgnoredStandup"})
				if err != nil {
					logrus.Errorf("reporting: Localize failed: %v\n", err)

				}
				report += fmt.Sprintf(text, user.SlackName)
			}
			report += "\n"

		}

	}
	return report
}

//getReportEntriesForPeriodByChannelAndUser returns report entries by channel
func getReportEntriesForPeriodByChannelAndUser(db storage.Storage, channelID string, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: setupDays failed: %v\n", err)
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)
		standupUser, err := db.FindStandupUserInChannel(user.SlackName, channelID)
		if err != nil {
			logrus.Errorf("reporting: FindStandupUserInChannel failed: %v\n", err)
			return nil, err
		}
		logrus.Infof("projectUserReport, user: %#v\n", standupUser)
		createdStandups, err := db.SelectStandupsByChannelIDForPeriod(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v\n", err)
			return nil, err
		}
		logrus.Infof("reporting: projectUserReport, created standups: %#v\n", createdStandups)
		currentDayStandups := make([]model.Standup, 0, 1)
		currentDayNonReporters := make([]model.StandupUser, 0, 1)

		if standupUser.Created.After(currentDateTo) {
			continue
		}
		found := false
		for _, standup := range createdStandups {
			if user.SlackUserID == standup.UsernameID {
				found = true
				currentDayStandups = append(currentDayStandups, standup)
				break
			}
		}
		if !found {
			currentDayNonReporters = append(currentDayNonReporters, standupUser)
		}
		logrus.Infof("reporting: projectUserReport, current day standups: %#v\n", currentDayStandups)
		logrus.Infof("reporting: projectUserReport, current day non reporters: %#v\n", currentDayNonReporters)
		if len(currentDayNonReporters) > 0 || len(currentDayStandups) > 0 {
			reportContents := make([]reportContent, 0, 1)
			reportContents = append(reportContents,
				reportContent{
					Standups:     currentDayStandups,
					NonReporters: currentDayNonReporters})

			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:       currentDateFrom,
					DateTo:         currentDateTo,
					ReportContents: reportContents})
		}
	}
	logrus.Infof("reporting: projectUserReport, final report entries: %#v\n", reportEntries)
	return reportEntries, nil
}

//setupDays gets dates and returns their differense in days
func setupDays(dateFrom, dateTo time.Time) (time.Time, int, error) {
	if dateTo.Before(dateFrom) {
		err := errors.New("Starting date is bigger than end date")
		logrus.Errorf("reporting: setupDays Before failed: %v\n", err)
		return time.Now(), 0, err
	}
	if dateTo.After(time.Now()) {
		err := errors.New("Report end time was in the future, time range was truncated")
		logrus.Errorf("reporting: setupDays After failed: %v\n", err)
		return time.Now(), 0, err
	}

	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	dateDiff := dateToRounded.Sub(dateFromRounded)
	numberOfDays := int(dateDiff.Hours() / 24)
	return dateFromRounded, numberOfDays, nil
}
