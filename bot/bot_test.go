package bot

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestHandleLeft(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)

	//create channel
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "chanid1",
	})
	assert.NoError(t, err)
	//channel members
	ChannelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	//create timetable for ChannelMember1
	TimeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: ChannelMember1.ID,
		Monday:          12345,
		Tuesday:         0,
		Thursday:        0,
		Friday:          12345,
		Saturday:        12345,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = bot.DB.UpdateTimeTable(TimeTable1)
	assert.NoError(t, err)

	testCase := []struct {
		ChannelID string
		UserID    string
	}{
		{channel.ChannelID, ChannelMember1.UserID},
	}
	for _, test := range testCase {
		bot.handleLeft(test.ChannelID, test.UserID)
	}
	//check that ChannelMember1 doesn't exist
	_, err = bot.DB.SelectChannelMember(ChannelMember1.ID)
	if err == nil {
		t.Error()
	}
	//check that timetable doesn't exist
	_, err = bot.DB.SelectTimeTable(ChannelMember1.ID)
	if err == nil {
		t.Error()
	}
	//delete channel
	err = bot.DB.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestHandleBotRemovedFromChannel(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)

	//create channel
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "channel1",
		ChannelID:   "chanid1",
	})
	assert.NoError(t, err)
	//channel members
	ChannelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	//create timetable for ChannelMember1
	TimeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: ChannelMember1.ID,
		Monday:          12345,
		Tuesday:         0,
		Thursday:        0,
		Friday:          12345,
		Saturday:        12345,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = bot.DB.UpdateTimeTable(TimeTable1)
	assert.NoError(t, err)
	ChannelMember2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	//create timetable for ChannelMember2
	TimeTable2, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: ChannelMember2.ID,
		Monday:          12345,
		Tuesday:         0,
		Thursday:        0,
		Friday:          12345,
		Saturday:        12345,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = bot.DB.UpdateTimeTable(TimeTable2)
	assert.NoError(t, err)

	testCase := []struct {
		ChannelID string
	}{
		{channel.ChannelID},
	}
	for _, test := range testCase {
		bot.handleBotRemovedFromChannel(test.ChannelID)
	}
	//check that ChannelMember1 doesn't exist
	_, err = bot.DB.SelectChannelMember(ChannelMember1.ID)
	if err == nil {
		t.Error()
	}
	//check that timetable doesn't exist
	_, err = bot.DB.SelectTimeTable(ChannelMember1.ID)
	if err == nil {
		t.Error()
	}
	//check that ChannelMember2 doesn't exist
	_, err = bot.DB.SelectChannelMember(ChannelMember2.ID)
	if err == nil {
		t.Error()
	}
	//check that timetable doesn't exist
	_, err = bot.DB.SelectTimeTable(ChannelMember2.ID)
	if err == nil {
		t.Error()
	}
	//delete channel
	err = bot.DB.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestIsStandup(t *testing.T) {
	testCases := []struct {
		title   string
		input   string
		confirm bool
	}{
		{"all key words", "Yesterday managed to get docker up and running, today will complete test #100, problems: I have multilang!", true},
		{"all key words", "Yesterday managed to get docker up and running, today will complete test #100, questions: to be or not to be?", true},
		{"no key words", "i want to create a standup but totaly forgot the way i should write it!", false},
		{"key words yesterday", "Yesterday it was fucking awesome!", false},
		{"key words yesterday and today", "Вчера ломал сервер, сегодня будет охренеть много дел", false},
		{"all key words capitalized", "Yesterday: launched MySQL, Today: will scream and should, Problems: SHIT IS ALL OVER!", true},
		{"keywords with mistakes", "Yesday completed shit, dotay will fap like crazy, promlems: no problems!", false},
		{"keywords caps", "YESTERDAY completed shit, TODAY will fap like crazy, PROBLEMS: no !", true},
		{"keywords caps", "shit, TODAY will fap like crazy, PROBLEMS: no!", false},
	}
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)
	for _, tt := range testCases {
		ok, _ := bot.analizeStandup(tt.input)
		if ok != tt.confirm {
			t.Errorf("Test %s: \n input: %s,\n expected confirm: %v\n actual confirm: %v \n", tt.title, tt.input, tt.confirm, ok)
		}
	}
}

func TestSendMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postEphemeral", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewBot(c)
	assert.NoError(t, err)
	err = s.SendMessage("YYYZZZVVV", "Hey!", nil)
	assert.NoError(t, err)

	err = s.SendEphemeralMessage("YYYZZZVVV", "USER!", "Hey!")
	assert.NoError(t, err)
}

func TestSendUserMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewBot(c)
	assert.NoError(t, err)

	su1, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:      "UBA5V5W9K",
		ChannelID:   "123qwe",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "123qwe", su1.ChannelID)
	assert.Equal(t, "UBA5V5W9K", su1.UserID)

	err = s.SendUserMessage("USLACKBOT", "MSG to User!")
	assert.NoError(t, err)

	assert.NoError(t, s.DB.DeleteChannelMember(su1.UserID, su1.ChannelID))

}

func TestHandleMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `"ok": true`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/reactions.add", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postEphemeral", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)

	botUserID := "BOTID"

	su1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "123qwe",
	})

	su, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	testCases := []struct {
		text      string
		user      string
		channel   string
		timestamp string
		subType   string
	}{
		{"Hey! Please, accept my standup", "testUser", "testChannel", "1500000", typeMessage},
		{"Hey! Please, accept my standup", "testUser", "testChannel", "1500000", typeEditMessage},
		{"Hey! Please, accept my standup", "testUser", "testChannel", "1500000", typeDeleteMessage},
		{"#standup Hey! Please, accept my standup", "testUser", "testChannel", "1500000", typeMessage},
		{"<@BOTID> Hey! Yesterday, today, problems", "", "testChannel", "1500000", typeMessage},
		{"<@BOTID> Hey My First Standup! Yesterday, today, problems", "testUser", "testChannel", "1500000", typeMessage},
		{"<@BOTID> My second standup! Yesterday, today, problems", "testUser", "testChannel", "1500000", typeMessage},
		{"<@BOTID> Hey I want to edit! Text of not a standup", "testUser", "testChannel", "1500000", typeEditMessage},
		{"<@BOTID> Hey I want to edit! Yesterday, today, problems", "testUser", "testChannel", "1500000", typeEditMessage},
		{"<@BOTID> Different MSGTS! No Standup", "testUser", "testChannel", "15", typeEditMessage},
		{"<@BOTID> Different MSGTS! Yesterday Today Problems", "testUser", "", "15", typeEditMessage},
		{"<@BOTID> Different MSGTS! Yesterday Today Problems", "testUser", "testChannel", "15", typeEditMessage},
	}

	for _, tt := range testCases {
		msg := &slack.MessageEvent{}
		msg.Text = tt.text
		msg.User = tt.user
		msg.Channel = tt.channel
		msg.Timestamp = tt.timestamp
		msg.SubType = tt.subType

		if tt.subType == typeEditMessage {
			msg = &slack.MessageEvent{
				SubMessage: &slack.Msg{
					User:      tt.user,
					Text:      tt.text,
					Timestamp: tt.timestamp,
				},
			}
			msg.Channel = tt.channel
			msg.SubType = tt.subType
		}

		bot.handleMessage(msg, botUserID)
	}

	// clean up
	standups, err := bot.DB.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		bot.DB.DeleteStandup(standup.ID)
	}
	assert.NoError(t, bot.DB.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(su.UserID, su.ChannelID))

}

func TestFillStandupsForNonReporters(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 30, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	bot.FillStandupsForNonReporters()

	d = time.Date(2018, 10, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	bot.FillStandupsForNonReporters()

	su1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	bot.FillStandupsForNonReporters()

	d = time.Date(2018, 10, 8, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	su2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	_, err = bot.DB.CreateStandup(model.Standup{
		ChannelID: su2.ChannelID,
		Comment:   "test test test!",
		UserID:    su2.UserID,
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 10, 8, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	bot.FillStandupsForNonReporters()

	// clean up
	standups, err := bot.DB.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		bot.DB.DeleteStandup(standup.ID)
	}

	assert.NoError(t, bot.DB.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(su2.UserID, su2.ChannelID))
}

func TestAutomaticActions(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)

	r := httpmock.NewStringResponder(200, `{
		"ok": true,
		"members": [
			{
				"id": "USER1D1",
				"team_id": "TEAMID1",
				"name": "UserAdmin",
				"deleted": false,
				"color": "9f69e7",
				"real_name": "admin",
				"is_admin": true,
				"is_owner": true,
				"is_primary_owner": true,
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": false
			},
			{
				"id": "BOTID",
				"team_id": "TEAMID1",
				"name": "comedian",
				"deleted": false,
				"color": "4bbe2e",
				"real_name": "comedian",
				"tz": "America\/Los_Angeles",
				"tz_label": "Pacific Daylight Time",
				"tz_offset": -25200,
				"is_admin": false,
				"is_owner": false,
				"is_primary_owner": false,
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": true,
				"is_app_user": false,
				"updated": 1529488035
			},
			{
				"id": "UBEGJBB9A",
				"team_id": "TEAMID1",
				"name": "anot",
				"deleted": false,
				"color": "674b1b",
				"real_name": "Anot",
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": false,
				"is_app_user": false
			},
			{
				"id": "xxx",
				"team_id": "TEAMID1",
				"name": "deleted user",
				"deleted": true,
				"color": "674b1b",
				"real_name": "test user",
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": false,
				"is_app_user": false
			}
		]
	}`)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://slack.com/api/conversations.info", httpmock.NewStringResponder(200, `{"ok": true, "channel": {"id": "CBAPFA2J2", "name": "general"}}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", r)

	_, err = bot.DB.CreateUser(model.User{
		UserName: "deleted user",
		UserID:   "xxx",
	})
	assert.NoError(t, err)

	chanMember, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "xxx",
		ChannelID: "YYY",
	})
	assert.NoError(t, err)

	_, err = bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chanMember.ID,
	})
	assert.NoError(t, err)

	bot.UpdateUsersList()

	bot.handleJoin("TESTCHANNELID")

	users, err := bot.DB.ListUsers()
	assert.NoError(t, err)
	for _, u := range users {
		assert.NoError(t, bot.DB.DeleteUser(u.ID))
	}
	//delete created channel `{"ok": true, "channel": {"id": "CBAPFA2J2", "name": "general"}}`
	channel, err := bot.DB.SelectChannel("CBAPFA2J2")
	assert.NoError(t, err)
	err = bot.DB.DeleteChannel(channel.ID)
}

func TestRecordBug(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := NewBot(c)
	assert.NoError(t, err)
	bot.CP.Language = "en_US"

	//creates user
	user1, err := bot.DB.CreateUser(model.User{
		UserID:   "uid1",
		UserName: "user1",
		Role:     "",
	})
	assert.NoError(t, err)

	testCase := []struct {
		ChannelID string
		userID    string
		Bug       string
	}{
		{"channel1", "user", "bug"},
		{"channel1", user1.UserID, "bug"},
	}
	for _, test := range testCase {
		bot.recordBug(test.ChannelID, test.userID, test.Bug)
	}
	//deletes user
	err = bot.DB.DeleteUser(user1.ID)
	assert.NoError(t, err)
}
