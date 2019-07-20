package comedianbot

import (
	"encoding/json"
	"errors"
	"github.com/maddevsio/comedian/botuser"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"testing"
)

func TestNew(t *testing.T) {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)
}

func TestBots(t *testing.T) {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)

	go comedian.StartBots()

	botSettings := &model.BotSettings{
		TeamID: "testTeam",
	}

	assert.Equal(t, 0, len(comedian.bots))

	bot := botuser.New(nil, nil, botSettings, nil)
	comedian.AddBot(bot)
	assert.Equal(t, 1, len(comedian.bots))

	_, err = comedian.SelectBot("randombot")
	assert.Error(t, err)

	_, err = comedian.SelectBot("testTeam")
	assert.NoError(t, err)

	bs := model.BotSettings{
		NotifierInterval:    30,
		Language:            "en_US",
		ReminderRepeatsMax:  3,
		ReminderTime:        int64(10),
		AccessToken:         "token",
		UserID:              "userID",
		TeamID:              "teamID",
		TeamName:            "teamName",
		ReportingChannel:    "",
		ReportingTime:       "9:00",
		IndividualReportsOn: false,
	}

	newBot, err := db.CreateBotSettings(bs)
	assert.NoError(t, err)

	err = comedian.SetBots()
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteBotSettingsByID(newBot.ID))
}

func TestHandleEvent(t *testing.T) {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)

	go comedian.StartBots()

	err = comedian.HandleEvent(model.ServiceEvent{
		TeamName: "test",
	})
	assert.Equal(t, errors.New("no bot found to implement the request"), err)

	botSettings := &model.BotSettings{
		TeamID:      "testTeam",
		AccessToken: "foo",
	}

	assert.Equal(t, 0, len(comedian.bots))

	bot := botuser.New(nil, nil, botSettings, nil)

	comedian.AddBot(bot)
	assert.Equal(t, 1, len(comedian.bots))

	err = comedian.HandleEvent(model.ServiceEvent{
		TeamName: "testTeam",
	})
	assert.Equal(t, "Wrong access token", err.Error())

	Dry = true

	err = comedian.HandleEvent(model.ServiceEvent{
		TeamName:    "testTeam",
		AccessToken: "foo",
	})

	assert.NoError(t, err)
}

func TestHandleCallbackEvent(t *testing.T) {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)

	go comedian.StartBots()

	bot := botuser.New(nil, bundle, &model.BotSettings{
		TeamID:      "testTeam",
		AccessToken: "foo",
	}, nil)

	comedian.AddBot(bot)

	assert.Equal(t, 1, len(comedian.bots))

	jsonStr := json.RawMessage(`{"type": "random"}`)

	err = comedian.HandleCallbackEvent(slackevents.EventsAPICallbackEvent{
		TeamID:     "testWrongTeam",
		InnerEvent: &jsonStr,
	})
	assert.Equal(t, "no bot found to implement the request", err.Error())

	err = comedian.HandleCallbackEvent(slackevents.EventsAPICallbackEvent{
		TeamID:     "testTeam",
		InnerEvent: &jsonStr,
	})
	assert.NoError(t, err)
}
