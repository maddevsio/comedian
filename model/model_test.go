package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandup(t *testing.T) {
	testCases := []struct {
		teamID       string
		userID       string
		channelID    string
		messageTS    string
		errorMessage string
	}{
		{"", "", "", "", "team ID cannot be empty"},
		{"teamID", "", "", "", "user ID cannot be empty"},
		{"teamID", "userID", "", "", "channel ID cannot be empty"},
		{"teamID", "userID", "channelID", "", "MessageTS cannot be empty"},
		{"teamID", "userID", "channelID", "12345", ""},
	}
	for _, tt := range testCases {
		st := Standup{
			TeamID:    tt.teamID,
			UserID:    tt.userID,
			ChannelID: tt.channelID,
			MessageTS: tt.messageTS,
		}
		err := st.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestBotSettings(t *testing.T) {
	testCases := []struct {
		teamID       string
		teamName     string
		accessToken  string
		errorMessage string
	}{
		{"", "", "", "team ID cannot be empty"},
		{"teamID", "", "", "team name cannot be empty"},
		{"teamID", "teamName", "", "accessToken cannot be empty"},
		{"teamID", "userID", "accessToken", ""},
	}
	for _, tt := range testCases {
		bs := BotSettings{
			TeamID:      tt.teamID,
			TeamName:    tt.teamName,
			AccessToken: tt.accessToken,
			Password:    "Foo",
		}
		err := bs.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestChannel(t *testing.T) {
	testCases := []struct {
		teamID       string
		channelName  string
		channelID    string
		errorMessage string
	}{
		{"", "", "", "team ID cannot be empty"},
		{"teamID", "", "", "channel name cannot be empty"},
		{"teamID", "chanName", "", "channel ID cannot be empty"},
		{"teamID", "chanName", "chanID", ""},
	}
	for _, tt := range testCases {
		ch := Channel{
			TeamID:      tt.teamID,
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
		teamID       string
		userID       string
		channelID    string
		errorMessage string
	}{
		{"", "", "", "team ID cannot be empty"},
		{"teamID", "", "", "user ID cannot be empty"},
		{"teamID", "teamName", "", "channel ID cannot be empty"},
		{"teamID", "userID", "accessToken", ""},
	}
	for _, tt := range testCases {
		bs := Standuper{
			TeamID:    tt.teamID,
			UserID:    tt.userID,
			ChannelID: tt.channelID,
		}
		err := bs.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestUser(t *testing.T) {
	testCases := []struct {
		teamID       string
		userName     string
		userID       string
		errorMessage string
	}{
		{"", "", "", "team ID cannot be empty"},
		{"teamID", "", "", "user name cannot be empty"},
		{"teamID", "userName", "", "user ID cannot be empty"},
		{"teamID", "userName", "userID", ""},
	}
	for _, tt := range testCases {
		bs := User{
			TeamID:   tt.teamID,
			UserName: tt.userName,
			UserID:   tt.userID,
		}
		err := bs.Validate()
		if err != nil {
			assert.Equal(t, errors.New(tt.errorMessage), err)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	user := User{
		Role: "admin",
	}
	admin := user.IsAdmin()
	assert.True(t, admin)

	standuper := Standuper{
		RoleInChannel: "pm",
	}
	pm := standuper.IsPM()
	assert.True(t, pm)

	standuper = Standuper{
		RoleInChannel: "designer",
	}
	designer := standuper.IsDesigner()
	assert.True(t, designer)
}
