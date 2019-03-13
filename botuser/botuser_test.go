package botuser

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

func TestNew(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(bundle, settings, db)
	assert.Equal(t, "TESTUSERID", bot.properties.UserID)

}

func TestAnalizeStandup(t *testing.T) {
	testCases := []struct {
		Message string
		Problem string
	}{
		{"", ":warning: No 'yesterday' related keywords detected! Please, use one of the following: 'yesterday' or weekdays such as 'friday' etc."},
		{"yesterday", ":warning: No 'today' related keywords detected! Please, use one of the following: 'today', 'going', 'plan'"},
		{"yesterday, today", ":warning: No 'problems' related keywords detected! Please, use one of the following: 'problem', 'difficult', 'stuck', 'question', 'issue'"},
		{"yesterday, today, problems", ""},
	}

	properties := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: properties,
	}

	for _, tt := range testCases {
		problem := bot.analizeStandup(tt.Message)
		assert.Equal(t, tt.Problem, problem)
	}

	testCasesErr := []struct {
		Message string
		Problem string
	}{
		{"", ""},
		{"yesterday", ""},
		{"yesterday, today", ""},
		{"yesterday, today, problems", ""},
	}

	wrongBundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err = wrongBundle.LoadMessageFile("active.en.toml")
	assert.Error(t, err)

	bot = &Bot{
		bundle:     wrongBundle,
		properties: properties,
	}

	for _, tt := range testCasesErr {
		problem := bot.analizeStandup(tt.Message)
		assert.Equal(t, tt.Problem, problem)
	}
}

func TestSuits(t *testing.T) {
	properties := model.BotSettings{
		TeamID:   "Foo",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: properties,
	}

	ok := bot.Suits("Foo")
	assert.Equal(t, true, ok)
}

func TestSetProperties(t *testing.T) {
	settings := model.BotSettings{
		TeamID:   "Foo",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: settings,
	}

	newSettings := model.BotSettings{
		TeamID:   "Bar",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot.SetProperties(newSettings)
	assert.Equal(t, "Bar", bot.properties.TeamID)
}
