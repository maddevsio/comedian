package api

import (
	"fmt"
	"testing"
	"time"

	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestAddCommand(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	user, err := bot.DB.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := bot.DB.CreateChannel(model.Channel{
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
		{2, channel.ChannelID, "<@userID|testUser> / wrongRole", "To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, default is a developer role, if the role is not selected! \n"},
	}

	for _, tt := range testCases {
		result := botAPI.addCommand(tt.accessLevel, tt.channelID, tt.params)
		assert.Equal(t, tt.output, result)

		members, err := bot.DB.ListAllChannelMembers()
		assert.NoError(t, err)
		for _, m := range members {
			assert.NoError(t, bot.DB.DeleteChannelMember(m.UserID, m.ChannelID))
		}
	}

	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
	assert.NoError(t, bot.DB.DeleteUser(user.ID))
}

func TestShowCommand(t *testing.T) {
	//modify test to cover more cases: no users, etc.
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	user, err := bot.DB.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "",
	})
	assert.NoError(t, err)

	admin, err := bot.DB.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "admin",
	})
	assert.NoError(t, err)

	memberPM, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        user.UserID,
		ChannelID:     channel.ChannelID,
		RoleInChannel: "pm",
	})

	memberDeveloper, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        user.UserID,
		ChannelID:     channel.ChannelID,
		RoleInChannel: "developer",
	})

	testCases := []struct {
		params string
		output string
	}{
		{"", "Standuper in this channel: <@userID>"},
		{"admin", "Admin in this workspace: <@testUser>"},
		{"developer", "Standuper in this channel: <@userID>"},
		{"pm", "PM in this channel: <@userID>"},
		{"randomRole", "To view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_ \n"},
	}

	for _, tt := range testCases {
		result := botAPI.showCommand(channel.ChannelID, tt.params)
		assert.Equal(t, tt.output, result)
	}

	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
	assert.NoError(t, bot.DB.DeleteUser(user.ID))
	assert.NoError(t, bot.DB.DeleteUser(admin.ID))
	assert.NoError(t, bot.DB.DeleteChannelMember(memberPM.UserID, memberPM.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(memberDeveloper.UserID, memberDeveloper.ChannelID))
}

func TestDeleteCommand(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	testCase := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, "chan1", "<@id|name> / admin", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{4, "chan1", "<@id|name> / pm", "Access Denied! You need to be at least PM in this project to use this command!"},
		{4, "chan1", "<@id|name> / random", "To remove members use `remove` command. If you provide a role name, you will remove members with this role. _admin, pm, developer, designer_ \n"},
		{4, "chan1", "<@id|name>", "Access Denied! You need to be at least PM in this project to use this command!"},
	}
	for _, test := range testCase {
		actual := botAPI.deleteCommand(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
}

func TestAddMembers(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel member with role pm
	_, err = bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserId1",
		ChannelID:     "chan1",
		RoleInChannel: "pm",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates channel member with role developer
	_, err = bot.DB.CreateChannelMember(model.ChannelMember{
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
		actual := botAPI.addMembers(test.Users, test.RoleInChannel, "chan1")
		assert.Equal(t, test.Expected, actual)
	}
	//deletes channelMembers
	//6 members will be created
	for i := 1; i <= 6; i++ {
		err = bot.DB.DeleteChannelMember(fmt.Sprintf("testUserId%v", i), "chan1")
		assert.NoError(t, err)
	}
}

func TestAddAdmins(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//not admin user
	User1, err := bot.DB.CreateUser(model.User{
		UserName: "UserName1",
		UserID:   "uid1",
		Role:     "",
	})
	assert.NoError(t, err)
	//admin user
	UserAdmin, err := bot.DB.CreateUser(model.User{
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
		actual := botAPI.addAdmins(test.Users)
		assert.Equal(t, test.Expected, actual)
	}
	//delete users
	err = bot.DB.DeleteUser(User1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(UserAdmin.ID)
	assert.NoError(t, err)
}

func TestListMembers(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel with pm members
	Channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "chanName1",
		ChannelID:   "cid1",
	})
	assert.NoError(t, err)
	//creates channel Channel1 members PMs
	ChanMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     Channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)
	ChanMember2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     Channel1.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)

	//creates channel with members
	Channel2, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "chanName2",
		ChannelID:   "cid2",
	})
	assert.NoError(t, err)
	//creates channel members
	ChanMember3, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     Channel2.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	ChanMember4, err := bot.DB.CreateChannelMember(model.ChannelMember{
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
		{"channel", "pm", "No PMs in this channel! To add one, please, use `/comedian add` slash command"},
		{Channel1.ChannelID, "pm", "PMs in this channel: <@uid1>, <@uid2>"},
		{"channel", "", "No standupers in this channel! To add one, please, use `/comedian add` slash command"},
		{Channel2.ChannelID, "", "Standupers in this channel: <@uid3>, <@uid4>"},
	}
	for _, test := range testCase {
		actual := botAPI.listMembers(test.Channel, test.Role)
		assert.Equal(t, test.Expected, actual)
	}
	//delete channel members
	err = bot.DB.DeleteChannelMember(ChanMember1.UserID, ChanMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(ChanMember2.UserID, ChanMember2.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(ChanMember3.UserID, ChanMember3.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(ChanMember4.UserID, ChanMember4.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = bot.DB.DeleteChannel(Channel1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannel(Channel2.ID)
	assert.NoError(t, err)
}

func TestListAdmins(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//no admins in db
	actual := botAPI.listAdmins()
	assert.Equal(t, "No admins in this workspace! To add one, please, use `/comedian add` slash command", actual)

	//creates users admins
	Admin1, err := bot.DB.CreateUser(model.User{
		UserName: "Username1",
		UserID:   "uid1",
		Role:     "admin",
	})
	assert.NoError(t, err)
	Admin2, err := bot.DB.CreateUser(model.User{
		UserName: "Username2",
		UserID:   "uid2",
		Role:     "admin",
	})
	assert.NoError(t, err)

	actual = botAPI.listAdmins()
	assert.Equal(t, "Admins in this workspace: <@Username1>, <@Username2>", actual)

	//delete users
	err = bot.DB.DeleteUser(Admin1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(Admin2.ID)
	assert.NoError(t, err)
}

func TestDeleteMembers(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//create channel
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "cid1",
	})
	assert.NoError(t, err)
	//creates channel members
	channelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
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
		{[]string{"<@" + channelMember1.UserID + "|username>"}, channel1.ChannelID, "The following members were removed: <@uid1|username>\n"},
	}
	for _, test := range testCase {
		actual := botAPI.deleteMembers(test.members, test.channelID)
		assert.Equal(t, test.expected, actual)
	}
	//delete channel members
	err = bot.DB.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channel
	err = bot.DB.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
}

func TestDeleteAdmins(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//create users
	User1, err := bot.DB.CreateUser(model.User{
		UserName: "username1",
		UserID:   "uid1",
		Role:     "",
	})
	assert.NoError(t, err)
	Admin, err := bot.DB.CreateUser(model.User{
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
		actual := botAPI.deleteAdmins(test.users)
		assert.Equal(t, test.expected, actual)
	}
	//delete users
	err = bot.DB.DeleteUser(User1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(Admin.ID)
	assert.NoError(t, err)
}
