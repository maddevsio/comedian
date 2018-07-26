package notifier

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

type ChatStub struct {
	LastMessage string
}

func (c *ChatStub) Run() error {
	return nil
}

func (c *ChatStub) SendMessage(chatID, message string) error {
	c.LastMessage = fmt.Sprintf("CHAT: %s, MESSAGE: %s", chatID, message)
	return nil
}

func (c *ChatStub) SendUserMessage(userID, message string) error {
	c.LastMessage = fmt.Sprintf("CHAT: %s, MESSAGE: %s", userID, message)
	return nil
}

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	ch := &ChatStub{}
	n, err := NewNotifier(c, ch)
	assert.NoError(t, err)

	channelID := "QWERTY123"

	d := time.Date(2018, 1, 1, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	su, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)
	su2, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     "chanName",
	})
	assert.NoError(t, err)

	st, err := n.DB.CreateStandupTime(model.StandupTime{
		ChannelID: channelID,
		Channel:   "chanName",
		Time:      time.Now().Unix(),
	})

	fmt.Printf("Standup time: %v", time.Unix(st.Time, 0))

	NonReporters, err := getNonReporters(n.DB, channelID)
	assert.NoError(t, err)
	assert.NotEmpty(t, NonReporters)
	assert.Equal(t, 2, len(NonReporters))

	n.SendWarning(channelID, NonReporters)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey! We are still waiting standup for today from you: <@userID1>, <@userID2>", ch.LastMessage)

	n.SendChannelNotification(channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: In this channel not all standupers wrote standup today, shame on you: userID1, userID2.", ch.LastMessage)

	n.NotifyChannels()
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: In this channel not all standupers wrote standup today, shame on you: userID1, userID2.", ch.LastMessage)

	d = time.Date(2018, 1, 1, 9, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	s, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "work hard",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "qweasdzxc",
	})
	assert.NoError(t, err)

	// add standup for user @user2
	s2, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "hello world",
		UsernameID: "userID2",
		Username:   "user2",
		MessageTS:  "qweasd",
	})

	n.SendChannelNotification(channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Congradulations! Everybody wrote their standups today!", ch.LastMessage)

	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su.SlackName, su.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su2.SlackName, su2.ChannelID))

	assert.NoError(t, n.DB.DeleteStandupTime(st.ChannelID))

	assert.NoError(t, n.DB.DeleteStandup(s.ID))
	assert.NoError(t, n.DB.DeleteStandup(s2.ID))
}
