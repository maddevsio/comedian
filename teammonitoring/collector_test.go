package teammonitoring

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestTeamMonitoringIsOFF(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	c.TeamMonitoringEnabled = false
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	_, err = NewTeamMonitoring(slack)
	assert.Error(t, err)
	assert.Equal(t, "team monitoring is disabled", err.Error())
}

func TestTeamMonitoringOnWeekEnd(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)

	tm, err := NewTeamMonitoring(slack)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 16, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	_, err = tm.RevealRooks()
	assert.Error(t, err)
	assert.Equal(t, "Day off today! Next report on Monday!", err.Error())
	tm.reportRooks()
}

func TestTeamMonitoringOnMonday(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)

	tm, err := NewTeamMonitoring(s)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	attachments, err := tm.RevealRooks()
	assert.NoError(t, err)
	assert.Equal(t, []slack.Attachment{}, attachments)
	tm.reportRooks()
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

func TestGetCollectorData(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)

	testCases := []struct {
		getDataOn string
		data      string
		dateFrom  string
		dateTo    string
	}{
		{"users", "U851AU1U0", "2018-10-12", "2018-10-14"},
		{"projects", "comedian-testing", "2018-10-11", "2018-10-11"},
		{"user-in-project", "UC1JNECA3/comedian-testing", "2018-10-11", "2018-10-11"},
		{"user-in-project", "UD6143K51/standups", "2018-10-12", "2018-10-14"},
		{"user-in-project", "UD6147Z4K/standups", "2018-10-12", "2018-10-14"},
	}

	for _, tt := range testCases {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		url := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, ""))
		result, err := GetCollectorData(c, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		assert.NoError(t, err)
		fmt.Printf("Report on user: Total Commits: %v, Total Worklogs: %v\n\n", result.TotalCommits, utils.SecondsToHuman(result.Worklogs))
	}
}
