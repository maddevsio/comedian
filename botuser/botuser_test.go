package botuser

import (
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"testing"
)

var bot = setupBot()

func setupBot() *Bot {
	bundle := i18n.NewBundle(language.English)

	config, err := config.Get()
	if err != nil {
		return nil
	}

	db, err := storage.New(config)
	if err != nil {
		return nil
	}

	settings := model.BotSettings{
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

	errors := bot.analizeStandup("yesterday, today, blockers")
	assert.Equal(t, "", errors)

	errors = bot.analizeStandup("wrong standup")
	assert.Equal(t, "- no 'yesterday' keywords detected: yesterday, friday, monday, tuesday, wednesday, thursday, saturday, sunday, вчера, пятниц, понедельник, вторник, сред, четверг, суббот, воскресенье, - no 'today' keywords detected: today, сегодня, - no 'problems' keywords detected: problem, difficult, issue, block, проблем, мешает", errors)
}

func TestHandleJoinNewUser(t *testing.T) {

	user, err := bot.HandleJoinNewUser(slack.User{
		IsBot: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, model.User{}, user)

	user, err = bot.HandleJoinNewUser(slack.User{
		Name: "slackbot",
	})
	assert.NoError(t, err)
	assert.Equal(t, model.User{}, user)

	user, err = bot.HandleJoinNewUser(slack.User{
		Name:     "Thor",
		ID:       "Thor123",
		RealName: "Loki",
		TZ:       "",
		TZOffset: 0,
	})
	assert.Error(t, err)
	assert.Equal(t, "team ID cannot be empty", err.Error())

	user, err = bot.HandleJoinNewUser(slack.User{
		TeamID:   "testTeam",
		Name:     "Thor",
		ID:       "Thor123",
		RealName: "Loki",
		TZ:       "",
		TZOffset: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Thor", user.UserName)

	user, err = bot.HandleJoinNewUser(slack.User{
		Name:     "Thor",
		ID:       "Thor123",
		RealName: "Loki",
		TZ:       "",
		TZOffset: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Thor", user.UserName)

	assert.NoError(t, bot.db.DeleteUser(user.ID))

}

func TestImplementCommands(t *testing.T) {

	resp := bot.ImplementCommands(slack.SlashCommand{
		Command:     "join",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "Welcome to the standup team, no standup deadline has been setup yet", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "join",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		Text:        "",
	})
	assert.Equal(t, "Welcome to the standup team, please, submit your standups no later than 12:00", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "join",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "Channel",
		Text:        "",
	})
	assert.Equal(t, "You are already a part of standup team", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "show",
		TeamID:      "testTeam",
		UserID:      "foo",
		ChannelID:   "CHAN123",
		ChannelName: "Channel",
		Text:        "",
	})
	assert.NotEqual(t, "", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "Channel",
		Text:        "",
	})
	assert.Equal(t, "You no longer have to submit standups, thanks for all your standups and messages", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "Channel",
		Text:        "",
	})
	assert.Equal(t, "You do not standup yet", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		Text:        "",
	})
	assert.Equal(t, "You no longer have to submit standups, thanks for all your standups and messages", resp)

}
