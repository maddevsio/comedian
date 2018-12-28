package api

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"gitlab.com/team-monitoring/comedian/model"
// )

// func TestGenerateReportOnProject(t *testing.T) {
// 	r := SetUp()

// 	//create channels
// 	Channel, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "channel",
// 		ChannelID:   "cid",
// 	})
// 	assert.NoError(t, err)
// 	Channel1, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "channel1",
// 		ChannelID:   "cid1",
// 	})
// 	assert.NoError(t, err)
// 	//create channel members
// 	ChannelMember1, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:        "uid",
// 		ChannelID:     Channel1.ChannelID,
// 		RoleInChannel: "",
// 	})
// 	assert.NoError(t, err)
// 	//create channel <@channelID2|channel2>
// 	Channel2, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "channel2",
// 		ChannelID:   "channelID2",
// 	})
// 	assert.NoError(t, err)

// 	testCase := []struct {
// 		AccessLevel int
// 		Params      string
// 		Expected    string
// 	}{
// 		//not enough arguments
// 		{1, "#channel 2018-12-13", "Wrong number of arguments"},
// 		{1, "<#randomchanid|randomchan> 2018-12-12", "Wrong number of arguments"},
// 		//doesn't existed channels
// 		{1, "#random 2018-12-12 2018-12-13", "Неверное название проекта!"},
// 		{1, "<#channelid1|channelName123> 2018-12-12 2018-12-13", "Неверное название проекта!"},
// 		//channel without members
// 		{1, "#" + Channel.ChannelName + " 2018-12-12 2018-12-13", "Full Report on project #channel from 2018-12-12 to 2018-12-13:\n\nNo standup data for this period\n"},
// 		//wrong data format
// 		{1, "#" + Channel.ChannelName + " 2018-12 2018-12-13", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 		{1, "#" + Channel.ChannelName + " 2018-12-12 2018-12", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 		//channel with member
// 		{1, "#" + Channel1.ChannelName + " 2018-12-12 2018-12-13", "Full Report on project #channel1 from 2018-12-12 to 2018-12-13:\n\nNo standup data for this period\n"},
// 		//existed channel Channel2
// 		{1, "<#" + Channel2.ChannelID + "|" + Channel2.ChannelName + "> 2018-12-12 2018-12-13", "Full Report on project #channel2 from 2018-12-12 to 2018-12-13:\n\nNo standup data for this period\n"},
// 	}
// 	for _, test := range testCase {
// 		actual := r.generateReportOnProject(test.AccessLevel, test.Params)
// 		assert.Equal(t, test.Expected, actual)
// 	}
// 	//delete channel members
// 	err = r.db.DeleteChannelMember(ChannelMember1.UserID, ChannelMember1.ChannelID)
// 	assert.NoError(t, err)
// 	//delete channels
// 	err = r.db.DeleteChannel(Channel.ID)
// 	assert.NoError(t, err)
// 	err = r.db.DeleteChannel(Channel1.ID)
// 	assert.NoError(t, err)
// 	err = r.db.DeleteChannel(Channel2.ID)
// 	assert.NoError(t, err)
// }

// func TestGenerateReportOnUser(t *testing.T) {
// 	r := SetUp()

// 	//create user
// 	User1, err := r.db.CreateUser(model.User{
// 		UserName: "user1",
// 		UserID:   "uid1",
// 		Role:     "",
// 	})
// 	assert.NoError(t, err)

