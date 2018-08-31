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
	"github.com/sirupsen/logrus"
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

	// UserData used to parse data on user from Collector
	UserData struct {
		TotalCommits int `json:"total_commits"`
		TotalMerges  int `json:"total_merges"`
		Worklogs     int `json:"worklogs"`
	}

	// ProjectData used to parse data on project from Collector
	ProjectData struct {
		TotalCommits int `json:"total_commits"`
		TotalMerges  int `json:"total_merges"`
	}

	// ProjectUserData used to parse data on user in project from Collector
	ProjectUserData struct {
		TotalCommits int `json:"total_commits"`
		TotalMerges  int `json:"total_merges"`
	}
)

//NewReporter creates new reporter instanse
func NewReporter(c config.Config) (*Reporter, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		logrus.Errorf("notifier: NewMySQL failed: %v\n", err)
		return nil, err
	}

	r := &Reporter{DB: conn, Config: c}
	logrus.Infof("notifier: Created Reporter: %v", r)
	return r, nil
}

// StandupReportByProject creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProject(channelID string, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	reportEntries, err := r.ChanReports(channel, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: ChanReports failed: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("report entries: %#v\n", reportEntries)

	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channel)
	report += r.ChanRepsToStr(reportEntries)

	var dataP ProjectData
	json.Unmarshal(collectorData, &dataP)

	report += fmt.Sprintf("\n\nCommits for period: %v \nMerges for period: %v\n", dataP.TotalCommits, dataP.TotalMerges)
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(user model.StandupUser, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	reportEntries, err := r.UserReports(user, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: UserReports failed: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("reporting: report entries: %#v\n", reportEntries)
	report := fmt.Sprintf(r.Config.Translate.ReportOnUserHead, user.SlackName)
	report += r.UserRepsToStr(reportEntries)

	var dataU UserData
	json.Unmarshal(collectorData, &dataU)

	report += fmt.Sprintf(r.Config.Translate.ReportCollectorDataUser, dataU.TotalCommits, dataU.TotalMerges, dataU.Worklogs/3600)

	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channelID string, user model.StandupUser, dateFrom, dateTo time.Time, collectorData []byte) (string, error) {
	channel := strings.Replace(channelID, "#", "", -1)
	reportEntries, err := r.ChanReportsAndUser(channel, user, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: ChanReportsAndUser: %v\n", err)
		return "Error!", err
	}
	logrus.Infof("reporting: report entries: %#v\n", reportEntries)

	report := fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, channel, user.SlackName)
	report += r.ChanRepsToStr(reportEntries)

	var dataPU ProjectUserData
	json.Unmarshal(collectorData, &dataPU)

	report += fmt.Sprintf("\n\nCommits for period: %v \nMerges for period: %v\n", dataPU.TotalCommits, dataPU.TotalMerges)

	return report, nil
}

//ChanReports returns report entries by channel
func (r *Reporter) ChanReports(channelID string, dateFrom, dateTo time.Time) ([]reportEntry, error) {
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

		currentDayStandups, err := r.DB.SelectStandupsByChannelIDForPeriod(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v", err)
			return nil, err
		}
		currentDayNonReporters, err := r.DB.GetNonReporters(channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v", err)
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
func (r *Reporter) ChanRepsToStr(reportEntries []reportEntry) string {
	var report string
	if len(reportEntries) == 0 {
		return report + r.Config.Translate.ReportNoData
	}

	for _, value := range reportEntries {
		currentDateFrom := value.DateFrom
		currentDateTo := value.DateTo
		report += fmt.Sprintf(r.Config.Translate.ReportPeriod, currentDateFrom.Format("2006-01-02"),
			currentDateTo.Format("2006-01-02"))
		for _, standup := range value.ReportContents[0].Standups {

			report += fmt.Sprintf(r.Config.Translate.ReportStandupFromUser, standup.UsernameID, standup.Comment)
		}
		for _, user := range value.ReportContents[0].NonReporters {
			report += fmt.Sprintf(r.Config.Translate.ReportIgnoredStandup, user.SlackName)
		}
	}
	return report
}

func (r *Reporter) UserReports(user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: setupDays failed: %v\n", err)
		return nil, err
	}
	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		standupUsers, err := r.DB.FindStandupUsers(user.SlackName)
		if err != nil {
			logrus.Errorf("reporting: FindStandupUser failed: %v\n", err)
			return nil, err
		}
		reportContents := make([]reportContent, 0, len(standupUsers))
		for _, standupUser := range standupUsers {
			currentDayNonReporter := []model.StandupUser{}
			currentDayStandup, err := r.DB.SelectStandupsFiltered(standupUser.SlackUserID, standupUser.ChannelID, currentDateFrom, currentDateTo)

			currentDayNonReporter, err = r.DB.GetNonReporter(user.SlackUserID, standupUser.ChannelID, currentDateFrom, currentDateTo)
			if err != nil {
				logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v", err)
				return nil, err
			}

			logrus.Infof("reporting: userReport, current day standups: %#v\n", currentDayStandup)
			logrus.Infof("reporting: userReport, current day non reporters: %#v\n", currentDayNonReporter)

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
				DateFrom:       currentDateFrom,
				DateTo:         currentDateTo,
				ReportContents: reportContents})
	}
	logrus.Infof("reporting: userReport, final report entries: %#v\n", reportEntries)
	return reportEntries, nil
}

//ReportEntriesByUserToString provides reporting entries in selected time period
func (r *Reporter) UserRepsToStr(reportEntries []reportEntry) string {
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

//ChanReportsAndUser returns report entries by channel
func (r *Reporter) ChanReportsAndUser(channelID string, user model.StandupUser, dateFrom, dateTo time.Time) ([]reportEntry, error) {
	dateFromRounded, numberOfDays, err := setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting: setupDays failed: %v\n", err)
		return nil, err
	}

	reportEntries := make([]reportEntry, 0, numberOfDays)
	for day := 0; day <= numberOfDays; day++ {
		currentDateFrom := dateFromRounded.Add(time.Duration(day*24) * time.Hour)
		currentDateTo := currentDateFrom.Add(24 * time.Hour)

		currentDayStandup, err := r.DB.SelectStandupsFiltered(user.SlackUserID, channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandups failed: %v\n", err)
		}

		currentDayNonReporter, err := r.DB.GetNonReporter(user.SlackUserID, channelID, currentDateFrom, currentDateTo)
		if err != nil {
			logrus.Errorf("reporting: SelectStandupsByChannelIDForPeriod failed: %v", err)
			return nil, err
		}

		logrus.Infof("reporting: projectUserReport, current day standups: %#v\n", currentDayStandup)
		logrus.Infof("reporting: projectUserReport, current day non reporters: %#v\n", currentDayNonReporter)

		if len(currentDayNonReporter) > 0 || len(currentDayStandup) > 0 {
			reportContents := make([]reportContent, 0, 1)
			reportContents = append(reportContents,
				reportContent{
					Standups:     currentDayStandup,
					NonReporters: currentDayNonReporter})

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
