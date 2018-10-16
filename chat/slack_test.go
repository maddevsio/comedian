package chat

import (
	"testing"
	"time"

	"github.com/bouk/monkey"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type MessageEvent struct {
}

type Channel struct {
	id string
}

type imOpenResp struct {
	ok      bool
	channel Channel
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
	}
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	for _, tt := range testCases {
		ok, _ := s.analizeStandup(tt.input)
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
	s, err := NewSlack(c)
	assert.NoError(t, err)
	err = s.SendMessage("YYYZZZVVV", "Hey!")
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
	s, err := NewSlack(c)
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
	s, err := NewSlack(c)
	assert.NoError(t, err)
	botUserID := "BOTID"

	su1, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "123qwe",
	})

	su, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	msg := &slack.MessageEvent{}
	msg.Text = "<@> some message"
	msg.Channel = su.ChannelID
	msg.Timestamp = "1"

	err = s.handleMessage(msg, botUserID)
	assert.NoError(t, err)

	editWrongMessage := &slack.MessageEvent{
		SubMessage: &slack.Msg{
			User:      su.UserID,
			Text:      "<@BOTID> Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problem",
			Timestamp: "123ksdlfsdkl",
		},
	}
	editWrongMessage.SubType = typeEditMessage

	err = s.handleMessage(msg, botUserID)
	assert.NoError(t, err)

	fakeChannel := "someotherChan"

	msg.Text = "<@BOTID> Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = su1.ChannelID
	msg.User = su1.UserID
	msg.Timestamp = "2"

	err = s.handleMessage(msg, botUserID)

	editmsg := &slack.MessageEvent{
		SubMessage: &slack.Msg{
			User:      su1.UserID,
			Text:      "<@BOTID> Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problem",
			Timestamp: "2",
		},
	}
	editmsg.SubType = typeEditMessage

	err = s.handleMessage(editmsg, botUserID)
	assert.NoError(t, err)

	msg.Text = "<@BOTID> Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = fakeChannel
	msg.User = su1.UserID

	err = s.handleMessage(msg, botUserID)
	assert.NoError(t, err)

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	// clean up
	standups, err := s.DB.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		s.DB.DeleteStandup(standup.ID)
	}
	assert.NoError(t, s.DB.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, s.DB.DeleteChannelMember(su.UserID, su.ChannelID))

}

func TestFillStandupsForNonReporters(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 30, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	s.FillStandupsForNonReporters()

	d = time.Date(2018, 10, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	s.FillStandupsForNonReporters()

	su1, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	s.FillStandupsForNonReporters()

	d = time.Date(2018, 10, 8, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	su2, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	_, err = s.DB.CreateStandup(model.Standup{
		ChannelID: su2.ChannelID,
		Comment:   "test test test!",
		UserID:    su2.UserID,
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 10, 8, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	s.FillStandupsForNonReporters()

	// clean up
	standups, err := s.DB.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		s.DB.DeleteStandup(standup.ID)
	}

	assert.NoError(t, s.DB.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, s.DB.DeleteChannelMember(su2.UserID, su2.ChannelID))
}

func TestAutomaticActions(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://slack.com/api/conversations.info", httpmock.NewStringResponder(200, `{"ok": true, "channel": {"id": "CBAPFA2J2", "name": "general"}}`))

	teamInfoResponse := `
	{
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
				"is_bot": false,
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
				"is_app_user": false,
			},
			{
				"id": "DELETEDUSERID",
				"team_id": "TEAMID1",
				"name": "deleted.user",
				"deleted": true,
				"color": "e96699",
				"real_name": "John Doe"
			},
		],
		"cache_ts": 1538988885
	}`

	httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", httpmock.NewStringResponder(200, teamInfoResponse))

	s.UpdateUsersList()

	s.handleJoin("TESTCHANNELID")
}
