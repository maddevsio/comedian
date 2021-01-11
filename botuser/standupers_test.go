package botuser

import (
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImplementStandupersCommands(t *testing.T) {

	resp := bot.ImplementCommands(slack.SlashCommand{
		Command:     "/start",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "Welcome to the standup team, no standup deadline has been setup yet", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/start",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		Text:        "",
	})
	assert.Equal(t, "Welcome to the standup team, please, submit your standups no later than 12:00", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/start",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "You are already a part of standup team", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/show",
		TeamID:      "testTeam",
		UserID:      "foo",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.NotEqual(t, "", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "You no longer have to submit standups, thanks for all your standups and messages", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "You do not standup yet", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/quit",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		Text:        "",
	})
	assert.Equal(t, "You no longer have to submit standups, thanks for all your standups and messages", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "/show",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN321",
		ChannelName: "ChannelWithDeadline",
		Text:        "",
	})
	assert.NotEqual(t, "", resp)
}
