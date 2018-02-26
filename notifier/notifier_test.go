package notifier

import (
	"fmt"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	ch := &ChatStub{LastMessage: "test initial"}
	n, err := NewNotifier(c, ch)
	assert.NoError(t, err)

	stubDate := time.Date(2000, 12, 15, 17, 8, 00, 0, time.UTC)
	channelID := "QWERTY123"
	managerStandupChannelID = channelID
	stubManagerName := "managerName"
	nowFunc = func() time.Time {
		return stubDate
	}
	storage.NowFunc = nowFunc

	su, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackName: "@test",
		FullName:  "Test Testtt",
		ChannelID: channelID,
		Channel:   "chanName",
	})
	assert.NoError(t, err)
	su2, err := n.DB.CreateStandupUser(model.StandupUser{
		SlackName: "@shmest",
		FullName:  "Test Testtt",
		ChannelID: channelID,
		Channel:   "chanName",
	})
	assert.NoError(t, err)

	_, err = n.DB.CreateStandupTime(model.StandupTime{
		ChannelID: channelID,
		Channel:   "chanName",
		Time:      stubDate.Unix(),
	})
	assert.NoError(t, err)

	standupReminderForChannel(ch, n.DB)
	assert.Equal(t, "test initial", ch.LastMessage)
	notifyStandupStart(ch, n.DB, channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey! We are still waiting standup from you: "+
		"@test, @shmest", ch.LastMessage)

	// add standup for user @test
	s, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "work hard",
		UsernameID: "QWE345asd",
		Username:   "@test",
		MessageTS:  "qweasdzxc",
	})
	assert.NoError(t, err)

	notifyNonReporters(ch, n.DB, channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: In this channel not all standupers wrote standup today, "+
		"shame on you: @shmest.", ch.LastMessage)

	// check that manager report prints @shmest
	managerStandupReport(ch, n.DB, stubManagerName, managerStandupChannelID, stubDate)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: managerName, in channel 'QWERTY123' not all standupers "+
		"wrote standup today, this users ignored standup today: @shmest.", ch.LastMessage)

	// add standup for user @shmest
	s2, err := n.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "hello world",
		UsernameID: "QWE345asd",
		Username:   "@shmest",
		MessageTS:  "qweasd",
	})
	assert.NoError(t, err)

	//nonreporters check
	notifyNonReporters(ch, n.DB, channelID)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: Hey, in this channel all standupers have "+
		"written standup today", ch.LastMessage)

	managerStandupReport(ch, n.DB, stubManagerName, managerStandupChannelID, stubDate)
	assert.Equal(t, "CHAT: QWERTY123, MESSAGE: managerName, in channel QWERTY123 all standupers "+
		"have written standup today", ch.LastMessage)

	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su.SlackName, su.ChannelID))
	assert.NoError(t, n.DB.DeleteStandupUserByUsername(su2.SlackName, su2.ChannelID))

	assert.NoError(t, n.DB.DeleteStandup(s.ID))
	assert.NoError(t, n.DB.DeleteStandup(s2.ID))
}