// 	testCase := []struct {
// 		accessLevel int
// 		params      string
// 		expected    string
// 	}{
// 		//not enough parameters
// 		{1, "@user 2018-12-12", "Wrong number of arguments"},
// 		{1, "<@userid|username> 2018-12-12", "Wrong number of arguments"},
// 		//doesn't existed users
// 		{1, "@user 2018-12-12 2018-12-13", "User does not exist!"},
// 		{1, "<@userid|username> 2018-12-12 2018-12-13", "User does not exist!"},
// 		//existed users
// 		{1, "@" + User1.UserName + " 2018-12-12 2018-12-13", "Full Report on user <@uid1> from 2018-12-12 to 2018-12-13:\n\nNo standup data for this period\n"},
// 		//wrong data format
// 		{1, "@" + User1.UserName + " 2018-12 2018-12-13", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 		{1, "@" + User1.UserName + " 2018-12-12 2018-12", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 	}
// 	for _, test := range testCase {
// 		actual := r.generateReportOnUser(test.accessLevel, test.params)
// 		assert.Equal(t, test.expected, actual)
// 	}
// 	//delete users
// 	err = r.db.DeleteUser(User1.ID)
// 	assert.NoError(t, err)
// }

// func TestGenerateReportOnUserInProject(t *testing.T) {
// 	r := SetUp()

// 	//create channel without members
// 	Channel1, err := r.db.CreateChannel(model.Channel{
// 		ChannelID:   "chanid1",
// 		ChannelName: "channel1",
// 	})
// 	assert.NoError(t, err)
// 	//create channel with members
// 	Channel2, err := r.db.CreateChannel(model.Channel{
// 		ChannelID:   "chanid2",
// 		ChannelName: "channel2",
// 	})
// 	assert.NoError(t, err)
// 	//create users
// 	//User1 member of Channel2
// 	User1, err := r.db.CreateUser(model.User{
// 		UserName: "user1",
// 		UserID:   "uid1",
// 	})
// 	assert.NoError(t, err)
// 	//user not a member of channel
// 	User2, err := r.db.CreateUser(model.User{
// 		UserName: "user2",
// 		UserID:   "uid2",
// 	})
// 	assert.NoError(t, err)
// 	//create channel members Channel2
// 	ChannelMember1, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:        User1.UserID,
// 		ChannelID:     Channel2.ChannelID,
// 		RoleInChannel: "",
// 	})
// 	assert.NoError(t, err)

// 	testCase := []struct {
// 		accessLevel int
// 		params      string
// 		expected    string
// 	}{
// 		//channel doesn't exist
// 		{1, "#channel @user 2018-12-12 2018-12-13", "Wrong project title!"},
// 		{1, "<#channelid|channel> <@userid|user> 2018-12-12 2018-12-13", "Wrong project title!"},
// 		//channel exist but hasn't members
// 		{1, "#" + Channel1.ChannelName + " @user 2018-12-12 2018-12-13", "No such user in your slack!"},
// 		{1, "<#" + Channel1.ChannelID + "|" + Channel1.ChannelName + "> <@userid|username> 2018-12-12 2018-12-13", "No such user in your slack!"},
// 		//crazy cases
// 		{1, "#" + Channel1.ChannelName + " <@userid|username> 2018-12-12 2018-12-13", "No such user in your slack!"},
// 		{1, "<#" + Channel1.ChannelID + "|" + Channel1.ChannelName + "> @user 2018-12-12 2018-12-13", "No such user in your slack!"},
// 		//channel has members, but this user not member of this channel
// 		{1, "#" + Channel2.ChannelName + " @" + User2.UserName + " 2018-12-12 2018-12-13", "<@uid2> does not have any role in this channel\n"},
// 		{1, "<#" + Channel2.ChannelID + "|" + Channel2.ChannelName + "> <@" + User2.UserID + "|" + User2.UserName + "> 2018-12-12 2018-12-13", "<@uid2> does not have any role in this channel\n"},
// 		//channel has members and user is member of this channel
// 		{1, "#" + Channel2.ChannelName + " @" + User1.UserName + " 2018-12-12 2018-12-13", "Report on user <@uid1> in project #channel2 from 2018-12-12 to 2018-12-13\n\nNo standup data for this period\n"},
// 		{1, "<#" + Channel2.ChannelID + "|" + Channel2.ChannelName + "> <@" + User1.UserID + "|" + User1.UserName + " 2018-12-12 2018-12-13", "Report on user <@uid1> in project #channel2 from 2018-12-12 to 2018-12-13\n\nNo standup data for this period\n"},
// 		//wrong data format
// 		{1, "#" + Channel2.ChannelName + " @" + User1.UserName + " 2018-12 2018-12-13", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 		{1, "#" + Channel2.ChannelName + " @" + User1.UserName + " 2018-12-12 2018-12", "parsing time \"2018-12\" as \"2006-01-02\": cannot parse \"\" as \"-\""},
// 	}
// 	for _, test := range testCase {
// 		actual := r.generateReportOnUserInProject(test.accessLevel, test.params)
// 		assert.Equal(t, test.expected, actual)
// 	}
// 	//delete user
// 	err = r.db.DeleteUser(User1.ID)
// 	assert.NoError(t, err)
// 	err = r.db.DeleteUser(User2.ID)
// 	assert.NoError(t, err)
// 	//delete channel members
// 	err = r.db.DeleteChannelMember(ChannelMember1.UserID, Channel2.ChannelID)
// 	assert.NoError(t, err)
// 	//delete channel
// 	err = r.db.DeleteChannel(Channel1.ID)
// 	assert.NoError(t, err)
// 	err = r.db.DeleteChannel(Channel2.ID)
// 	assert.NoError(t, err)
// }

