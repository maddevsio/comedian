package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStandup(t *testing.T) {
	s := Standup{
		UserID:    "uid",
		ChannelID: "channel",
	}
	//userid empty standup
	s1 := Standup{
		UserID:    "",
		ChannelID: "channel1",
	}
	//channelid empty
	s2 := Standup{
		UserID:    "uid2",
		ChannelID: "",
	}
	testCase := []struct {
		s        Standup
		expected error
	}{
		{s, nil},
		{s1, errors.New("User cannot be empty")},
		{s2, errors.New("Channel cannot be empty")},
	}
	for _, test := range testCase {
		actual := test.s.Validate()
		assert.Equal(t, test.expected, actual)
	}
}

func TestValidateChannelMember(t *testing.T) {
	cm := ChannelMember{
		UserID:    "uid",
		ChannelID: "channel",
	}
	cm1 := ChannelMember{
		UserID:    "",
		ChannelID: "channel1",
	}
	cm2 := ChannelMember{
		UserID:    "uid2",
		ChannelID: "",
	}
	cm3 := ChannelMember{
		UserID:    "",
		ChannelID: "",
	}

	testCase := []struct {
		chm      ChannelMember
		expected error
	}{
		{cm, nil},
		{cm1, nil},
		{cm2, nil},
		{cm3, errors.New("User/Channel cannot be empty")},
	}
	for _, test := range testCase {
		actual := test.chm.Validate()
		assert.Equal(t, test.expected, actual)
	}
}

func TestValidateStandupEditHistory(t *testing.T) {
	s1 := StandupEditHistory{
		StandupText: "standup",
	}
	s2 := StandupEditHistory{
		StandupText: "",
	}
	testCase := []struct {
		s        StandupEditHistory
		expected error
	}{
		{s1, nil},
		{s2, errors.New("Text cannot be empty")},
	}
	for _, test := range testCase {
		actual := test.s.Validate()
		assert.Equal(t, test.expected, actual)
	}
}

func TestIsAdmin(t *testing.T) {
	admin := User{
		Role: "admin",
	}
	notAdmin := User{
		Role: "dev",
	}
	testCase := []struct {
		user     User
		expected bool
	}{
		{admin, true},
		{notAdmin, false},
	}
	for _, test := range testCase {
		actual := test.user.IsAdmin()
		assert.Equal(t, test.expected, actual)
	}
}

func TestShowDeadlineOn(t *testing.T) {
	timeTable := TimeTable{
		Monday:    1,
		Tuesday:   0,
		Wednesday: 1,
		Thursday:  0,
		Friday:    1,
		Saturday:  1,
		Sunday:    0,
	}
	testCase := []struct {
		Day      string
		Expected int64
	}{
		{"monday", 1},
		{"tuesday", 0},
		{"wednesday", 1},
		{"thursday", 0},
		{"friday", 1},
		{"saturday", 1},
		{"sunday", 0},
		{"random", 0},
	}
	for _, test := range testCase {
		actual := timeTable.ShowDeadlineOn(test.Day)
		assert.Equal(t, test.Expected, actual)
	}
}

func TestIsEmpty(t *testing.T) {
	timeTable1 := TimeTable{
		Monday:    1,
		Tuesday:   0,
		Wednesday: 1,
		Thursday:  0,
		Friday:    1,
		Saturday:  1,
		Sunday:    0,
	}
	timeTable2 := TimeTable{
		Monday:    0,
		Tuesday:   0,
		Wednesday: 0,
		Thursday:  0,
		Friday:    0,
		Saturday:  0,
		Sunday:    0,
	}
	testCase := []struct {
		tt       TimeTable
		expected bool
	}{
		{timeTable1, false},
		{timeTable2, true},
	}
	for _, test := range testCase {
		actual := test.tt.IsEmpty()
		assert.Equal(t, test.expected, actual)
	}
}
