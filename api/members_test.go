package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestAddCommand(t *testing.T) {
	r := SetUp()

	user, err := r.db.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	testCases := []struct {
		accessLevel int
		channelID   string
		params      string
		output      string
	}{
		{1, channel.ChannelID, "<@userID|testUser>", "Members are assigned: <@userID|testUser>\n"},
		{4, channel.ChannelID, "<@userID|testUser>", "Access Denied! You need to be at least PM in this project to use this command!"},
		{2, channel.ChannelID, "<@userID|testUser> / admin", "Users are assigned as admins: <@userID|testUser>\n"},
		{3, channel.ChannelID, "<@userID|testUser> / admin", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{2, channel.ChannelID, "<@userID|testUser> / pm", "Users are assigned as PMs: <@userID|testUser>\n"},
		{3, channel.ChannelID, "<@userID|testUser> / pm", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{2, channel.ChannelID, "<@userID|testUser> / wrongRole", "Please, check correct role name (admin, developer, pm)"},
	}

	for _, tt := range testCases {
		result := r.addCommand(tt.accessLevel, tt.channelID, tt.params)
		assert.Equal(t, tt.output, result)

		members, err := r.db.ListAllChannelMembers()
		assert.NoError(t, err)
		for _, m := range members {
			assert.NoError(t, r.db.DeleteChannelMember(m.UserID, m.ChannelID))
		}
	}

	assert.NoError(t, r.db.DeleteChannel(channel.ID))
	assert.NoError(t, r.db.DeleteUser(user.ID))
}

func TestListCommand(t *testing.T) {
	//modify test to cover more cases: no users, etc.
	r := SetUp()

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	user, err := r.db.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "",
	})
	assert.NoError(t, err)

	admin, err := r.db.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "admin",
	})
	assert.NoError(t, err)

	memberPM, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        user.UserID,
		ChannelID:     channel.ChannelID,
		RoleInChannel: "pm",
	})

	memberDeveloper, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        user.UserID,
		ChannelID:     channel.ChannelID,
		RoleInChannel: "developer",
	})

	testCases := []struct {
		params string
		output string
	}{
		{"", "Standupers in this channel: <@userID>"},
		{"admin", "Admins in this workspace: <@testUser>"},
		{"developer", "Standupers in this channel: <@userID>"},
		{"pm", "PMs in this channel: <@userID>"},
		{"randomRole", "Please, check correct role name (admin, developer, pm)"},
	}

	for _, tt := range testCases {
		result := r.listCommand(channel.ChannelID, tt.params)
		assert.Equal(t, tt.output, result)
	}

	assert.NoError(t, r.db.DeleteChannel(channel.ID))
	assert.NoError(t, r.db.DeleteUser(user.ID))
	assert.NoError(t, r.db.DeleteUser(admin.ID))
	assert.NoError(t, r.db.DeleteChannelMember(memberPM.UserID, memberPM.ChannelID))
	assert.NoError(t, r.db.DeleteChannelMember(memberDeveloper.UserID, memberDeveloper.ChannelID))
}

