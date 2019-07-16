package botuser

import (
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFindChannelNonReporters(t *testing.T) {
	t.Skip("Need to fix test and only then run")
	nonReportes, err := bot.findChannelNonReporters(model.Channel{
		ChannelID: "CHAN123",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nonReportes))

	standuper, err := bot.db.CreateStanduper(model.Standuper{
		Created:   time.Now().UTC(),
		TeamID:    "testTeam",
		ChannelID: "CHAN123",
		UserID:    "Foo",
	})

	assert.NoError(t, err)

	nonReportes, err = bot.findChannelNonReporters(model.Channel{
		ChannelID: "CHAN123",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nonReportes))
	assert.Equal(t, "<@"+standuper.UserID+">", nonReportes[0])

	standup, err := bot.db.CreateStandup(model.Standup{
		Created:   time.Now().UTC(),
		Modified:  time.Now().UTC(),
		TeamID:    "testTeam",
		ChannelID: "CHAN123",
		UserID:    "Foo",
		MessageTS: "12345",
	})
	assert.NoError(t, err)

	nonReportes, err = bot.findChannelNonReporters(model.Channel{
		ChannelID: "CHAN123",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(nonReportes))

	assert.NoError(t, bot.db.DeleteStanduper(standuper.ID))
	assert.NoError(t, bot.db.DeleteStandup(standup.ID))
}
