package model

import (
	"errors"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// Standup model used for serialization/deserialization stored standups
type Standup struct {
	ID          int64  `db:"id" json:"id"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	WorkspaceID string `db:"workspace_id" json:"workspace_id"`
	ChannelID   string `db:"channel_id" json:"channel_id"`
	UserID      string `db:"user_id" json:"user_id"`
	Comment     string `db:"comment" json:"comment"`
	MessageTS   string `db:"message_ts" json:"message_ts"`
}

// Project model used for serialization/deserialization stored Projects
type Project struct {
	ID               int64  `db:"id" json:"id"`
	CreatedAt        int64  `db:"created_at" json:"created_at"`
	WorkspaceID      string `db:"workspace_id" json:"workspace_id"`
	ChannelName      string `db:"channel_name" json:"channel_name"`
	ChannelID        string `db:"channel_id" json:"channel_id"`
	Deadline         string `db:"deadline" json:"deadline"`
	TZ               string `db:"tz" json:"tz"`
	OnbordingMessage string `db:"onbording_message" json:"onbording_message,omitempty"`
	SubmissionDays   string `db:"submission_days" json:"submission_days,omitempty"`
}

// Standuper model used for serialization/deserialization stored ChannelMembers
type Standuper struct {
	ID          int64  `db:"id" json:"id"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	WorkspaceID string `db:"workspace_id" json:"workspace_id"`
	UserID      string `db:"user_id" json:"user_id"`
	ChannelID   string `db:"channel_id" json:"channel_id"`
	Role        string `db:"role" json:"role"`
	RealName    string `db:"real_name" json:"real_name"`
	ChannelName string `db:"channel_name" json:"channel_name"`
}

// Workspace is used for updating and storing different bot configuration parameters
type Workspace struct {
	ID                     int64  `db:"id" json:"id"`
	CreatedAt              int64  `db:"created_at" json:"created_at"`
	BotUserID              string `db:"bot_user_id" json:"bot_user_id"`
	NotifierInterval       int    `db:"notifier_interval" json:"notifier_interval" `
	Language               string `db:"language" json:"language" `
	MaxReminders           int    `db:"max_reminders" json:"max_reminders" `
	ReminderOffset         int64  `db:"reminder_offset" json:"reminder_offset" `
	BotAccessToken         string `db:"bot_access_token" json:"bot_access_token" `
	WorkspaceID            string `db:"workspace_id" json:"workspace_id" `
	WorkspaceName          string `db:"workspace_name" json:"workspace_name" `
	ReportingChannel       string `db:"reporting_channel" json:"reporting_channel"`
	ReportingTime          string `db:"reporting_time" json:"reporting_time"`
	ProjectsReportsEnabled bool   `db:"projects_reports_enabled" json:"projects_reports_enabled"`
}

// ServiceEvent event coming from services
type ServiceEvent struct {
	TeamName    string             `json:"team_name"`
	AccessToken string             `json:"bot_access_token"`
	Channel     string             `json:"channel"`
	Message     string             `json:"message"`
	Attachments []slack.Attachment `json:"attachments,omitempty"`
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

//NotificationThread ...
type NotificationThread struct {
	ID               int64  `db:"id" json:"id"`
	ChannelID        string `db:"channel_id" json:"channel_id"`
	UserIDs          string `db:"user_ids" json:"user_ids"`
	NotificationTime int64  `db:"notification_time" json:"notification_time"`
	ReminderCounter  int    `db:"reminder_counter" json:"reminder_counter"`
}

// Validate validates Standup struct
func (st Standup) Validate() error {
	if st.WorkspaceID == "" {
		err := errors.New("workspace ID cannot be empty")
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

// Validate validates Workspace struct
func (bs Workspace) Validate() error {
	if bs.WorkspaceID == "" {
		err := errors.New("workspace ID cannot be empty")
		return err
	}

	if bs.WorkspaceName == "" {
		err := errors.New("team name cannot be empty")
		return err
	}

	if bs.BotAccessToken == "" {
		err := errors.New("accessToken cannot be empty")
		return err
	}

	if bs.ReminderOffset <= 0 {
		err := errors.New("reminder time cannot be zero or negative")
		return err
	}

	if bs.MaxReminders < 0 {
		err := errors.New("reminder repeats max cannot be negative")
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

// Validate validates Project struct
func (ch Project) Validate() error {
	if ch.WorkspaceID == "" {
		err := errors.New("workspace ID cannot be empty")
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
	if s.WorkspaceID == "" {
		err := errors.New("workspace ID cannot be empty")
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

// Validate validates NotificationsThread struct
func (nt NotificationThread) Validate() error {
	if strings.TrimSpace(nt.ChannelID) == "" {
		return errors.New("Field ChannelID is empty")
	}
	if strings.TrimSpace(nt.UserIDs) == "" {
		return errors.New("Field UserIDs is empty")
	}
	if nt.NotificationTime < 0 {
		return errors.New("Field NotificationTime cannot be negative")
	}
	if nt.ReminderCounter < 0 {
		return errors.New("Field ReminderCounter cannot be negative")
	}
	return nil
}
