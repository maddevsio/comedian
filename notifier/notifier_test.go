package notifier

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestNotifier(t *testing.T) {
	c, err := config.Get()
	c.ReminderRepeatsMax = 0
	c.ReminderTime = 0
	c.NotifierInterval = 0
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	n, err := NewNotifier(c, slack)
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

	su, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	su2, err := n.db.CreateChannelMember(model.ChannelMember{
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

	s, err := n.db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		ChannelID: channelID,
		Comment:   "work hard",
		UserID:    "userID1",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	// add standup for user @user2
	s2, err := n.db.CreateStandup(model.Standup{
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

	assert.NoError(t, n.db.DeleteChannelMember(su.UserID, su.ChannelID))
	assert.NoError(t, n.db.DeleteChannelMember(su2.UserID, su2.ChannelID))

	assert.NoError(t, n.db.DeleteStandup(s.ID))
	assert.NoError(t, n.db.DeleteStandup(s2.ID))
}

func TestCheckUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	n, err := NewNotifier(c, slack)
	assert.NoError(t, err)

	users, err := n.db.ListAllChannelMembers()
	assert.NoError(t, err)
	for _, user := range users {
		assert.NoError(t, n.db.DeleteChannelMember(user.UserID, user.ChannelID))
	}

	d := time.Date(2018, 6, 24, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channelID := "QWERTY123"
	err = n.db.CreateStandupTime(time.Now().Unix(), channelID)

	d = time.Date(2018, 6, 25, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	u1, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	u2, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: channelID,
	})
	assert.NoError(t, err)

	u3, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID3",
		ChannelID: channelID,
	})
	assert.NoError(t, err)
	u4, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID4",
		ChannelID: channelID,
	})
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID1/2018-06-25/2018-06-25/", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 2, "total_merges": 1, "worklogs": 100000}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID2/2018-06-25/2018-06-25/", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 30, "total_merges": 0, "worklogs": 13600}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID3/2018-06-25/2018-06-25/", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 0, "total_merges": 0, "worklogs": 0}`))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/rest/api/v1/logger/users/userID4/2018-06-25/2018-06-25/", c.CollectorURL),
		httpmock.NewStringResponder(200, `{"total_commits": 20, "total_merges": 0, "worklogs": 50000}`))

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
		isNonReporter, err := n.db.IsNonReporter(tt.user.UserID, tt.user.ChannelID, time.Now(), time.Now())
		assert.Error(t, err)
		assert.Equal(t, tt.isNonReporter, isNonReporter)
	}

	assert.NoError(t, n.db.DeleteChannelMember(u1.UserID, u1.ChannelID))
	assert.NoError(t, n.db.DeleteChannelMember(u2.UserID, u2.ChannelID))
	assert.NoError(t, n.db.DeleteChannelMember(u3.UserID, u3.ChannelID))
	assert.NoError(t, n.db.DeleteChannelMember(u4.UserID, u4.ChannelID))

	assert.NoError(t, n.db.DeleteStandupTime(channelID))
}

func TestIndividualNotification(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	n, err := NewNotifier(c, slack)
	assert.NoError(t, err)

	d := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	user, err := n.db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := n.db.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	m, err := n.db.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	tt, err := n.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: m.ID,
	})
	assert.NoError(t, err)
	timeNow := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	tt.Monday = timeNow.Unix()
	tt.Tuesday = timeNow.Unix()
	tt.Wednesday = timeNow.Unix()
	tt.Thursday = timeNow.Unix()
	tt.Friday = timeNow.Unix()

	tt, err = n.db.UpdateTimeTable(tt)
	assert.NoError(t, err)

	d = time.Date(2018, 10, 9, 15, 58, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send warning
	n.NotifyIndividuals()

	d = time.Date(2018, 10, 9, 16, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	//send push
	n.NotifyIndividuals()

	assert.NoError(t, n.db.DeleteUser(user.ID))
	assert.NoError(t, n.db.DeleteChannel(channel.ID))
	assert.NoError(t, n.db.DeleteChannelMember(user.UserID, channel.ChannelID))
	assert.NoError(t, n.db.DeleteTimeTable(tt.ID))

}

func TestChannelsNotification(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage",
		httpmock.NewStringResponder(200, `{"OK": true}`))

	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	n, err := NewNotifier(c, slack)
	assert.NoError(t, err)

	d := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	user, err := n.db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	standupTime := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC).Unix()
	channel, err := n.db.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: standupTime,
	})
	assert.NoError(t, err)

	err = n.db.CreateStandupTime(standupTime, channel.ChannelID)
	assert.NoError(t, err)

	m, err := n.db.CreateChannelMember(model.ChannelMember{
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

	assert.NoError(t, n.db.DeleteUser(user.ID))
	assert.NoError(t, n.db.DeleteChannel(channel.ID))
	assert.NoError(t, n.db.DeleteChannelMember(m.UserID, m.ChannelID))
	assert.NoError(t, n.db.DeleteStandupTime(channel.ChannelID))

}
