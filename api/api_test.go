package api

import (
	"errors"
	"testing"

	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestGetAccessLevel(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	bot.CP.ManagerSlackUserID = "SuperAdminID"
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//create user with superadmin id
	superadmin, err := bot.DB.CreateUser(model.User{
		UserName: "sadmin",
		UserID:   "SuperAdminID",
		Role:     "",
	})
	assert.NoError(t, err)
	//create user with admin role
	admin, err := bot.DB.CreateUser(model.User{
		UserName: "admin",
		UserID:   "AdminID",
		Role:     "admin",
	})
	assert.NoError(t, err)
	//create user pm
	UserPm, err := bot.DB.CreateUser(model.User{
		UserName: "userpm",
		UserID:   "idpm",
		Role:     "pm",
	})
	assert.NoError(t, err)
	//create channel
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "chanid1",
	})
	assert.NoError(t, err)
	//create channel member with role pm
	PM, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        UserPm.UserID,
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)
	//create usual user
	User, err := bot.DB.CreateUser(model.User{
		UserID:   "userid",
		UserName: "username",
		Role:     "",
	})
	assert.NoError(t, err)

	testCase := []struct {
		UserID              string
		ChannelID           string
		ExceptedAccessLevel int
		ExceptedError       error
	}{
		{"random", "", 0, errors.New("sql: no rows in result set")},
		{superadmin.UserID, "", 1, nil},
		{admin.UserID, "", 2, nil},
		{PM.UserID, channel1.ChannelID, 3, nil},
		{User.UserID, "", 4, nil},
	}
	for _, test := range testCase {
		actualLevel, err := botAPI.getAccessLevel(test.UserID, test.ChannelID)
		assert.Equal(t, test.ExceptedAccessLevel, actualLevel)
		assert.Equal(t, test.ExceptedError, err)
	}
	//deletes users
	err = bot.DB.DeleteUser(superadmin.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(admin.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(UserPm.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(User.ID)
	assert.NoError(t, err)
	//delete channel members
	err = bot.DB.DeleteChannelMember(PM.UserID, PM.ChannelID)
	assert.NoError(t, err)
	//delete channel
	err = bot.DB.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
}
