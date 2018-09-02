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

	reportContent struct {
		Channel      string
		Standups     []model.Standup
		NonReporters []model.StandupUser
	}

	reportEntry struct {
		DateFrom       time.Time
		DateTo         time.Time
		ReportContents []reportContent
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
	reportEntries, err := r.getReportsOnChannel(channel, dateFrom, dateTo)
	if err != nil {
		// make this translatable!
		return "Could not create standup report by project", err
	}
	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channel)
	report += r.parseReporstOnChannelToString(reportEntries)
	report += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(user model.StandupUser, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	reportEntries, err := r.getReportsOnUser(user, dateFrom, dateTo)
	if err != nil {
		// make this translatable!
		return "Could not create standup report by user", err
	}
	report := fmt.Sprintf(r.Config.Translate.ReportOnUserHead, user.SlackName)
	report += r.parseReporstOnUserToString(reportEntries)
	report += r.fetchCollectorData(collectorData)
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channelID string, user model.StandupUser, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	reportEntries, err := r.getReportsOnChannelAndUser(channel, user, dateFrom, dateTo)
	if err != nil {
		return "Could not create standup report by project and user", err
	}
	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, channel, user.SlackName)
	report += r.parseReporstOnChannelToString(reportEntries)
	report += r.fetchCollectorData(collectorData)
	return report, nil
}

//getReportsOnChannel returns report entries by channel
func (r *Reporter) getReportsOnChannel(channelID string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)

		currentDayStandups, err := r.DB.SelectStandupsByChannelIDForPeriod(channelID, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}
		currentDayNonReporters, err := r.DB.GetNonReporters(channelID, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}
		if len(currentDayNonReporters) > 0 || len(currentDayStandups) > 0 {
			reportContents := make([]reportContent, 0, 1)
			reportContents = append(reportContents,
				reportContent{
					Standups:     currentDayStandups,
					NonReporters: currentDayNonReporters})

			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:       dateFrom,
					DateTo:         dateTo,
					ReportContents: reportContents})
		}
	}
	return reportEntries, nil
}

//parseReporstOnChannelToString returns report entries by channel in text format
func (r *Reporter) parseReporstOnChannelToString(reportEntries []reportEntry) string {
	var report string
	if len(reportEntries) == 0 {
		return report + r.Config.Translate.ReportNoData
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		dateTo := value.DateTo
		report += fmt.Sprintf(r.Config.Translate.ReportPeriod, currentDateFrom.Format("2006-01-02"),
			dateTo.Format("2006-01-02"))
		for _, standup := range value.ReportContents[0].Standups {
			report += fmt.Sprintf(r.Config.Translate.ReportStandupFromUser, standup.UsernameID, standup.Comment)
		}
		for _, user := range value.ReportContents[0].NonReporters {
			report += fmt.Sprintf(r.Config.Translate.ReportIgnoredStandup, user.SlackName)
		}
	}
	return report
}

//getReportsOnUser provides report entries on selected user
func (r *Reporter) getReportsOnUser(user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)

		//can retrieve by SlackUserID. Refactor!
		standupUsers, err := r.DB.FindStandupUsers(user.SlackName)
		if err != nil {
			return nil, err
		}
		reportContents := make([]reportContent, 0, len(standupUsers))
		for _, standupUser := range standupUsers {
			currentDayNonReporter := []model.StandupUser{}
			currentDayStandup, err := r.DB.SelectStandupsFiltered(standupUser.SlackUserID, standupUser.ChannelID, dateFrom, dateTo)
			currentDayNonReporter, err = r.DB.GetNonReporter(user.SlackUserID, standupUser.ChannelID, dateFrom, dateTo)
			if err != nil {
				return nil, err
			}

			if len(currentDayNonReporter) > 0 || len(currentDayStandup) > 0 {
				reportContents = append(reportContents,
					reportContent{
						Channel:      standupUser.ChannelID,
						Standups:     currentDayStandup,
						NonReporters: currentDayNonReporter})
			}

		}
		reportEntries = append(reportEntries,
			reportEntry{
				DateFrom:       dateFrom,
				DateTo:         dateTo,
				ReportContents: reportContents})
	}
	return reportEntries, nil
}

//parseReporstOnUserToString returns report entries by user in text format
func (r *Reporter) parseReporstOnUserToString(reportEntries []reportEntry) string {
	var report string
	emptyReport := true
	for _, reportEntry := range reportEntries {
		if len(reportEntry.ReportContents) != 0 {
			emptyReport = false
		}
	}
	if emptyReport {
		return report + r.Config.Translate.ReportNoData
	}

	for _, value := range reportEntries {
		if len(value.ReportContents) == 0 {
			continue
		}
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo

		report += fmt.Sprintf(r.Config.Translate.ReportPeriod, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, reportContent := range value.ReportContents {
			report += fmt.Sprintf(r.Config.Translate.ReportShowChannel, reportContent.Channel)
			for _, standup := range reportContent.Standups {
				report += standup.Comment + "\n"
			}
			for _, user := range reportContent.NonReporters {
				report += fmt.Sprintf(r.Config.Translate.ReportIgnoredStandup, user.SlackName)
			}
			report += "\n"

		}

	}
	return report
}

//getReportsOnChannelAndUser returns report entries by channel
func (r *Reporter) getReportsOnChannelAndUser(channelID string, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)

		currentDayStandup, err := r.DB.SelectStandupsFiltered(user.SlackUserID, channelID, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}

		currentDayNonReporter, err := r.DB.GetNonReporter(user.SlackUserID, channelID, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}

		if len(currentDayNonReporter) > 0 || len(currentDayStandup) > 0 {
			reportContents := make([]reportContent, 0, 1)
			reportContents = append(reportContents,
				reportContent{
					Standups:     currentDayStandup,
					NonReporters: currentDayNonReporter})

			reportEntries = append(reportEntries,
				reportEntry{
					DateFrom:       dateFrom,
					DateTo:         dateTo,
					ReportContents: reportContents})
		}

	}
	return reportEntries, nil
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
func setupDays(dateFrom, dateTo time.Time) (time.Time, int, error) {
	if dateTo.Before(dateFrom) {
		return time.Now(), 0, errors.New("Starting date is bigger than end date")
	}
	if dateTo.After(time.Now()) {
		return time.Now(), 0, errors.New("Report end time was in the future, time range was truncated")
	}
	dateFromRounded := time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateToRounded := time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 0, 0, 0, 0, time.UTC)
	numberOfDays := int(dateToRounded.Sub(dateFromRounded).Hours() / 24)
	return dateFromRounded, numberOfDays, nil
}
