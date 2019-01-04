package notifier

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)

	bot.CP.ReminderRepeatsMax = 0
	bot.CP.ReminderTime = 0
	bot.CP.NotifierInterval = 0
	assert.NoError(t, err)

	n, err := NewNotifier(bot)
	assert.NoError(t, err)

	channelID := "QWERTY123"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	su, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	su2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	assert.NoError(t, err)
	assert.NotEmpty(t, nonReporters)
	assert.Equal(t, 2, len(nonReporters))

	n.SendWarning(channelID)

	n.SendChannelNotification(channelID)

	n.NotifyChannels()

	d = time.Date(2018, 1, 2, 9, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	s, err := bot.DB.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		ChannelID: channelID,
		Comment:   "work hard",
		UserID:    "userID1",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	// add standup for user @user2
	s2, err := bot.DB.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		ChannelID: channelID,
		Comment:   "hello world",
		UserID:    "userID2",
		MessageTS: "qweasd",
	})

	d = time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	nonReporters, err = n.getCurrentDayNonReporters(channelID)
	assert.NoError(t, err)
	assert.Empty(t, nonReporters)

	n.SendChannelNotification(channelID)

	assert.NoError(t, bot.DB.DeleteChannelMember(su.UserID, su.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(su2.UserID, su2.ChannelID))

	assert.NoError(t, bot.DB.DeleteStandup(s.ID))
	assert.NoError(t, bot.DB.DeleteStandup(s2.ID))
}

func TestCheckUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)

	users, err := bot.DB.ListAllChannelMembers()
	assert.NoError(t, err)
	for _, user := range users {
		assert.NoError(t, bot.DB.DeleteChannelMember(user.UserID, user.ChannelID))
	}

	d := time.Date(2018, 6, 24, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channelID := "QWERTY123"
	err = bot.DB.CreateStandupTime(time.Now().Unix(), channelID)

	d = time.Date(2018, 6, 25, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	u1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	u2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: channelID,
	})
	assert.NoError(t, err)

	u3, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID3",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	u4, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userID4",
		ChannelID: channelID,
	})
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	testCases := []struct {
		title         string
		user          model.ChannelMember
		isNonReporter bool
		err           error
	}{
		{"test 1", u1, false, nil},
		{"test 2", u2, false, nil},
		{"test 3", u3, false, nil},
		{"test 4", u4, false, nil},
	}

	for _, tt := range testCases {
		isNonReporter, err := bot.DB.IsNonReporter(tt.user.UserID, tt.user.ChannelID, time.Now(), time.Now())
		assert.Error(t, err)
		assert.Equal(t, tt.isNonReporter, isNonReporter)
	}

	assert.NoError(t, bot.DB.DeleteChannelMember(u1.UserID, u1.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(u2.UserID, u2.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(u3.UserID, u3.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannelMember(u4.UserID, u4.ChannelID))

	assert.NoError(t, bot.DB.DeleteStandupTime(channelID))
}

func TestIndividualNotification(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	n, err := NewNotifier(bot)
	assert.NoError(t, err)

	d := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	user, err := bot.DB.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	m, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	tt, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: m.ID,
	})
	assert.NoError(t, err)
	timeNow := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	tt.Monday = timeNow.Unix()
	tt.Tuesday = timeNow.Unix()
	tt.Wednesday = timeNow.Unix()
	tt.Thursday = timeNow.Unix()
	tt.Friday = timeNow.Unix()

	tt, err = bot.DB.UpdateTimeTable(tt)
	assert.NoError(t, err)

	d = time.Date(2018, 10, 9, 15, 58, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send warning
	n.NotifyIndividuals()

	d = time.Date(2018, 10, 9, 16, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send push
	n.NotifyIndividuals()

	assert.NoError(t, bot.DB.DeleteUser(user.ID))
	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
	assert.NoError(t, bot.DB.DeleteChannelMember(user.UserID, channel.ChannelID))
	assert.NoError(t, bot.DB.DeleteTimeTable(tt.ID))

}

func TestChannelsNotification(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	n, err := NewNotifier(bot)
	assert.NoError(t, err)

	d := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	user, err := bot.DB.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	standupTime := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC).Unix()
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: standupTime,
	})
	assert.NoError(t, err)

	err = bot.DB.CreateStandupTime(standupTime, channel.ChannelID)
	assert.NoError(t, err)

	m, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	d = time.Date(2018, 10, 9, 15, 58, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send warning
	n.NotifyChannels()

	d = time.Date(2018, 10, 9, 16, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send push
	n.NotifyChannels()

	assert.NoError(t, bot.DB.DeleteUser(user.ID))
	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
	assert.NoError(t, bot.DB.DeleteChannelMember(m.UserID, m.ChannelID))
	assert.NoError(t, bot.DB.DeleteStandupTime(channel.ChannelID))

}

func TestSendIndividualWarning(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	n, err := NewNotifier(bot)
	assert.NoError(t, err)

	//creates channel member
	cm1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		ChannelID:     "cid1",
		UserID:        "uid1",
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	testCase := []struct {
		id int64
	}{
		//random id
		{1},
		{cm1.ID},
	}
	for _, test := range testCase {
		n.SendIndividualWarning(test.id)
	}
	//delete channel member
	err = bot.DB.DeleteChannelMember(cm1.UserID, cm1.ChannelID)
	assert.NoError(t, err)

}