// func TestGetChannelNameFromString(t *testing.T) {
// 	testCase := []struct {
// 		channel     string
// 		expected    string
// 		expectedErr error
// 	}{
// 		{"channel", "channel", nil},
// 		{"#channel", "channel", nil},
// 		{"<#channelid|channelname>", "channelname", nil},
// 	}
// 	for _, test := range testCase {
// 		actual, err := GetChannelNameFromString(test.channel)
// 		assert.Equal(t, test.expected, actual)
// 		assert.Equal(t, test.expectedErr, err)
// 	}
// }

// func TestGetUserNameFromString(t *testing.T) {
// 	testCase := []struct {
// 		channel     string
// 		expected    string
// 		expectedErr error
// 	}{
// 		{"user", "user", nil},
// 		{"@user", "user", nil},
// 		{"<@userid|username>", "username", nil},
// 	}
// 	for _, test := range testCase {
// 		actual, err := GetUserNameFromString(test.channel)
// 		assert.Equal(t, test.expected, actual)
// 		assert.Equal(t, test.expectedErr, err)
// 	}
// }

// func TestStandupReportByProject(t *testing.T) {
// 	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })

// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	bot, err := bot.NewBot(c)
// 	assert.NoError(t, err)
// 	r, err := NewReporter(bot)
// 	assert.NoError(t, err)
// 	bot.CP.CollectorEnabled = true

// 	channel, err := bot.DB.CreateChannel(model.Channel{
// 		ChannelName: "channame",
// 		ChannelID:   "chanid",
// 		StandupTime: int64(0),
// 	})
// 	assert.NoError(t, err)

// 	dateTo := time.Now()
// 	dateFrom := time.Now().AddDate(0, 0, -2)

// 	//First test when no data
// 	report, err := r.StandupReportByProject(channel, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected := "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, 0, len(report.ReportBody))

// 	d = time.Date(2018, 6, 4, 12, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })

// 	//create user who did not write standup
// 	user1, err := bot.DB.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID1",
// 		ChannelID: channel.ChannelID,
// 	})
// 	assert.NoError(t, err)

// 	standup0, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		UserID:    user1.UserID,
// 		Comment:   "",
// 		MessageTS: "1234",
// 	})
// 	assert.NoError(t, err)

// 	d = time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })

// 	//test for no standup submitted
// 	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)

// 	//create standup for user
// 	standup1, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		Comment:   "my standup",
// 		UserID:    user1.UserID,
// 		MessageTS: "123",
// 	})
// 	assert.NoError(t, err)

// 	//test if user submitted standup success
// 	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)
// 	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[1].Text)

