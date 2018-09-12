package chat

import (
	"testing"

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
		{"no key words", "i want to create a standup but totaly forgot the way i should write it!", false},
		{"key words yesterday", "Yesterday it was fucking awesome!", false},
		{"key words yesterday and today", "Вчера ломал сервер, сегодня будет охренеть много дел", false},
		{"all key words capitalized", "Yesterday: launched MySQL, Today: will scream and should, Problems: SHIT IS ALL OVER!", true},
		{"keywords with mistakes", "Yesday completed shit, dotay will fap like crazy, promlems: no problems!", false},
	}
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	for _, tt := range testCases {
		_, ok := s.isStandup(tt.input)
		if ok != tt.confirm {
			t.Errorf("Test %s: \n input: %s,\n expected confirm: %v\n actual confirm: %v \n", tt.title, tt.input, tt.confirm, ok)
		}
	}
}

func TestSendMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	err = s.SendMessage("YYYZZZVVV", "Hey!")
	assert.NoError(t, err)
}

func TestSendUserMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	su1, err := s.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "UBA5V5W9K",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user1", su1.SlackName)
	assert.NotEqual(t, "userID1", su1.SlackUserID)

	err = s.SendUserMessage("USLACKBOT", "MSG to User!")

	assert.NoError(t, s.db.DeleteStandupUser(su1.SlackName, su1.ChannelID))

}

func TestHandleMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `"ok": true`))

	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	su1, err := s.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "testchannel",
	})
	assert.NoError(t, err)

	msg := &slack.MessageEvent{}
	msg.Text = "<@> some message"
	msg.Channel = su1.Channel
	msg.Username = su1.SlackName
	msg.Timestamp = "1"

	err = s.handleMessage(msg)
	assert.NoError(t, err)

	fakeChannel := "someotherChan"

	msg.Text = "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = su1.Channel
	msg.Username = su1.SlackName
	msg.Timestamp = "2"
	err = s.handleMessage(msg)

	editmsg := &slack.MessageEvent{
		SubMessage: &slack.Msg{
			Text:      "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problem",
			Timestamp: "2",
		},
	}
	editmsg.SubType = typeEditMessage

	err = s.handleMessage(editmsg)
	assert.NoError(t, err)

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": false, "error": "channel_not_found"}`))
	msg.Text = "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = fakeChannel
	msg.Username = su1.SlackName

	err = s.handleMessage(msg)
	assert.Error(t, err)

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))
	err = s.handleConnection()
	assert.NoError(t, err)

	// clean up
	standups, err := s.db.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		s.db.DeleteStandup(standup.ID)
	}
	assert.NoError(t, s.db.DeleteStandupUser(su1.SlackName, su1.ChannelID))

}

func TestGetChannelName(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/channels.info", httpmock.NewStringResponder(200, `{"ok": true, "channel": {"id": "RANDOMID","name": "randomName"}}`))

	httpmock.RegisterResponder("POST", "https://slack.com/api/groups.info", httpmock.NewStringResponder(200, `{"ok": true, "group": {"id": "RANDOMID","name": "randomName"}}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	chanName, err := s.GetChannelName("RANDOMID")
	assert.NoError(t, err)
	assert.Equal(t, "randomName", chanName)
}
