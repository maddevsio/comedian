package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func TestAddTime(t *testing.T) {
	r := SetUp()

	//creates channel without members
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChan1",
		ChannelID:   "testChanId1",
	})
	assert.NoError(t, err)
	//creates channel with members
	channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChan2",
		ChannelID:   "testChanId2",
	})
	assert.NoError(t, err)
	//creates channel members
	ChanMem1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "userId1",
		ChannelID:     channel2.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//parse 10:30 text to int to use it in testCases
	tm, err := utils.ParseTimeTextToInt("10:30")
	assert.NoError(t, err)
	testCase := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, "", "", "Access Denied! You need to be at least PM in this project to use this command!"},
		{3, channel1.ChannelID, "10:30", fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>", tm)},
		{3, channel2.ChannelID, "10:30", fmt.Sprintf("<!date^%v^Standup time set at {time}|Standup time set at 12:00>", tm)},
		{3, "random", "10:30", fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>", tm)},
		{3, "", "25:25", "Wrong time! Please, check the time format and try again!"},
	}
	for _, test := range testCase {
		actual := r.addTime(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	err = r.db.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
	//delete channel member
	err = r.db.DeleteChannelMember(ChanMem1.UserID, ChanMem1.ChannelID)
	assert.NoError(t, err)
}

func TestRemoveTime(t *testing.T) {
	r := SetUp()
	//creates channel with members
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: int64(100),
	})
	assert.NoError(t, err)
	//adds channel members
	chanMemb1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates channel without members
	channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "testChannelId2",
		StandupTime: int64(100),
	})
	assert.NoError(t, err)

	testCase := []struct {
		accessL  int
		chanID   string
		expected string
	}{
		{4, channel1.ChannelID, "Access Denied! You need to be at least PM in this project to use this command!"},
		{2, channel1.ChannelID, "standup time for this channel removed, but there are people marked as a standuper."},
		{3, channel2.ChannelID, "standup time for channel deleted"},
	}
	for _, test := range testCase {
		actual := r.removeTime(test.accessL, test.chanID)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	assert.NoError(t, r.db.DeleteChannel(channel1.ID))
	assert.NoError(t, r.db.DeleteChannel(channel2.ID))
	//deletes channel members
	err = r.db.DeleteChannelMember(chanMemb1.UserID, chanMemb1.ChannelID)
	assert.NoError(t, err)
}

func TestShowTime(t *testing.T) {
	r := SetUp()
	//create a channel with standuptime
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
	})
	assert.NoError(t, err)
	//set a standuptime for channel
	err = r.db.CreateStandupTime(12345, channel1.ChannelID)
	assert.NoError(t, err)
	//create channel without standuptime
	channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "testChannelId2",
	})
	assert.NoError(t, err)
	testCase := []struct {
		channelID string
		expected  string
	}{
		{channel1.ChannelID, "<!date^12345^Standup time is {time}|Standup time set at 12:00>"},
		{channel2.ChannelID, "No standup time set for this channel yet! Please, add a standup time using `/standup_time_set` command!"},
		{"doesntExistedChan", "No standup time set for this channel yet! Please, add a standup time using `/standup_time_set` command!"},
	}
	for _, test := range testCase {
		actual := r.showTime(test.channelID)
		assert.Equal(t, test.expected, actual)
	}
	//delete channels
	assert.NoError(t, r.db.DeleteChannel(channel1.ID))
	assert.NoError(t, r.db.DeleteChannel(channel2.ID))
}
