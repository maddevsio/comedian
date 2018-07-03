package chat

import (
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

type MessageEvent struct {
}

func TestCleanMessage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	s.myUsername = "comedian"
	assert.NoError(t, err)
	text, ok := s.cleanMessage("<@comedian> hey there")
	assert.Equal(t, "hey there", text)
	assert.True(t, ok)
	text, ok = s.cleanMessage("What's up?")
	assert.Equal(t, "What's up?", text)
	assert.False(t, ok)
}

func TestHandleMessage(t *testing.T) {
	// c, err := config.Get()
	// assert.NoError(t, err)
	// s, err := NewSlack(c)
	// assert.NoError(t, err)

	// edited := &slack.Edited{
	// 	User:      "",
	// 	Timestamp: "",
	// }

	// icon := &slack.Icon{
	// 	// IconURL:   "",
	// 	// IconEmoji: "",
	// }

	// file := &slack.File{}

	// comment := &slack.Comment{}

	// msg := &slack.MessageEvent{
	// 	slack.Msg{
	// 		Type:             "",
	// 		Channel:          "",
	// 		User:             "",
	// 		Text:             "",
	// 		Timestamp:        "",
	// 		ThreadTimestamp:  "",
	// 		IsStarred:        false,
	// 		PinnedTo:         make([]string, 1),
	// 		Attachments:      make([]slack.Attachment, 1),
	// 		Edited:           edited,
	// 		SubType:          "",
	// 		Hidden:           false,
	// 		DeletedTimestamp: "",
	// 		EventTimestamp:   "",
	// 		BotID:            "",
	// 		Username:         "",
	// 		Icons:            icon,
	// 		Inviter:          "",
	// 		Topic:            "",
	// 		Purpose:          "",
	// 		Name:             "",
	// 		OldName:          "",
	// 		Members:          make([]string, 1),
	// 		ReplyCount:       0,
	// 		Replies:          make([]slack.Reply, 1),
	// 		ParentUserId:     "",
	// 		File:             file,
	// 		Upload:           false,
	// 		Comment:          comment,
	// 		ItemType:         "",
	// 		ReplyTo:          0,
	// 		Team:             "",
	// 		Reactions:        make([]slack.ItemReaction, 1),
	// 	},
	// }
	// msg := &slack.MessageEvent{}
	// err = s.handleMessage(msg)
	// assert.NoError(t, err)
}

func TestHandleEditMessage(t *testing.T) {
	// c, err := config.Get()
	// assert.NoError(t, err)
	// s, err := NewSlack(c)
	// assert.NoError(t, err)
	// msg := &slack.MessageEvent{}
	// err = s.handleMessage(msg)
	// assert.NoError(t, err)
}

func TestSendMessage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	err = s.SendMessage(c.DirectManagerChannelID, "MSG to manager!")
	assert.NoError(t, err)
}

func TestSendUserMessage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	su1, err := s.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user1", su1.SlackName)

	// err = s.SendUserMessage(su1.SlackUserID, "MSG to User!")
	// assert.NoError(t, err)

	assert.NoError(t, s.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

}

func TestGetAllUsersToDB(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)

	usersInChan, err := s.db.ListStandupUsersByChannelName("general")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(usersInChan))

	err = s.GetAllUsersToDB()
	usersInChan, err = s.db.ListStandupUsersByChannelName("general")
	assert.NoError(t, err)

	assert.True(t, len(usersInChan) > 0)

	for _, user := range usersInChan {
		assert.NoError(t, s.db.DeleteStandupUserByUsername(user.SlackName, user.ChannelID))
	}

}
