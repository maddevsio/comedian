package chat

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"

	"github.com/maddevsio/comedian/chat"
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
		{"keywords caps", "shit, TODAY will fap like crazy, PROBLEMS: no!", false},
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

		s.handleMessage(msg, botUserID)
	}

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

	_, err = s.DB.CreateUser(model.User{
		UserName: "deleted user",
		UserID:   "xxx",
	})
	assert.NoError(t, err)

	chanMember, err := s.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "xxx",
		ChannelID: "YYY",
	})
	assert.NoError(t, err)

	_, err = s.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chanMember.ID,
	})
	assert.NoError(t, err)

	s.UpdateUsersList()

	s.handleJoin("TESTCHANNELID")

	users, err := s.DB.ListUsers()
	assert.NoError(t, err)
	for _, u := range users {
		assert.NoError(t, s.DB.DeleteUser(u.ID))
	}
}

func TestTeamMonitoringOnWeekDay(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)

	tm, err := NewTeamMonitoring(s)
	assert.NoError(t, err)

	testCases := []struct {
		standupComment                           string
		userID                                   string
		collectorResponseUserStatusCode          int
		collectorResponseUserBody                string
		collectorResponseUserInProjectStatusCode int
		collectorResponseUserInProjectBody       string
		color                                    string
		value                                    string
	}{
		{"users", "userID1", 200, `{"worklogs": 35000, "total_commits": 23}`, 200,
			`{"worklogs": 35000, "total_commits": 23}`, "good",
			" worklogs: 9:43 :sunglasses: | commits: 23 :tada: | standup :heavy_check_mark: |\n"},
		{"", "userID2", 200, `{"worklogs": 0, "total_commits": 0}`, 200,
			`{"worklogs": 0, "total_commits": 0}`, "danger",
			" worklogs: 0:00 :angry: | commits: 0 :shit: | standup :x: |\n"},
		{"test", "userID3", 200, `{"worklogs": 28800, "total_commits": 1}`, 200,
			`{"worklogs": 28800, "total_commits": 1}`, "good",
			" worklogs: 8:00 :wink: | commits: 1 :tada: | standup :heavy_check_mark: |\n"},
		{"test", "userID4", 200, `{"worklogs": 14400, "total_commits": 1}`, 200,
			`{"worklogs": 12400, "total_commits": 1}`, "warning",
			" worklogs: 3:26 out of 4:00 :disappointed: | commits: 1 :tada: | standup :heavy_check_mark: |\n"},
	}

	for _, tt := range testCases {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		channel, err := tm.db.CreateChannel(model.Channel{
			ChannelID:   "QWERTY123",
			ChannelName: "chanName1",
			StandupTime: int64(0),
		})
		assert.NoError(t, err)

		user, err := tm.db.CreateUser(model.User{
			UserID:   tt.userID,
			UserName: "Alfa",
		})
		assert.NoError(t, err)

		channelMember, err := tm.db.CreateChannelMember(model.ChannelMember{
			UserID:    tt.userID,
			ChannelID: channel.ChannelID,
		})
		assert.NoError(t, err)

		// if create stanupd yes, create standup!

		standup, err := tm.db.CreateStandup(model.Standup{
			Created:   time.Now(),
			Modified:  time.Now(),
			ChannelID: channel.ChannelID,
			Comment:   tt.standupComment,
			UserID:    channelMember.UserID,
			MessageTS: "randomMessageTS",
		})
		assert.NoError(t, err)

		d = time.Date(2018, 9, 18, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		dateOfRequest := fmt.Sprintf("%d-%02d-%02d", time.Now().AddDate(0, 0, -1).Year(), time.Now().AddDate(0, 0, -1).Month(), time.Now().AddDate(0, 0, -1).Day())

		linkURLUsers := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "users", user.UserID, dateOfRequest, dateOfRequest)
		linkURLUserInProject := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "user-in-project", fmt.Sprintf("%v/%v", user.UserID, channel.ChannelName), dateOfRequest, dateOfRequest)
		httpmock.RegisterResponder("GET", linkURLUsers, httpmock.NewStringResponder(tt.collectorResponseUserStatusCode, tt.collectorResponseUserBody))
		httpmock.RegisterResponder("GET", linkURLUserInProject, httpmock.NewStringResponder(tt.collectorResponseUserInProjectStatusCode, tt.collectorResponseUserInProjectBody))

		attachments, err := tm.RevealRooks()
		assert.NoError(t, err)
		assert.Equal(t, tt.color, attachments[0].Color)
		assert.Equal(t, fmt.Sprintf("<@%v> in #%v", user.UserID, channel.ChannelName), attachments[0].Text)
		assert.Equal(t, tt.value, attachments[0].Fields[0].Value)

		tm.reportRooks()

		assert.NoError(t, tm.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
		assert.NoError(t, tm.db.DeleteStandup(standup.ID))
		assert.NoError(t, tm.db.DeleteUser(user.ID))
		assert.NoError(t, tm.db.DeleteChannel(channel.ID))
	}
}
