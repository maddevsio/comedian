package storage

import (
	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestCreateChannel(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	_, err = mysql.CreateChannel(model.Channel{})
	assert.Error(t, err)

	ch, err := mysql.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", ch.TeamID)

	assert.NoError(t, mysql.DeleteChannel(ch.ID))
}

func TestGetChannels(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	ch, err := mysql.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)

	_, err = mysql.ListChannels()
	assert.NoError(t, err)

	_, err = mysql.SelectChannel("")
	assert.Error(t, err)

	_, err = mysql.SelectChannel("bar12")
	assert.NoError(t, err)

	_, err = mysql.GetChannel(int64(0))
	assert.Error(t, err)

	_, err = mysql.GetChannel(ch.ID)
	assert.NoError(t, err)

	assert.NoError(t, mysql.DeleteChannel(ch.ID))
}

func TestUpdateChannels(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	ch, err := mysql.CreateChannel(model.Channel{
		TeamID:      "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", ch.StandupTime)

	ch.StandupTime = "10:00"
	ch, err = mysql.UpdateChannel(ch)
	assert.NoError(t, err)
	assert.Equal(t, "10:00", ch.StandupTime)

	assert.NoError(t, mysql.DeleteChannel(ch.ID))
}
