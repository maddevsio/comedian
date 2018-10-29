package reporting

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestStandupReportByProject(t *testing.T) {
	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	//First test when no data
	report, err := r.StandupReportByProject(channel, dateFrom, dateTo)
	assert.NoError(t, err)
	expected := "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, 0, len(report.ReportBody))

	d = time.Date(2018, 6, 4, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	//create user who did not write standup
	user1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	standup0, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		UserID:    user1.UserID,
		Comment:   "",
		MessageTS: "1234",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	//test for no standup submitted
	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)

	//create standup for user
	standup1, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "my standup",
		UserID:    user1.UserID,
		MessageTS: "123",
	})
	assert.NoError(t, err)

	//test if user submitted standup success
	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)
	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[1].Text)

	//create another user
	user2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n", report.ReportBody[0].Text)
	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[1].Text)

	//create standup for user2
	standup2, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "user2 standup",
		UserID:    "userID2",
		MessageTS: "1234",
	})
	assert.NoError(t, err)

	//test if both users had written standups
	report, err = r.StandupReportByProject(channel, dateFrom, dateTo)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-04\n<@userID1> did not submit standup!\n================================================\n<@userID2> submitted standup: user2 standup \n================================================\n", report.ReportBody[0].Text)
	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n================================================\n<@userID2> submitted standup: user2 standup \n================================================\n", report.ReportBody[1].Text)

	assert.NoError(t, r.db.DeleteStandup(standup0.ID))
	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
	assert.NoError(t, r.db.DeleteStandup(standup2.ID))
	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
	assert.NoError(t, r.db.DeleteChannelMember(user2.UserID, user2.ChannelID))
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
}

func TestStandupReportByUser(t *testing.T) {
	d := time.Date(2018, 6, 5, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	user, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	_, err = r.StandupReportByUser(user.UserID, dateTo, dateFrom)
	assert.Error(t, err)
	_, err = r.StandupReportByUser(user.UserID, dateNext, dateTo)
	assert.Error(t, err)
	_, err = r.StandupReportByUser(user.UserID, dateFrom, dateNext)
	assert.Error(t, err)

	expected := "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\n"
	report, err := r.StandupReportByUser(user.UserID, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, expected, report.ReportHead)

	standup1, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "my standup",
		UserID:    user.UserID,
		MessageTS: "123",
	})
	expected = "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\n"
	report, err = r.StandupReportByUser(user.UserID, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-05\nIn #chanName <@userID1> submitted standup: my standup \n================================================\n", report.ReportBody[0].Text)

	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
	assert.NoError(t, r.db.DeleteChannelMember(user.UserID, user.ChannelID))
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
}

func TestStandupReportByProjectAndUser(t *testing.T) {
	d := time.Date(2018, 6, 5, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	user1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channel.ChannelID,
	})

	report, err := r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo)
	assert.NoError(t, err)
	expected := "Report on user <@userID1> in project #chanName from 2018-06-03 to 2018-06-05\n\n"
	assert.Equal(t, expected, report.ReportHead)

	standup1, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "my standup",
		UserID:    "userID1",
		MessageTS: "123",
	})
	assert.NoError(t, err)

	report, err = r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo)
	assert.NoError(t, err)
	expected = "Report on user <@userID1> in project #chanName from 2018-06-03 to 2018-06-05\n\n"
	assert.Equal(t, expected, report.ReportHead)
	assert.Equal(t, "Report for: 2018-06-05\n<@userID1> submitted standup: my standup \n", report.ReportBody[0].Text)

	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
}

