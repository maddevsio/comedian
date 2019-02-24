package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
