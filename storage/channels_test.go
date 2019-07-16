package storage

import (
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateChannel(t *testing.T) {
	_, err := db.CreateChannel(model.Channel{})
	assert.Error(t, err)

	ch, err := db.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
		StandupTime: "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", ch.TeamID)

	assert.NoError(t, db.DeleteChannel(ch.ID))
}

func TestGetChannels(t *testing.T) {
	ch, err := db.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
		StandupTime: "",
	})
	assert.NoError(t, err)

	_, err = db.ListChannels()
	assert.NoError(t, err)

	_, err = db.SelectChannel("")
	assert.Error(t, err)

	_, err = db.SelectChannel("bar12")
	assert.NoError(t, err)

	_, err = db.GetChannel(int64(0))
	assert.Error(t, err)

	_, err = db.GetChannel(ch.ID)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteChannel(ch.ID))
}

func TestUpdateChannels(t *testing.T) {
	ch, err := db.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", ch.StandupTime)

	ch.StandupTime = "10:00"
	ch, err = db.UpdateChannel(ch)
	assert.NoError(t, err)
	assert.Equal(t, "10:00", ch.StandupTime)

	assert.NoError(t, db.DeleteChannel(ch.ID))
}