// 	//create another user
// 	user2, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID2",
// 		ChannelID: channel.ChannelID,
// 	})
// 	assert.NoError(t, err)

// 	//test if one user wrote standup and the other did not
// 	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)
// 	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[1].Text)

// 	//create standup for user2
// 	standup2, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		Comment:   "user2 standup",
// 		UserID:    "userID2",
// 		MessageTS: "1234",
// 	})
// 	assert.NoError(t, err)

// 	//test if both users had written standups
// 	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n<@userID2> submitted standup: user2 standup \n================================================\n", report.ReportBody[0].Text)
// 	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n<@userID2> submitted standup: user2 standup \n================================================\n", report.ReportBody[1].Text)

// 	assert.NoError(t, r.db.DeleteStandup(standup0.ID))
// 	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
// 	assert.NoError(t, r.db.DeleteStandup(standup2.ID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user2.UserID, user2.ChannelID))
// 	assert.NoError(t, r.db.DeleteChannel(channel.ID))
// }

// func TestStandupReportByUser(t *testing.T) {
// 	d := time.Date(2018, 6, 5, 12, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	s, err := chat.NewSlack(c)
// 	assert.NoError(t, err)
// 	r := NewReporter(s)

// 	channel, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "chanName",
// 		ChannelID:   "chanid",
// 		StandupTime: int64(0),
// 	})
// 	assert.NoError(t, err)

// 	dateNext := time.Now().AddDate(0, 0, 1)
// 	dateTo := time.Now()
// 	dateFrom := time.Now().AddDate(0, 0, -2)

// 	user, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID1",
// 		ChannelID: channel.ChannelID,
// 	})
// 	assert.NoError(t, err)

// 	_, err = r.StandupReportByUser(user.UserID, dateTo, dateFrom)
// 	assert.Error(t, err)
// 	_, err = r.StandupReportByUser(user.UserID, dateNext, dateTo)
// 	assert.Error(t, err)
// 	_, err = r.StandupReportByUser(user.UserID, dateFrom, dateNext)
// 	assert.Error(t, err)

// 	expected := "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\n"
// 	report, err := r.StandupReportByUser(user.UserID, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, report.ReportHead)

// 	standup1, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		Comment:   "my standup",
// 		UserID:    user.UserID,
// 		MessageTS: "123",
// 	})
// 	expected = "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\n"
// 	report, err = r.StandupReportByUser(user.UserID, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-05\nIn #chanName <@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[0].Text)

// 	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user.UserID, user.ChannelID))
// 	assert.NoError(t, r.db.DeleteChannel(channel.ID))
// }

// func TestStandupReportByProjectAndUser(t *testing.T) {
// 	d := time.Date(2018, 6, 5, 12, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	s, err := chat.NewSlack(c)
// 	assert.NoError(t, err)
// 	r := NewReporter(s)

// 	channel, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "chanName",
// 		ChannelID:   "chanid",
// 		StandupTime: int64(0),
// 	})

// 	dateTo := time.Now()
// 	dateFrom := time.Now().AddDate(0, 0, -2)

// 	user1, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID1",
// 		ChannelID: channel.ChannelID,
// 	})

// 	report, err := r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected := "Report on user <@userID1> in project #chanName from 2018-06-03 to 2018-06-05\n\n"
// 	assert.Equal(t, expected, report.ReportHead)

// 	standup1, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		Comment:   "my standup",
// 		UserID:    "userID1",
// 		MessageTS: "123",
// 	})
// 	assert.NoError(t, err)

// 	report, err = r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo)
// 	assert.NoError(t, err)
// 	expected = "Report on user <@userID1> in project #chanName from 2018-06-03 to 2018-06-05\n\n"
// 	assert.Equal(t, expected, report.ReportHead)
// 	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n", report.ReportBody[0].Text)

// 	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
// 	assert.NoError(t, r.db.DeleteChannel(channel.ID))
// }
