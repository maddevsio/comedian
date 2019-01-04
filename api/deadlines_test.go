package api

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestAddTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel without members
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChan1",
		ChannelID:   "testChanId1",
	})
	assert.NoError(t, err)
	//creates channel with members
	channel2, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChan2",
		ChannelID:   "testChanId2",
	})
	assert.NoError(t, err)
	//creates channel members
	ChanMem1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "userId1",
		ChannelID:     channel2.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//parse 10:30 text to int to use it in testCases
	tm, err := botAPI.ParseTimeTextToInt("10:30")
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
		actual := botAPI.addTime(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	err = bot.DB.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
	//delete channel member
	err = bot.DB.DeleteChannelMember(ChanMem1.UserID, ChanMem1.ChannelID)
	assert.NoError(t, err)
}

func TestRemoveTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel with members
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: int64(100),
	})
	assert.NoError(t, err)
	//adds channel members
	chanMemb1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates channel without members
	channel2, err := bot.DB.CreateChannel(model.Channel{
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
		actual := botAPI.removeTime(test.accessL, test.chanID)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	assert.NoError(t, bot.DB.DeleteChannel(channel1.ID))
	assert.NoError(t, bot.DB.DeleteChannel(channel2.ID))
	//deletes channel members
	err = bot.DB.DeleteChannelMember(chanMemb1.UserID, chanMemb1.ChannelID)
	assert.NoError(t, err)
}

func TestShowTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//create a channel with standuptime
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
	})
	assert.NoError(t, err)
	//set a standuptime for channel
	err = bot.DB.CreateStandupTime(12345, channel1.ChannelID)
	assert.NoError(t, err)
	//create channel without standuptime
	channel2, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "testChannelId2",
	})
	assert.NoError(t, err)
	testCase := []struct {
		channelID string
		expected  string
	}{
		{channel1.ChannelID, "<!date^12345^Standup time is {time}|Standup time set at 12:00>"},
		{channel2.ChannelID, "No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!"},
		{"doesntExistedChan", "No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!"},
	}
	for _, test := range testCase {
		actual := botAPI.showTime(test.channelID)
		assert.Equal(t, test.expected, actual)
	}
	//delete channels
	assert.NoError(t, bot.DB.DeleteChannel(channel1.ID))
	assert.NoError(t, bot.DB.DeleteChannel(channel2.ID))
}

func TestParseTimeTextToInt(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	d := time.Date(2018, 10, 4, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases := []struct {
		timeText string
		time     int64
		err      error
	}{
		{"0", 0, nil},
		{"10:00", 1538625600, nil},
		{"xx:00", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
		{"00:xx", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
		{"00:62", 0, errors.New("Wrong time! Please, check the time format and try again!")},
		{"10am", 0, errors.New("Seems like you used short time format, please, use 24:00 hour format instead!")},
		{"20", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
	}
	for _, tt := range testCases {
		_, err := botAPI.ParseTimeTextToInt(tt.timeText)
		assert.Equal(t, tt.err, err)
		//assert.Equal(t, tt.time, time)
	}
}
