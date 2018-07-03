package chat

import (
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/stretchr/testify/assert"
)

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

func TestSendMessage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	err = s.SendMessage(c.DirectManagerChannelID, "MSG to manager!")
	assert.NoError(t, err)
}

// func TestSendUserMessage(t *testing.T) {
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	s, err := NewSlack(c)
// 	assert.NoError(t, err)

// 	su1, err := s.db.CreateStandupUser(model.StandupUser{
// 		SlackUserID: "userID1",
// 		SlackName:   "user1",
// 		ChannelID:   "123qwe",
// 		Channel:     "channel1",
// 	})
// 	assert.NoError(t, err)

// 	err = s.SendUserMessage("userID1", "MSG to manager!")
// 	assert.NoError(t, err)

// 	assert.NoError(t, s.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

// }
