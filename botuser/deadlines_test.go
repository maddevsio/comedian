package botuser

import (
	"errors"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/bouk/monkey"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"golang.org/x/text/language"
)

func TestAddTime(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		accessLevel                int
		params                     string
		SelectedChannel            model.Channel
		UpdatedChannel             model.Channel
		ChannelStandupers          []model.Standuper
		SelectedChannelError       error
		UpdatedChannelError        error
		ListChannelStandupersError error
	}{
		{"Access Denied! You need to be at least PM in this project to use this command!", 4, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"Unable to recognize time for a deadline", 3, "1230_fds+fs%", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"Unable to recognize time for a deadline", 3, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"could not recognize channel, please add me to the channel and try again", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, errors.New("select channel"), nil, nil},
		{"could not set channel deadline", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, errors.New("update channel"), nil},
		{"<!date^1514894400^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, errors.New("list standupers")},
		{"<!date^1514894400^Standup time set at {time}|Standup time set at 12:00>", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{model.Standuper{ID: int64(1)}}, nil, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:            tt.SelectedChannel,
			UpdatedChannel:             tt.UpdatedChannel,
			ChannelStandupers:          tt.ChannelStandupers,
			SelectedChannelError:       tt.SelectedChannelError,
			UpdatedChannelError:        tt.UpdatedChannelError,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		text := bot.addTime(tt.accessLevel, "Foo", tt.params)
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestAddTimeFailBundle(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		accessLevel                int
		params                     string
		SelectedChannel            model.Channel
		UpdatedChannel             model.Channel
		ChannelStandupers          []model.Standuper
		SelectedChannelError       error
		UpdatedChannelError        error
		ListChannelStandupersError error
	}{
		{"", 4, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"Unable to recognize time for a deadline", 3, "1230_fds+fs%", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"Unable to recognize time for a deadline", 3, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"could not recognize channel, please add me to the channel and try again", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, errors.New("select channel"), nil, nil},
		{"could not set channel deadline", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, errors.New("update channel"), nil},
		{"", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, errors.New("list standupers")},
		{"", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{model.Standuper{ID: int64(1)}}, nil, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:            tt.SelectedChannel,
			UpdatedChannel:             tt.UpdatedChannel,
			ChannelStandupers:          tt.ChannelStandupers,
			SelectedChannelError:       tt.SelectedChannelError,
			UpdatedChannelError:        tt.UpdatedChannelError,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		text := bot.addTime(tt.accessLevel, "Foo", tt.params)
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestRemoveTime(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		accessLevel                int
		params                     string
		SelectedChannel            model.Channel
		UpdatedChannel             model.Channel
		ChannelStandupers          []model.Standuper
		SelectedChannelError       error
		UpdatedChannelError        error
		ListChannelStandupersError error
	}{
		{"Access Denied! You need to be at least PM in this project to use this command!", 4, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"could not recognize channel, please add me to the channel and try again", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, errors.New("select channel"), nil, nil},
		{"could not remove channel deadline", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, errors.New("update channel"), nil},
		{"Standup time for channel deleted.", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, errors.New("list standupers")},
		{"Standup time for this channel removed, but there are people marked as a standuper.", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{model.Standuper{ID: int64(1)}}, nil, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:            tt.SelectedChannel,
			UpdatedChannel:             tt.UpdatedChannel,
			ChannelStandupers:          tt.ChannelStandupers,
			SelectedChannelError:       tt.SelectedChannelError,
			UpdatedChannelError:        tt.UpdatedChannelError,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		text := bot.removeTime(tt.accessLevel, "Foo")
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestRemoveTimeFailBundle(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		accessLevel                int
		params                     string
		SelectedChannel            model.Channel
		UpdatedChannel             model.Channel
		ChannelStandupers          []model.Standuper
		SelectedChannelError       error
		UpdatedChannelError        error
		ListChannelStandupersError error
	}{
		{"", 4, "", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, nil},
		{"could not recognize channel, please add me to the channel and try again", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, errors.New("select channel"), nil, nil},
		{"could not remove channel deadline", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, errors.New("update channel"), nil},
		{"", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{}, nil, nil, errors.New("list standupers")},
		{"", 3, "12:00", model.Channel{}, model.Channel{}, []model.Standuper{model.Standuper{ID: int64(1)}}, nil, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:            tt.SelectedChannel,
			UpdatedChannel:             tt.UpdatedChannel,
			ChannelStandupers:          tt.ChannelStandupers,
			SelectedChannelError:       tt.SelectedChannelError,
			UpdatedChannelError:        tt.UpdatedChannelError,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		text := bot.removeTime(tt.accessLevel, "Foo")
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestShowTime(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText         string
		SelectedChannel      model.Channel
		SelectedChannelError error
	}{
		{"No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!", model.Channel{StandupTime: ""}, nil},
		{"No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!", model.Channel{}, errors.New("select channel")},
		{"Standup time is 12:00", model.Channel{StandupTime: "12:00"}, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:      tt.SelectedChannel,
			SelectedChannelError: tt.SelectedChannelError,
		})

		text := bot.showTime("Foo")
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestShowTimeFailBundle(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText         string
		SelectedChannel      model.Channel
		SelectedChannelError error
	}{
		{"", model.Channel{StandupTime: ""}, nil},
		{"", model.Channel{}, errors.New("select channel")},
		{"", model.Channel{StandupTime: ""}, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedChannel:      tt.SelectedChannel,
			SelectedChannelError: tt.SelectedChannelError,
		})

		text := bot.showTime("Foo")
		assert.Equal(t, tt.expectedText, text)
	}
}
