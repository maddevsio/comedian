package botuser

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImplementDeadlineCommands(t *testing.T) {

	resp := bot.ImplementCommands(slack.SlashCommand{
		Command:     "edit_deadline",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "12:00",
	})
	assert.Equal(t, "Updated standup deadline to 12:00", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "show_deadline",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "Standup deadline set to 12:00", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "edit_deadline",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "Standup deadline removed", resp)

	resp = bot.ImplementCommands(slack.SlashCommand{
		Command:     "show_deadline",
		TeamID:      "testTeam",
		UserID:      "foo123",
		ChannelID:   "CHAN123",
		ChannelName: "ChannelWithNoDeadline",
		Text:        "",
	})
	assert.Equal(t, "Standup deadline is not set", resp)

}
