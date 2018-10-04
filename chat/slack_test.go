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
		{"all key words", "Yesterday managed to get docker up and running, today will complete test #100, questions: to be or not to be?", true},
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
		_, ok, _ := s.analizeStandup(tt.input)
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
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	su1, err := s.db.CreateChannelMember(model.ChannelMember{
		UserID:      "UBA5V5W9K",
		ChannelID:   "123qwe",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "123qwe", su1.ChannelID)
	assert.Equal(t, "UBA5V5W9K", su1.UserID)

	err = s.SendUserMessage("USLACKBOT", "MSG to User!")
	assert.NoError(t, err)

	assert.NoError(t, s.db.DeleteChannelMember(su1.UserID, su1.ChannelID))

}

func TestHandleMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `"ok": true`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/reactions.add", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	botUserID := "BOTID"

	su1, err := s.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	msg := &slack.MessageEvent{}
	msg.Text = "<@> some message"
	msg.Channel = su1.ChannelID
	msg.Timestamp = "1"

	err = s.handleMessage(msg, botUserID)
	assert.NoError(t, err)

	fakeChannel := "someotherChan"

	msg.Text = "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = su1.ChannelID
	msg.User = su1.UserID
	msg.Timestamp = "2"

	err = s.handleMessage(msg, botUserID)

	editmsg := &slack.MessageEvent{
		SubMessage: &slack.Msg{
			User:      su1.UserID,
			Text:      "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problem",
			Timestamp: "2",
		},
	}
	editmsg.SubType = typeEditMessage

	err = s.handleMessage(editmsg, botUserID)
	assert.NoError(t, err)

	msg.Text = "Yesterday: did crazy tests, today: doing a lot of crazy tests, problems: no problems!"
	msg.Channel = fakeChannel
	msg.User = su1.UserID

	err = s.handleMessage(msg, botUserID)
	assert.NoError(t, err)

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	// clean up
	standups, err := s.db.ListStandups()
	assert.NoError(t, err)
	for _, standup := range standups {
		s.db.DeleteStandup(standup.ID)
	}
	assert.NoError(t, s.db.DeleteChannelMember(su1.UserID, su1.ChannelID))

}
