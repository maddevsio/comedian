package botuser

import (
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

var bot = setupBot()

func setupBot() *Bot {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	if err != nil {
		return nil
	}

	db, err := storage.New(config.DatabaseURL)
	if err != nil {
		return nil
	}

	settings := &model.BotSettings{
		TeamID:      "testTeam",
		AccessToken: "foo",
	}

	bot := New(config, bundle, settings, db)

	bot.db.CreateChannel(model.Channel{
		TeamID:      "testTeam",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
	})

	bot.db.CreateChannel(model.Channel{
		TeamID:      "testTeam",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		StandupTime: "12:00",
	})

	return bot
}

func TestAnalizeStandup(t *testing.T) {

	errors := bot.analizeStandup("yesterday, today, issues")
	assert.Equal(t, "", errors)

	errors = bot.analizeStandup("wrong standup")
	assert.Equal(t, "- no 'yesterday' keywords detected: yesterday, friday, вчера, пятниц, - no 'today' keywords detected: today, сегодня, - no 'problems' keywords detected: issue, мешает", errors)
}