func TestDeleteCommand(t *testing.T) {
	r := SetUp()

	testCase := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, "chan1", "<@id|name> / admin", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{4, "chan1", "<@id|name> / pm", "Access Denied! You need to be at least PM in this project to use this command!"},
		{4, "chan1", "<@id|name> / random", "Please, check correct role name (admin, developer, pm)"},
		{4, "chan1", "<@id|name>", "Access Denied! You need to be at least PM in this project to use this command!"},
	}
	for _, test := range testCase {
		actual := r.deleteCommand(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
}

func TestAddMembers(t *testing.T) {
	r := SetUp()

	//creates channel member with role pm
	_, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserId1",
		ChannelID:     "chan1",
		RoleInChannel: "pm",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates channel member with role developer
	_, err = r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserId2",
		ChannelID:     "chan1",
		RoleInChannel: "dev",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	testCase := []struct {
		Users         []string
		RoleInChannel string
		Expected      string
	}{
		//existed channel member with role pm
		{[]string{"<@testUserId1|testUserName1>"}, "pm", "Users already have roles: <@testUserId1|testUserName1>\n"},
		//existed channel member with role dev
		{[]string{"<@testUserId2|testUserName2>"}, "dev", "Members already have roles: <@testUserId2|testUserName2>\n"},
		//doesn't existed member with role pm
		{[]string{"<@testUserId3|testUserName3>"}, "pm", "Users are assigned as PMs: <@testUserId3|testUserName3>\n"},
		//two doesn't existed members with role pm
		{[]string{"<@testUserId4|testUserName4>", "<@testUserId5|testUserName5>"}, "pm", "Users are assigned as PMs: <@testUserId4|testUserName4><@testUserId5|testUserName5>\n"},
		//doesn't existed member with role dev
		{[]string{"<@testUserId6|testUserName6>"}, "dev", "Members are assigned: <@testUserId6|testUserName6>\n"},
		//wrong parameters
		{[]string{"user1"}, "pm", "Could not assign users as PMs: user1\n"},
		{[]string{"user1"}, "", "Could not assign members: user1\n"},
		{[]string{"user1", "<>"}, "", "Could not assign members: user1<>\n"},
	}
	for _, test := range testCase {
		actual := r.addMembers(test.Users, test.RoleInChannel, "chan1")
		assert.Equal(t, test.Expected, actual)
	}
	//deletes channelMembers
	//6 members will be created
	for i := 1; i <= 6; i++ {
		err = r.db.DeleteChannelMember(fmt.Sprintf("testUserId%v", i), "chan1")
		assert.NoError(t, err)
	}
}

func TestAddAdmins(t *testing.T) {
	r := SetUp()
	//not admin user
	User1, err := r.db.CreateUser(model.User{
		UserName: "UserName1",
		UserID:   "uid1",
		Role:     "",
	})
	assert.NoError(t, err)
	//admin user
	UserAdmin, err := r.db.CreateUser(model.User{
		UserName: "Admin1",
		UserID:   "uid2",
		Role:     "admin",
	})
	assert.NoError(t, err)

	testCase := []struct {
		Users    []string
		Expected string
	}{
		//wrong format
		{[]string{"user"}, "Could not assign users as admins: user\n"},
		//doesn't existed user
		{[]string{"<@User|username>"}, "Could not assign users as admins: <@User|username>\n"},
		//existed user User1, but his role not admin
		{[]string{"<@" + User1.UserID + "|" + User1.UserName + ">"}, "Users are assigned as admins: <@uid1|UserName1>\n"},
		{[]string{"<@" + UserAdmin.UserID + "|" + UserAdmin.UserName + ">"}, "Users were already assigned as admins: <@uid2|Admin1>\n"},
	}
	for _, test := range testCase {
		actual := r.addAdmins(test.Users)
		assert.Equal(t, test.Expected, actual)
	}
	//delete users
	err = r.db.DeleteUser(User1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(UserAdmin.ID)
	assert.NoError(t, err)
}

func TestListMembers(t *testing.T) {
	r := SetUp()

	//creates channel with pm members
	Channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName1",
		ChannelID:   "cid1",
	})
	assert.NoError(t, err)
	//creates channel Channel1 members PMs
	ChanMember1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     Channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)
	ChanMember2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     Channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)

	//creates channel with members
	Channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName2",
		ChannelID:   "cid2",
	})
	assert.NoError(t, err)
	//creates channel members
	ChanMember3, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     Channel2.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	ChanMember4, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid4",
		ChannelID:     Channel2.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)

	testCase := []struct {
		Channel  string
		Role     string
		Expected string
	}{
		{"channel", "pm", "No PMs in this channel! To add one, please, use `/add` slash command"},
		{Channel1.ChannelID, "pm", "PMs in this channel: <@uid1>, <@uid2>"},
		{"channel", "", "No standupers in this channel! To add one, please, use `/add` slash command"},
		{Channel2.ChannelID, "", "Standupers in this channel: <@uid3>, <@uid4>"},
	}
	for _, test := range testCase {
		actual := r.listMembers(test.Channel, test.Role)
		assert.Equal(t, test.Expected, actual)
	}
	//delete channel members
	err = r.db.DeleteChannelMember(ChanMember1.UserID, ChanMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(ChanMember2.UserID, ChanMember2.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(ChanMember3.UserID, ChanMember3.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(ChanMember4.UserID, ChanMember4.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = r.db.DeleteChannel(Channel1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteChannel(Channel2.ID)
	assert.NoError(t, err)
}

func TestListAdmins(t *testing.T) {
	r := SetUp()

	//no admins in db
	actual := r.listAdmins()
	assert.Equal(t, "No admins in this workspace! To add one, please, use `/add` slash command", actual)

	//creates users admins
	Admin1, err := r.db.CreateUser(model.User{
		UserName: "Username1",
		UserID:   "uid1",
		Role:     "admin",
	})
	assert.NoError(t, err)
	Admin2, err := r.db.CreateUser(model.User{
		UserName: "Username2",
		UserID:   "uid2",
		Role:     "admin",
	})
	assert.NoError(t, err)

	actual = r.listAdmins()
	assert.Equal(t, "Admins in this workspace: <@Username1>, <@Username2>", actual)

	//delete users
	err = r.db.DeleteUser(Admin1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(Admin2.ID)
	assert.NoError(t, err)
}

func TestDeleteMembers(t *testing.T) {
	r := SetUp()

	//create channel
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "cid1",
	})
	assert.NoError(t, err)
	//creates channel members
	channelMember1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)

	testCase := []struct {
		members   []string
		channelID string
		expected  string
	}{
		//wrong format
		{[]string{"user"}, "", "Could not remove the following members: user\n"},
		//doesn't existed channel members
		{[]string{"<@userid1|username1>", "<@userid2|username2>"}, channel1.ChannelID, "Could not remove the following members: <@userid1|username1><@userid2|username2>\n"},
		{[]string{"<@" + channelMember1.UserID + "|username"}, channel1.ChannelID, "The following members were removed: <@uid1|username\n"},
	}
	for _, test := range testCase {
		actual := r.deleteMembers(test.members, test.channelID)
		assert.Equal(t, test.expected, actual)
	}
	//delete channel members
	err = r.db.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channel
	err = r.db.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
}

func TestDeleteAdmins(t *testing.T) {
	r := SetUp()

	//create users
	User1, err := r.db.CreateUser(model.User{
		UserName: "username1",
		UserID:   "uid1",
		Role:     "",
	})
	assert.NoError(t, err)
	Admin, err := r.db.CreateUser(model.User{
		UserName: "username2",
		UserID:   "uid2",
		Role:     "admin",
	})
	assert.NoError(t, err)

	testCase := []struct {
		users    []string
		expected string
	}{
		//wrong format
		{[]string{"user1", "user2"}, "Could not remove users as admins: user1user2\n"},
		//doesn't existed users
		{[]string{"<@userid1|username1>", "<@userid2|username2>"}, "Could not remove users as admins: <@userid1|username1><@userid2|username2>\n"},
		//user is not admin
		{[]string{"<@" + User1.UserID + "|" + User1.UserName + ">"}, "Could not remove users as admins: <@uid1|username1>\n"},
		{[]string{"<@" + Admin.UserID + "|" + Admin.UserName + ">"}, "Users were removed as admins: <@uid2|username2>\n"},
	}
	for _, test := range testCase {
		actual := r.deleteAdmins(test.users)
		assert.Equal(t, test.expected, actual)
	}
	//delete users
	err = r.db.DeleteUser(User1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(Admin.ID)
	assert.NoError(t, err)
}
