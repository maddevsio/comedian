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
	assert.Equal(t, int64(0), ch.StandupTime)

	ch.StandupTime = int64(1)
	ch, err = mysql.UpdateChannel(ch)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), ch.StandupTime)

	assert.NoError(t, mysql.DeleteChannel(ch.ID))
}

func TestListChannelsByTeamID(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	teamID := "team1"
	var expectedListOfChannels []model.Channel

	channel1, err := mysql.CreateChannel(model.Channel{TeamID: teamID, ChannelID: "cid1", ChannelName: "ch1"})
	assert.NoError(t, err)
	channel2, err := mysql.CreateChannel(model.Channel{TeamID: teamID, ChannelID: "cid2", ChannelName: "ch2"})
	assert.NoError(t, err)
	randomChannel, err := mysql.CreateChannel(model.Channel{TeamID: "random", ChannelID: "random", ChannelName: "random"})

	expectedListOfChannels = append(expectedListOfChannels, channel1, channel2)

	actualListOfChannels, err := mysql.ListChannelsByTeamID(teamID)
	assert.NoError(t, err)
	assert.Equal(t, expectedListOfChannels, actualListOfChannels)

	err = mysql.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = mysql.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
	err = mysql.DeleteChannel(randomChannel.ID)
	assert.NoError(t, err)
}
