package model

import (
	"errors"
	"time"

	"github.com/nlopes/slack"
)

// Standup model used for serialization/deserialization stored standups
type Standup struct {
	ID        int64     `db:"id" json:"id"`
	TeamID    string    `db:"team_id" json:"team_id"`
	Created   time.Time `db:"created" json:"created"`
	Modified  time.Time `db:"modified" json:"modified"`
	ChannelID string    `db:"channel_id" json:"channel_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Comment   string    `db:"comment" json:"comment"`
	MessageTS string    `db:"message_ts" json:"message_ts"`
}

// User model used for serialization/deserialization stored Users
type User struct {
	ID       int64  `db:"id" json:"id"`
	TeamID   string `db:"team_id" json:"team_id"`
	UserName string `db:"user_name" json:"user_name"`
	UserID   string `db:"user_id" json:"user_id"`
	Role     string `db:"role" json:"role"`
	RealName string `db:"real_name" json:"real_name"`
}

// Channel model used for serialization/deserialization stored Channels
type Channel struct {
	ID          int64  `db:"id" json:"id"`
	TeamID      string `db:"team_id" json:"team_id"`
	ChannelName string `db:"channel_name" json:"channel_name"`
	ChannelID   string `db:"channel_id" json:"channel_id"`
	StandupTime string `db:"channel_standup_time" json:"channel_standup_time"`
}

// Standuper model used for serialization/deserialization stored ChannelMembers
type Standuper struct {
	ID                    int64     `db:"id" json:"id"`
	TeamID                string    `db:"team_id" json:"team_id"`
	UserID                string    `db:"user_id" json:"user_id"`
	ChannelID             string    `db:"channel_id" json:"channel_id"`
	RoleInChannel         string    `db:"role_in_channel" json:"role_in_channel"`
	SubmittedStandupToday bool      `db:"submitted_standup_today" json:"submitted_standup_today"`
	Created               time.Time `db:"created" json:"created"`
	RealName              string    `db:"real_name" json:"real_name"`
	ChannelName           string    `db:"channel_name" json:"channel_name"`
}

// BotSettings is used for updating and storing different bot configuration parameters
type BotSettings struct {
	ID                  int64  `db:"id" json:"id"`
	UserID              string `db:"user_id" json:"user_id"`
	NotifierInterval    int    `db:"notifier_interval" json:"notifier_interval" `
	Language            string `db:"language" json:"language" `
	ReminderRepeatsMax  int    `db:"reminder_repeats_max" json:"reminder_repeats_max" `
	ReminderTime        int64  `db:"reminder_time" json:"reminder_time" `
	AccessToken         string `db:"bot_access_token" json:"bot_access_token" `
	TeamID              string `db:"team_id" json:"team_id" `
	TeamName            string `db:"team_name" json:"team_name" `
	Admin               bool   `db:"admin" json:"admin" `
	Password            string `db:"password" json:"password" `
	ReportingChannel    string `db:"reporting_channel" json:"reporting_channel"`
	ReportingTime       string `db:"reporting_time" json:"reporting_time"`
	IndividualReportsOn bool   `db:"individual_reports_on" json:"individual_reports_on"`
}

//Role sets the abilities for different roles
type Role struct {
	ID            int64  `db:"id" json:"id"`
	Title         string `db:"title" json:"title"`
	AccessLevel   int    `db:"access_level" json:"access_level"`
	ShouldStandup bool   `db:"should_standup" json:"should_standup"`
	ShouldCommit  bool   `db:"should_commit" json:"should_commit"`
	ShouldLogTime bool   `db:"should_log_time" json:"should_log_time"`
}

// ServiceEvent event coming from services
type ServiceEvent struct {
	TeamName    string             `json:"team_name"`
	AccessToken string             `json:"bot_access_token"`
	Channel     string             `json:"channel"`
	Message     string             `json:"message"`
	Attachments []slack.Attachment `json:"attachments"`
}

// InfoEvent event coming from services
type InfoEvent struct {
	TeamName    string `json:"team_name"`
	InfoType    string `json:"info_type"`
	AccessToken string `json:"bot_access_token"`
	Channel     string `json:"channel"`
	Message     string `json:"message"`
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

//AttachmentItem is needed to sort attachments
type AttachmentItem struct {
	SlackAttachment slack.Attachment
	Points          int
}

// Validate validates Standup struct
func (st Standup) Validate() error {
	if st.TeamID == "" {
		err := errors.New("team ID cannot be empty")
		return err
	}
	if st.UserID == "" {
		err := errors.New("user ID cannot be empty")
		return err
	}
	if st.ChannelID == "" {
		err := errors.New("channel ID cannot be empty")
		return err
	}
	if st.MessageTS == "" {
		err := errors.New("MessageTS cannot be empty")
		return err
	}
	return nil
}

// Validate validates BotSettings struct
func (bs BotSettings) Validate() error {
	if bs.TeamID == "" {
		err := errors.New("team ID cannot be empty")
		return err
	}

	if bs.TeamName == "" {
		err := errors.New("team name cannot be empty")
		return err
	}

	if bs.AccessToken == "" {
		err := errors.New("accessToken cannot be empty")
		return err
	}

	if bs.Password == "" {
		err := errors.New("password cannot be empty")
		return err
	}

	if bs.ReminderTime <= 0 {
		err := errors.New("reminder time cannot be zero or negative")
		return err
	}

	if bs.ReminderRepeatsMax <= 0 {
		err := errors.New("reminder repeats max cannot be zero or negative")
		return err
	}

	if bs.ReportingTime == "" {
		err := errors.New("reporting time cannot be empty")
		return err
	}

	if bs.Language == "" {
		err := errors.New("language cannot be empty")
		return err
	}

	return nil
}

// Validate validates Channel struct
func (ch Channel) Validate() error {
	if ch.TeamID == "" {
		err := errors.New("team ID cannot be empty")
		return err
	}

	if ch.ChannelName == "" {
		err := errors.New("channel name cannot be empty")
		return err
	}

	if ch.ChannelID == "" {
		err := errors.New("channel ID cannot be empty")
		return err
	}

	return nil
}

// Validate validates Standuper struct
func (s Standuper) Validate() error {
	if s.TeamID == "" {
		err := errors.New("team ID cannot be empty")
		return err
	}

	if s.UserID == "" {
		err := errors.New("user ID cannot be empty")
		return err
	}

	if s.ChannelID == "" {
		err := errors.New("channel ID cannot be empty")
		return err
	}

	return nil
}

// Validate validates User struct
func (u User) Validate() error {
	if u.TeamID == "" {
		err := errors.New("team ID cannot be empty")
		return err
	}

	if u.UserName == "" {
		err := errors.New("user name cannot be empty")
		return err
	}

	if u.UserID == "" {
		err := errors.New("user ID cannot be empty")
		return err
	}

	return nil
}

//IsAdmin returns true if user has admin role
func (u User) IsAdmin() bool {
	if u.Role == "admin" || u.Role == "super-admin" {
		return true
	}
	return false
}

//IsPM returns true if standuper has pm status
func (s Standuper) IsPM() bool {
	return s.RoleInChannel == "pm"
}

//IsDesigner returns true if standuper has designer status
func (s Standuper) IsDesigner() bool {
	return s.RoleInChannel == "designer"
}
