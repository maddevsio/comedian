package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestSplitUser(t *testing.T) {
	user := "<@USERID|userName"
	id, name := SplitUser(user)
	assert.Equal(t, "USERID", id)
	assert.Equal(t, "userName", name)
}

func TestSplitChannel(t *testing.T) {
	channel := "<#ChannelID|channelName"
	id, name := SplitChannel(channel)
	assert.Equal(t, "ChannelID", id)
	assert.Equal(t, "channelName", name)
}

func TestSecondsToHuman(t *testing.T) {
	testCase := []struct {
		secondsInt int
		secondsStr string
	}{
		{3600, "1:00"},
	}
	for _, test := range testCase {
		actual := SecondsToHuman(test.secondsInt)
		assert.Equal(t, test.secondsStr, actual)
	}
}

func TestFormatTime(t *testing.T) {
	testCase := []struct {
		time  string
		eHour int
		eMin  int
		err   error
	}{
		{"10:00", 10, 0, nil},
		{"10", 0, 0, errors.New("time format error")},
		{"-10:00", -10, 0, errors.New("time format error")},
		{"24:00", 24, 0, errors.New("time format error")},
		{"10:-01", 10, -1, errors.New("time format error")},
		{"10:69", 10, 69, errors.New("time format error")},
	}
	for _, test := range testCase {
		aHour, aMin, err := FormatTime(test.time)
		assert.Equal(t, test.eHour, aHour)
		assert.Equal(t, test.eMin, aMin)
		assert.Equal(t, test.err, err)
	}
}

func TestPrepareTimetable(t *testing.T) {
	c, err := config.Get()
	bot, err := bot.NewBot(c)

	m, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "testUser",
		ChannelID: "testChannel",
	})
	assert.NoError(t, err)

	tt, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: m.ID,
	})
	assert.NoError(t, err)

	timeNow := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	tt.Monday = timeNow.Unix()
	tt.Tuesday = timeNow.Unix()
	tt.Wednesday = timeNow.Unix()
	tt.Thursday = timeNow.Unix()
	tt.Friday = timeNow.Unix()

	tt, err = bot.DB.UpdateTimeTable(tt)

	assert.NoError(t, err)
	assert.Equal(t, timeNow.Unix(), tt.Monday)

	timeUpdate := time.Date(2018, 10, 7, 12, 0, 0, 0, time.UTC).Unix()

	tt = PrepareTimeTable(tt, "mon tue wed thu fri sat sun", timeUpdate)
	assert.Equal(t, timeUpdate, tt.Monday)
	assert.NoError(t, bot.DB.DeleteChannelMember(m.UserID, m.ChannelID))
	assert.NoError(t, bot.DB.DeleteTimeTable(tt.ID))

}

func TestCommandParsing(t *testing.T) {
	testCase := []struct {
		text         string
		commandTitle string
		commandBody  string
	}{
		{"add user1", "add", "user1"},
		{"add user1 user2", "add", "user1 user2"},
		{"add user1 user2 user3", "add", "user1 user2 user3"},
	}
	for _, test := range testCase {
		aTitle, aBody := CommandParsing(test.text)
		assert.Equal(t, test.commandTitle, aTitle)
		assert.Equal(t, test.commandBody, aBody)
	}
}
