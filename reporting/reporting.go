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

//Report used to generate report structure
type Report struct {
	ReportHead string
	ReportBody []ReportBodyContent
}

//ReportBodyContent used to generate report body content
type ReportBodyContent struct {
	Date time.Time
	Text string
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
func (r *Reporter) StandupReportByProject(channel model.Channel, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.Config.Translate.ReportOnProjectHead, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("reporting:setupDays failed: %v", err)
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		chanMembers, err := r.db.ListChannelMembers(channel.ChannelID)
		if err != nil || len(chanMembers) == 0 {
			continue
		}
		dayInfo := ""
		for _, member := range chanMembers {
			if !r.db.MemberShouldBeTracked(member.ID, dateFrom, dateTo) {
				logrus.Infof("member should not be tracked: %v", member.UserID)
				continue
			}
			userIsNonReporter, err := r.db.IsNonReporter(member.UserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProject IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, member.UserID)
			} else {
				standup, err := r.db.SelectStandupsFiltered(member.UserID, channel.ChannelID, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting:SelectStandupsFiltered failed: %v", err)
					continue
				}
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandup, member.UserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}

// StandupReportByUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByUser(slackUserID string, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.Config.Translate.ReportOnUserHead, slackUserID, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		channels, err := r.db.GetUserChannels(slackUserID)
		if err != nil || len(channels) == 0 {
			continue
		}
		dayInfo := ""
		for _, channel := range channels {
			channelName, err := r.db.GetChannelName(channel)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser GetChannelName failed: %v", err)
				continue
			}
			member, err := r.db.FindChannelMemberByUserID(slackUserID, channel)
			if err != nil {
				logrus.Infof("FindChannelMemberByUserID failed: %v", err)
				continue
			}
			if !r.db.MemberShouldBeTracked(member.ID, dateFrom, dateTo) {
				logrus.Infof("member should not be tracked: %v", slackUserID)
				continue
			}
			userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByUser IsNonReporter failed: %v", err)
				continue
			}
			if userIsNonReporter {
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandupInChannel, channelName, slackUserID)
			} else {
				standup, err := r.db.SelectStandupsFiltered(slackUserID, channel, dateFrom, dateTo)
				if err != nil {
					logrus.Errorf("reporting.go reportByUser SelectStandupsFiltered failed: %v", err)
				}
				dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandupInChannel, channelName, slackUserID)
				dayInfo += fmt.Sprintf("%v \n", standup.Comment)
			}
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
	return report, nil
}

// StandupReportByProjectAndUser creates a standup report for a specified period of time
func (r *Reporter) StandupReportByProjectAndUser(channel model.Channel, slackUserID string, dateFrom, dateTo time.Time) (Report, error) {
	report := Report{}
	report.ReportHead = fmt.Sprintf(r.Config.Translate.ReportOnProjectAndUserHead, slackUserID, channel.ChannelName, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
	dateFromBegin, numberOfDays, err := r.setupDays(dateFrom, dateTo)
	if err != nil {
		return report, err
	}
	for day := 0; day <= numberOfDays; day++ {
		dayInfo := ""
		dateFrom := dateFromBegin.Add(time.Duration(day*24) * time.Hour)
		dateTo := dateFrom.Add(24 * time.Hour)
		logrus.Infof("reportByProjectAndUser: dateFrom: '%v', dateTo: '%v'", dateFrom, dateTo)
		member, err := r.db.FindChannelMemberByUserID(slackUserID, channel.ChannelID)
		if err != nil {
			logrus.Infof("FindChannelMemberByUserID failed: %v", err)
			continue
		}
		if !r.db.MemberShouldBeTracked(member.ID, dateFrom, dateTo) {
			logrus.Infof("member should not be tracked: %v", slackUserID)
			continue
		}
		userIsNonReporter, err := r.db.IsNonReporter(slackUserID, channel.ChannelID, dateFrom, dateTo)
		if err != nil {
			logrus.Errorf("reporting.go reportByProjectAndUser IsNonReporter failed: %v", err)
			continue
		}
		if userIsNonReporter {
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidNotStandup, slackUserID)
			dayInfo += "\n"
		} else {
			standup, err := r.db.SelectStandupsFiltered(slackUserID, channel.ChannelID, dateFrom, dateTo)
			if err != nil {
				logrus.Errorf("reporting.go reportByProjectAndUser SelectStandupsFiltered failed: %v", err)
				continue
			}
			dayInfo += fmt.Sprintf(r.Config.Translate.UserDidStandup, slackUserID)
			dayInfo += fmt.Sprintf("%v \n", standup.Comment)
		}
		if dayInfo != "" {
			text := fmt.Sprintf(r.Config.Translate.ReportDate, dateFrom.Format("2006-01-02"))
			text += dayInfo
			rbc := ReportBodyContent{dateFrom, text}
			report.ReportBody = append(report.ReportBody, rbc)
		}
	}
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
