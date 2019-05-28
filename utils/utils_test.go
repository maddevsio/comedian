package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitUser(t *testing.T) {
	user := "<@USERID|userName"
	id, name := SplitUser(user)
	assert.Equal(t, "USERID", id)
	assert.Equal(t, "userName", name)
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
	testCase := []struct {
		text  string
		day   int
		month int
		year  int
	}{
		{"5.1.2019 ", 1, 5, 2019},
		{"5.01.2019 ", 1, 5, 2019},
		{"5/1/2019 ", 1, 5, 2019},
		{"5/1/18 ", 1, 5, 2018},
		{"5.1.18 ", 1, 5, 2018},
	}
	for _, tt := range testCase {
		time, err := StringToTime(tt.text)
		assert.NoError(t, err)
		assert.Equal(t, tt.day, time.Day())
		assert.Equal(t, tt.month, int(time.Month()))
		assert.Equal(t, tt.year, time.Year())
	}
}
