package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestGetAccessLevel(t *testing.T) {
	r := SetUp()

	//create user with superadmin id
	superadmin, err := r.db.CreateUser(model.User{
		UserName: "sadmin",
		UserID:   "SuperAdminID",
		Role:     "",
	})
	assert.NoError(t, err)
	//create user with admin role
	admin, err := r.db.CreateUser(model.User{
		UserName: "admin",
		UserID:   "AdminID",
		Role:     "admin",
	})
	assert.NoError(t, err)
	//create user pm
	UserPm, err := r.db.CreateUser(model.User{
		UserName: "userpm",
		UserID:   "idpm",
		Role:     "pm",
	})
	assert.NoError(t, err)
	//create channel
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "chanid1",
	})
	assert.NoError(t, err)
	//create channel member with role pm
	PM, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        UserPm.UserID,
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)
	//create usual user
	User, err := r.db.CreateUser(model.User{
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
		actualLevel, err := r.getAccessLevel(test.UserID, test.ChannelID)
		assert.Equal(t, test.ExceptedAccessLevel, actualLevel)
		assert.Equal(t, test.ExceptedError, err)
	}
	//deletes users
	err = r.db.DeleteUser(superadmin.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(admin.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(UserPm.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(User.ID)
	assert.NoError(t, err)
	//delete channel members
	err = r.db.DeleteChannelMember(PM.UserID, PM.ChannelID)
	assert.NoError(t, err)
	//delete channel
	err = r.db.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
}
