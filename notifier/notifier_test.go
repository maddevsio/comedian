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

func (c *ChatStub) GetAllUsersToDB() error {
	return nil
}

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	ch := &ChatStub{}
	n, err := NewNotifier(c, ch)
	assert.NoError(t, err)

	c.ReportTime = "random time"
	err = n.Start(c)
	assert.Error(t, err)

	c, err = config.Get()
	assert.NoError(t, err)

	channelID := "QWERTY123"

	d := time.Date(2000, 12, 15, 17, 8, 00, 0, time.UTC)
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
		Time:      int64(12),
	})

	notifyStandupStart(ch, n.DB, channelID)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey! We are still waiting standup for today from you: <@user1>, <@user2>", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Друзья, я всё еще жду стэндапы от вас! Если их не будет, вам пизда: <@user1>, <@user2>", ch.LastMessage)
	}
	nonReporters, err := getNonReporters(ch, n.DB, channelID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nonReporters))

	standupReminderForChannel(ch, n.DB)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey! We are still waiting standup for today from you: <@user1>, <@user2>", ch.LastMessage) // херня какая-то
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Друзья, я всё еще жду стэндапы от вас! Если их не будет, вам пизда: <@user1>, <@user2>", ch.LastMessage)
	}

	managerStandupReport(ch, c, n.DB, d)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, in channel <#QWERTY123> not all standupers wrote standup today, this users ignored standup today: <@user1>, <@user2>.", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, в канале <#QWERTY123> не все написали стэндапы сегодня, игнорировали: <@user1>, <@user2>.", ch.LastMessage)
	}

	// add standup for user @test
	s, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "work hard",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "qweasdzxc",
	})
	assert.NoError(t, err)

	standupReminderForChannel(ch, n.DB)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, in channel <#QWERTY123> not all standupers wrote standup today, this users ignored standup today: <@user1>, <@user2>.", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, в канале <#QWERTY123> не все написали стэндапы сегодня, игнорировали: <@user1>, <@user2>.", ch.LastMessage)
	}

	managerStandupReport(ch, c, n.DB, d)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, in channel <#QWERTY123> not all standupers wrote standup today, this users ignored standup today: <@user2>.", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, в канале <#QWERTY123> не все написали стэндапы сегодня, игнорировали: <@user2>.", ch.LastMessage)
	}
	// add standup for user @user2
	s2, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "hello world",
		UsernameID: "userID2",
		Username:   "user2",
		MessageTS:  "qweasd",
	})
	assert.NoError(t, err)

	notifyStandupStart(ch, n.DB, channelID)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Congradulations! Everybody wrote their standups today!", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Поздравляю, сегодня все написали стэндапы!", ch.LastMessage)
	}
	standupReminderForChannel(ch, n.DB)

	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Congradulations! Everybody wrote their standups today!", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Поздравляю, сегодня все написали стэндапы!", ch.LastMessage)
	}

	managerStandupReport(ch, c, n.DB, d)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, in channel <#QWERTY123> all standupers have written standup today", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, в канале <#QWERTY123> все написали стэндапы сегодня", ch.LastMessage)
	}

	err = directRemindStandupers(ch, n.DB, channelID)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, in channel <#QWERTY123> all standupers have written standup today", ch.LastMessage)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "CHAT: CBAP453GV, MESSAGE: <@fedorenko.tolik>, в канале <#QWERTY123> все написали стэндапы сегодня", ch.LastMessage)
	}
	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su.SlackName, su.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su2.SlackName, su2.ChannelID))

	assert.NoError(t, n.DB.DeleteStandupTime(st.ChannelID))

	assert.NoError(t, n.DB.DeleteStandup(s.ID))
	assert.NoError(t, n.DB.DeleteStandup(s2.ID))
}
