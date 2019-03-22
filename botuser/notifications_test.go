package botuser

import (
	"errors"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"golang.org/x/text/language"
)

func TestNotifyChannels(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		TeamID:       "comedian",
		UserID:       "TESTUSERID",
		Language:     "en_US",
		ReminderTime: int64(0),
	}

	testCases := []struct {
		time                       time.Time
		ListedChannels             []model.Channel
		ListedChannelsError        error
		Standupers                 []model.Standuper
		ListStandupersError        error
		ChannelStandupers          []model.Standuper
		ListChannelStandupersError error
		SelectedChannel            model.Channel
		SelectedChannelError       error
	}{
		{time.Date(2018, 3, 10, 10, 0, 0, 0, time.Local), []model.Channel{}, nil, []model.Standuper{}, nil, []model.Standuper{}, nil, model.Channel{}, nil},
		{time.Date(2018, 3, 12, 10, 0, 0, 0, time.Local), []model.Channel{}, errors.New("list channels"), []model.Standuper{}, nil, []model.Standuper{}, nil, model.Channel{}, nil},
		{time.Date(2018, 3, 12, 10, 0, 0, 0, time.Local), []model.Channel{{TeamID: "foo"}}, nil, []model.Standuper{}, nil, []model.Standuper{}, nil, model.Channel{}, nil},
		{time.Date(2018, 3, 12, 10, 0, 0, 0, time.Local), []model.Channel{{TeamID: "comedian", StandupTime: int64(0)}}, nil, []model.Standuper{}, nil, []model.Standuper{}, nil, model.Channel{}, nil},
		{time.Date(2018, 3, 12, 10, 0, 0, 0, time.Local), []model.Channel{{TeamID: "comedian", StandupTime: time.Date(2018, 3, 12, 10, 0, 0, 0, time.Local).Unix()}}, nil, []model.Standuper{}, errors.New("err"), []model.Standuper{}, errors.New("err"), model.Channel{}, nil},
	}

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			ListedChannels:             tt.ListedChannels,
			ListedChannelsError:        tt.ListedChannelsError,
			Standupers:                 tt.Standupers,
			ListStandupersError:        tt.ListStandupersError,
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
			SelectedChannel:            tt.SelectedChannel,
			SelectedChannelError:       tt.SelectedChannelError,
		})
		bot.NotifyChannels(tt.time)
	}
}

func TestSendWarning(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		TeamID:       "comedian",
		UserID:       "TESTUSERID",
		Language:     "en_US",
		ReminderTime: int64(0),
	}

	testCases := []struct {
		dry                 bool
		Standupers          []model.Standuper
		ListStandupersError error
		err                 error
	}{
		{false, []model.Standuper{}, errors.New("list standupers error"), errors.New("list standupers error")},
		{false, []model.Standuper{}, nil, nil},
		{false, []model.Standuper{{}}, nil, nil},
		{false, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, errors.New("Could not post message to a channel")},
		{true, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
	}

	for _, tt := range testCases {
		Dry = tt.dry
		bot := New(bundle, settings, MockedDB{
			Standupers:          tt.Standupers,
			ListStandupersError: tt.ListStandupersError,
		})
		err := bot.SendWarning("Foo")
		assert.Equal(t, tt.err, err)
	}
}

func TestSendWarningBundleFail(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	settings := model.BotSettings{
		TeamID:       "comedian",
		UserID:       "TESTUSERID",
		Language:     "en_US",
		ReminderTime: int64(0),
	}

	testCases := []struct {
		dry                 bool
		Standupers          []model.Standuper
		ListStandupersError error
		err                 error
	}{
		{false, []model.Standuper{}, errors.New("list standupers error"), errors.New("list standupers error")},
		{false, []model.Standuper{}, nil, nil},
		{false, []model.Standuper{{}}, nil, nil},
		{false, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, errors.New("Could not post message to a channel")},
		{true, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
	}

	for _, tt := range testCases {
		Dry = tt.dry
		bot := New(bundle, settings, MockedDB{
			Standupers:          tt.Standupers,
			ListStandupersError: tt.ListStandupersError,
		})
		err := bot.SendWarning("Foo")
		assert.Equal(t, tt.err, err)
	}
}

func TestSendChannelNotification(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		TeamID:       "comedian",
		UserID:       "TESTUSERID",
		Language:     "en_US",
		ReminderTime: int64(0),
	}

	testCases := []struct {
		dry                        bool
		ReminderRepeatsMax         int
		ChannelStandupers          []model.Standuper
		ListChannelStandupersError error
		err                        error
	}{
		{false, 0, []model.Standuper{}, errors.New("list standupers error"), errors.New("list standupers error")},
		{false, 0, []model.Standuper{}, nil, nil},
		{false, 0, []model.Standuper{{}}, nil, nil},
		{true, 0, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
		{true, 1, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
	}

	for _, tt := range testCases {
		Dry = tt.dry
		bot := New(bundle, settings, MockedDB{
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})
		bot.properties.ReminderRepeatsMax = tt.ReminderRepeatsMax
		err := bot.SendChannelNotification("Foo")
		assert.Equal(t, tt.err, err)
	}
}

func TestSendChannelNotificationBundleFailed(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	settings := model.BotSettings{
		TeamID:       "comedian",
		UserID:       "TESTUSERID",
		Language:     "en_US",
		ReminderTime: int64(0),
	}

	testCases := []struct {
		dry                        bool
		ReminderRepeatsMax         int
		ChannelStandupers          []model.Standuper
		ListChannelStandupersError error
		err                        error
	}{
		{false, 0, []model.Standuper{}, errors.New("list standupers error"), errors.New("list standupers error")},
		{false, 0, []model.Standuper{}, nil, nil},
		{false, 0, []model.Standuper{{}}, nil, nil},
		{true, 0, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
		{true, 1, []model.Standuper{{ChannelID: "Foo", SubmittedStandupToday: false}}, nil, nil},
	}

	for _, tt := range testCases {
		Dry = tt.dry
		bot := New(bundle, settings, MockedDB{
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})
		bot.properties.ReminderRepeatsMax = tt.ReminderRepeatsMax
		err := bot.SendChannelNotification("Foo")
		assert.Equal(t, tt.err, err)
	}
}