func TestYesterdayReport(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		standupComment                     string
		userID                             string
		collectorResponseUserBody          string
		collectorResponseUserInProjectBody string
		color                              string
		value                              string
	}{
		{"users", "userID1", `{"worklogs": 35000, "total_commits": 23}`,
			`{"worklogs": 35000, "total_commits": 23}`, "good",
			" worklogs: 9:43 :sunglasses: | commits: 23 :tada: | standup :heavy_check_mark: |\n"},
		{"", "userID2", `{"worklogs": 0, "total_commits": 0}`,
			`{"worklogs": 0, "total_commits": 0}`, "danger",
			" worklogs: 0:00 :angry: | commits: 0 :shit: | standup :x: |\n"},
		{"test", "userID3", `{"worklogs": 28800, "total_commits": 1}`,
			`{"worklogs": 28800, "total_commits": 1}`, "good",
			" worklogs: 8:00 :wink: | commits: 1 :tada: | standup :heavy_check_mark: |\n"},
		{"test", "userID4", `{"worklogs": 14400, "total_commits": 1}`,
			`{"worklogs": 12400, "total_commits": 1}`, "warning",
			" worklogs: 3:26 out of 4:00 :disappointed: | commits: 1 :tada: | standup :heavy_check_mark: |\n"},
	}

	for _, tt := range testCases {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		channel, err := r.db.CreateChannel(model.Channel{
			ChannelID:   "QWERTY123",
			ChannelName: "chanName1",
			StandupTime: int64(0),
		})
		assert.NoError(t, err)

		user, err := r.db.CreateUser(model.User{
			UserID:   tt.userID,
			UserName: "Alfa",
		})
		assert.NoError(t, err)

		channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
			UserID:    tt.userID,
			ChannelID: channel.ChannelID,
		})
		assert.NoError(t, err)

		standup, err := r.db.CreateStandup(model.Standup{
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
		httpmock.RegisterResponder("GET", linkURLUsers, httpmock.NewStringResponder(200, tt.collectorResponseUserBody))
		httpmock.RegisterResponder("GET", linkURLUserInProject, httpmock.NewStringResponder(200, tt.collectorResponseUserInProjectBody))

		attachments, err := r.generateReportForYesterday()
		assert.NoError(t, err)
		assert.Equal(t, tt.color, attachments[0].Color)
		assert.Equal(t, fmt.Sprintf("<@%v> in #%v", user.UserID, channel.ChannelName), attachments[0].Text)
		assert.Equal(t, tt.value, attachments[0].Fields[0].Value)

		assert.NoError(t, r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
		assert.NoError(t, r.db.DeleteStandup(standup.ID))
		assert.NoError(t, r.db.DeleteUser(user.ID))
		assert.NoError(t, r.db.DeleteChannel(channel.ID))
	}
}

func TestWeeklyReport(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		collectorResponseUserBody          string
		collectorResponseUserInProjectBody string
		value                              string
	}{
		{`{"worklogs": 35000, "total_commits": 23}`,
			`{"worklogs": 35000, "total_commits": 23}`,
			" worklogs: 9:43 | commits: 23 |\n"},
	}

	for _, tt := range testCases {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		channel, err := r.db.CreateChannel(model.Channel{
			ChannelID:   "QWERTY123",
			ChannelName: "chanName1",
			StandupTime: int64(0),
		})
		assert.NoError(t, err)

		user, err := r.db.CreateUser(model.User{
			UserID:   "userID",
			UserName: "user",
		})
		assert.NoError(t, err)

		channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
			UserID:    user.UserID,
			ChannelID: channel.ChannelID,
		})
		assert.NoError(t, err)

		d = time.Date(2018, 9, 16, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		beginDateOfRequest := fmt.Sprintf("%d-%02d-%02d", time.Now().AddDate(0, 0, -7).Year(), time.Now().AddDate(0, 0, -7).Month(), time.Now().AddDate(0, 0, -7).Day())
		endDateOfRequest := fmt.Sprintf("%d-%02d-%02d", time.Now().AddDate(0, 0, -1).Year(), time.Now().AddDate(0, 0, -1).Month(), time.Now().AddDate(0, 0, -1).Day())

		linkURLUsers := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "users", user.UserID, beginDateOfRequest, endDateOfRequest)
		fmt.Println(linkURLUsers)
		linkURLUserInProject := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "user-in-project", fmt.Sprintf("%v/%v", user.UserID, channel.ChannelName), beginDateOfRequest, endDateOfRequest)
		fmt.Println(linkURLUserInProject)

		httpmock.RegisterResponder("GET", linkURLUsers, httpmock.NewStringResponder(200, tt.collectorResponseUserBody))
		httpmock.RegisterResponder("GET", linkURLUserInProject, httpmock.NewStringResponder(200, tt.collectorResponseUserInProjectBody))

		attachments, err := r.generateReportForWeek()
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("<@%v> in #%v", user.UserID, channel.ChannelName), attachments[0].Text)
		assert.Equal(t, tt.value, attachments[0].Fields[0].Value)

		assert.NoError(t, r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
		assert.NoError(t, r.db.DeleteUser(user.ID))
		assert.NoError(t, r.db.DeleteChannel(channel.ID))
	}
}

func TestMondayReport(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		standupComment                     string
		userID                             string
		collectorResponseUserBody          string
		collectorResponseUserInProjectBody string
		value                              string
	}{
		{"users", "userID1", `{"worklogs": 35000, "total_commits": 23}`,
			`{"worklogs": 35000, "total_commits": 23}`,
			" worklogs: 9:43 | commits: 23 | standup :heavy_check_mark: |\n"},
		{"", "userID2", `{"worklogs": 0, "total_commits": 0}`,
			`{"worklogs": 0, "total_commits": 0}`,
			""},
		{"test", "userID3", `{"worklogs": 28800, "total_commits": 0}`,
			`{"worklogs": 28800, "total_commits": 0}`,
			" worklogs: 8:00 | standup :heavy_check_mark: |\n"},
		{"test", "userID4", `{"worklogs": 14400, "total_commits": 0}`,
			`{"worklogs": 12400, "total_commits": 0}`,
			" worklogs: 3:26 out of 4:00 | standup :heavy_check_mark: |\n"},
	}

	for _, tt := range testCases {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
		monkey.Patch(time.Now, func() time.Time { return d })

		channel, err := r.db.CreateChannel(model.Channel{
			ChannelID:   "QWERTY123",
			ChannelName: "chanName1",
			StandupTime: int64(0),
		})
		assert.NoError(t, err)

		user, err := r.db.CreateUser(model.User{
			UserID:   tt.userID,
			UserName: "Alfa",
		})
		assert.NoError(t, err)

		channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
			UserID:    tt.userID,
			ChannelID: channel.ChannelID,
		})
		assert.NoError(t, err)

		standup, err := r.db.CreateStandup(model.Standup{
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

		httpmock.RegisterResponder("GET", linkURLUsers, httpmock.NewStringResponder(200, tt.collectorResponseUserBody))
		httpmock.RegisterResponder("GET", linkURLUserInProject, httpmock.NewStringResponder(200, tt.collectorResponseUserInProjectBody))

		attachments, err := r.generateReportForSunday()

		assert.NoError(t, err)
		if len(attachments) == 0 {
			continue
		}
		assert.Equal(t, "good", attachments[0].Color)
		assert.Equal(t, fmt.Sprintf("<@%v> in #%v", user.UserID, channel.ChannelName), attachments[0].Text)
		assert.Equal(t, tt.value, attachments[0].Fields[0].Value)

		assert.NoError(t, r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
		assert.NoError(t, r.db.DeleteStandup(standup.ID))
		assert.NoError(t, r.db.DeleteUser(user.ID))
		assert.NoError(t, r.db.DeleteChannel(channel.ID))
	}
}
