package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandup(t *testing.T) {
	testCases := []struct {
		workspaceID  string
		userID       string
		channelID    string
		messageTS    string
		errorMessage string
	}{
		{"", "", "", "", "workspace ID cannot be empty"},
		{"workspaceID", "", "", "", "user ID cannot be empty"},
		{"workspaceID", "userID", "", "", "channel ID cannot be empty"},
		{"workspaceID", "userID", "channelID", "", "MessageTS cannot be empty"},
		{"workspaceID", "userID", "channelID", "12345", ""},
	}
	for _, tt := range testCases {
		st := Standup{
			WorkspaceID: tt.workspaceID,
			UserID:      tt.userID,
			ChannelID:   tt.channelID,
			MessageTS:   tt.messageTS,
		}
		err := st.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestWorkspace(t *testing.T) {
	testCases := []struct {
		workspaceID   string
		workspace     string
		accessToken   string
		reminderTime  int64
		maxReminders  int
		reportingTime string
		language      string
		errorMessage  string
	}{
		{"", "", "", 1, 1, "01:00", "en_US", "workspace ID cannot be empty"},
		{"tID", "", "", 1, 1, "01:00", "en_US", "team name cannot be empty"},
		{"tID", "tName", "", 1, 1, "01:00", "en_US", "accessToken cannot be empty"},
		{"tID", "tName", "accToken", 1, 1, "01:00", "en_US", ""},
		{"tID", "tName", "accToken", 0, 1, "01:00", "en_US", "reminder time cannot be zero or negative"},
		{"tID", "tName", "accToken", -1, 1, "01:00", "en_US", "reminder time cannot be zero or negative"},
		{"tID", "tName", "accToken", 1, 0, "01:00", "en_US", "reminder repeats max cannot be negative"},
		{"tID", "tName", "accToken", 1, -1, "01:00", "en_US", "reminder repeats max cannot be negative"},
		{"tID", "tName", "accToken", 1, 1, "", "en_US", "reporting time cannot be empty"},
		{"tID", "tName", "accToken", 1, 1, "01:00", "", "language cannot be empty"},
	}
	for _, tt := range testCases {
		bs := Workspace{
			WorkspaceID:    tt.workspaceID,
			WorkspaceName:  tt.workspace,
			BotAccessToken: tt.accessToken,
			ReminderOffset: tt.reminderTime,
			MaxReminders:   tt.maxReminders,
			ReportingTime:  tt.reportingTime,
			Language:       tt.language,
		}
		err := bs.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestChannel(t *testing.T) {
	testCases := []struct {
		workspaceID  string
		channelName  string
		channelID    string
		errorMessage string
	}{
		{"", "", "", "workspace ID cannot be empty"},
		{"workspaceID", "", "", "channel name cannot be empty"},
		{"workspaceID", "chanName", "", "channel ID cannot be empty"},
		{"workspaceID", "chanName", "chanID", ""},
	}
	for _, tt := range testCases {
		ch := Project{
			WorkspaceID: tt.workspaceID,
			ChannelName: tt.channelName,
			ChannelID:   tt.channelID,
		}
		err := ch.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestStanduper(t *testing.T) {
	testCases := []struct {
		workspaceID  string
		userID       string
		channelID    string
		errorMessage string
	}{
		{"", "", "", "workspace ID cannot be empty"},
		{"workspaceID", "", "", "user ID cannot be empty"},
		{"workspaceID", "teamName", "", "channel ID cannot be empty"},
		{"workspaceID", "userID", "accessToken", ""},
	}
	for _, tt := range testCases {
		bs := Standuper{
			WorkspaceID: tt.workspaceID,
			UserID:      tt.userID,
			ChannelID:   tt.channelID,
		}
		err := bs.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestNotificationThread(t *testing.T) {
	testCases := []struct {
		channelid        string
		userid           string
		notificationTime int64
		reminderCounter  int
		errorMessage     string
	}{
		{"", "1", int64(2), 0, "Field ChannelID is empty"},
		{"12", "", int64(2), 0, "Field UserID is empty"},
		{"12", "1", -1, -1, "Field NotificationTime cannot be negative"},
		{"12", "1", int64(2), -1, "Field ReminderCounter cannot be negative"},
		{"12", "1", int64(2), 1, ""},
	}
	for _, e := range testCases {
		nt := NotificationThread{
			ChannelID:        e.channelid,
			UserID:           e.userid,
			NotificationTime: e.notificationTime,
			ReminderCounter:  e.reminderCounter,
		}
		err := nt.Validate()
		if err != nil {
			assert.Equal(t, errors.New(e.errorMessage), err)
		}
	}
}
