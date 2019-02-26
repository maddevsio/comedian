package model

import (
	"errors"
	"time"

	"github.com/nlopes/slack"
)

type (
	// Standup model used for serialization/deserialization stored standups
	Standup struct {
		ID        int64     `db:"id" json:"id"`
		TeamID    string    `db:"team_id" json:"team_id"`
		Created   time.Time `db:"created" json:"created"`
		Modified  time.Time `db:"modified" json:"modified"`
		ChannelID string    `db:"channel_id" json:"channelId"`
		UserID    string    `db:"user_id" json:"userId"`
		Comment   string    `db:"comment" json:"comment"`
		MessageTS string    `db:"message_ts" json:"message_ts"`
	}

	// User model used for serialization/deserialization stored Users
	User struct {
		ID       int64  `db:"id" json:"id"`
		TeamID   string `db:"team_id" json:"team_id"`
		UserName string `db:"user_name" json:"user_name"`
		UserID   string `db:"user_id" json:"user_id"`
		Role     string `db:"role" json:"role"`
		RealName string `db:"real_name" json:"real_name"`
	}

	// Channel model used for serialization/deserialization stored Channels
	Channel struct {
		ID          int64  `db:"id" json:"id"`
		TeamID      string `db:"team_id" json:"team_id"`
		ChannelName string `db:"channel_name" json:"channel_name"`
		ChannelID   string `db:"channel_id" json:"channel_id"`
		StandupTime int64  `db:"channel_standup_time" json:"time"`
	}

	// ChannelMember model used for serialization/deserialization stored ChannelMembers
	ChannelMember struct {
		ID            int64     `db:"id" json:"id"`
		TeamID        string    `db:"team_id" json:"team_id"`
		UserID        string    `db:"user_id" json:"user_id"`
		ChannelID     string    `db:"channel_id" json:"channel_id"`
		RoleInChannel string    `db:"role_in_channel" json:"role_in_channel"`
		StandupTime   int64     `db:"standup_time" json:"time"`
		Created       time.Time `db:"created" json:"created"`
	}

	// ControlPannel used for updating and storing different configuration parameters
	ControlPannel struct {
		ID                        int64  `db:"id"`
		NotifierInterval          int    `db:"notifier_interval" json:"notifier_interval" schema:"notifier_interval"`
		ManagerSlackUserID        string `db:"manager_slack_user_id" json:"manager_slack_user_id" schema:"manager_slack_user_id"`
		ReportingChannel          string `db:"reporting_channel" json:"reporting_channel" schema:"reporting_channel"`
		IndividualReportingStatus bool   `db:"individual_reporting_status" json:"individual_reporting_status" schema:"individual_reporting_status"`
		ReportTime                string `db:"report_time" json:"report_time" schema:"report_time"`
		Language                  string `db:"language" json:"language" schema:"language"`
		ReminderRepeatsMax        int    `db:"reminder_repeats_max" json:"reminder_repeats_max" schema:"reminder_repeats_max"`
		ReminderTime              int64  `db:"reminder_time" json:"reminder_time" schema:"reminder_time"`
		CollectorEnabled          bool   `db:"collector_enabled" json:"collector_enabled" schema:"collector_enabled"`
		SprintReportStatus        bool   `db:"sprint_report_status" json:"sprint_report_status" schema:"sprint_report_status"`
		SprintReportTime          string `db:"sprint_report_time" json:"sprint_report_time" schema:"sprint_report_time"`
		SprintReportChannel       string `db:"sprint_report_channel" json:"sprint_report_channel" schema:"sprint_report_channel"`
		SprintWeekdays            string `db:"sprint_weekdays" json:"sprint_weekdays" schema:"sprint_weekdays"`
		TaskDoneStatus            string `db:"task_done_status" json:"task_done_status" schema:"task_done_status"`
		AccessToken               string `db:"bot_access_token" json:"bot_access_token" schema:"bot_access_token"`
		TeamID                    string `db:"team_id" json:"team_id" schema:"team_id"`
		TeamName                  string `db:"team_name" json:"team_name" schema:"team_name"`
		Password                  string `db:"password" json:"password" schema:"password"`
	}

	// FullSlackForm struct used for parsing full payload from slack
	FullSlackForm struct {
		Command     string `schema:"command"`
		Text        string `schema:"text"`
		ChannelID   string `schema:"channel_id"`
		ChannelName string `schema:"channel_name"`
		UserID      string `schema:"user_id"`
		UserName    string `schema:"user_name"`
		TeamID      string `schema:"team_id"`
		TeamDomain  string `schema:"team_domain"`
	}
)

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

//AttachmentItem is needed to sort attachments
type AttachmentItem struct {
	SlackAttachment slack.Attachment
	Points          int
}

// Validate validates Standup struct
func (c Standup) Validate() error {
	if c.UserID == "" {
		err := errors.New("User cannot be empty")
		return err
	}
	if c.ChannelID == "" {
		err := errors.New("Channel cannot be empty")
		return err
	}
	return nil
}

// Validate validates StandupUser struct
func (c ChannelMember) Validate() error {
	if c.UserID == "" && c.ChannelID == "" {
		err := errors.New("User/Channel cannot be empty")
		return err
	}
	return nil
}

//IsAdmin returns user status
func (u User) IsAdmin() bool {
	return u.Role == "admin"
}
