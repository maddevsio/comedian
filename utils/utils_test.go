package utils

import (
	"testing"

	"github.com/labstack/gommon/log"
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

func TestStringToTime(t *testing.T) {
	time, err := StringToTime(" 5.1.2019 ")
	assert.NoError(t, err)
	log.Info(time)
	assert.Equal(t, 1, time.Day())
	assert.Equal(t, 5, int(time.Month()))
	assert.Equal(t, 2019, time.Year())

	time, err = StringToTime(" 5/1/19 ")
	assert.NoError(t, err)
	log.Info(time)
	assert.Equal(t, 1, time.Day())
	assert.Equal(t, 5, int(time.Month()))
	assert.Equal(t, 2019, time.Year())
}
