package comedianbot

import (
	"errors"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

func TestNew(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)
}

func TestBots(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}

	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.New(config)
	assert.NoError(t, err)

	comedian := New(bundle, db)
	assert.NotEqual(t, nil, comedian)

	botuser.Dry = true

	go comedian.StartBots()

	botSettings := model.BotSettings{
		TeamID: "testTeam",
	}

	assert.Equal(t, 0, len(comedian.bots))

	comedian.AddBot(botSettings)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(comedian.bots))

	_, err = comedian.SelectBot("randombot")
	assert.Error(t, err)

	err = comedian.SetBots()
	assert.NoError(t, err)
}

func TestHandleEvent(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}

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

	comedian.AddBot(model.BotSettings{
		TeamID: "testTeam",
	})
	assert.Equal(t, 1, len(comedian.bots))

	err = comedian.HandleEvent(model.ServiceEvent{
		TeamName: "testTeam",
	})
	assert.Equal(t, nil, err)
}
