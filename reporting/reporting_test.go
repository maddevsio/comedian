package reporting

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/teammonitoring"
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

func TestGetCollectorDataOnMember(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelID:   "testChannelID",
		ChannelName: "testChannel",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "testUserID",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	d := time.Date(2018, 9, 18, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	startDate := time.Now().AddDate(0, 0, -1)
	endDate := time.Now().AddDate(0, 0, -1)

	testCases := []struct {
		totalWorklogs               int
		projectWorklogs             int
		commits                     int
		userRespStatusCode          int
		userInProjectRespStatusCode int
		collectorErr                error
	}{
		{35000, 3500, 20, 200, 200, nil},
		{0, 0, 0, 500, 200, errors.New("could not get data on this request")},
		{0, 0, 0, 200, 500, errors.New("could not get data on this request")},
	}

	dateOfRequest := fmt.Sprintf("%d-%02d-%02d", time.Now().AddDate(0, 0, -1).Year(), time.Now().AddDate(0, 0, -1).Month(), time.Now().AddDate(0, 0, -1).Day())

	linkURLUsers := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "users", channelMember.UserID, dateOfRequest, dateOfRequest)
	linkURLUserInProject := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, "user-in-project", fmt.Sprintf("%v/%v", channelMember.UserID, channel.ChannelName), dateOfRequest, dateOfRequest)

	for _, tt := range testCases {
		httpmock.RegisterResponder("GET", linkURLUsers, httpmock.NewStringResponder(tt.userRespStatusCode, fmt.Sprintf(`{"worklogs": %v, "total_commits": %v}`, tt.totalWorklogs, tt.commits)))
		httpmock.RegisterResponder("GET", linkURLUserInProject, httpmock.NewStringResponder(tt.userInProjectRespStatusCode, fmt.Sprintf(`{"worklogs": %v, "total_commits": %v}`, tt.projectWorklogs, tt.commits)))

		dataOnUser, dataOnUserInProject, err := r.GetCollectorDataOnMember(channelMember, channel, startDate, endDate)
		assert.Equal(t, tt.totalWorklogs, dataOnUser.Worklogs)
		assert.Equal(t, tt.commits, dataOnUser.TotalCommits)
		assert.Equal(t, tt.projectWorklogs, dataOnUserInProject.Worklogs)
		assert.Equal(t, tt.commits, dataOnUserInProject.TotalCommits)
		assert.Equal(t, tt.collectorErr, err)

		assert.NoError(t, r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
		assert.NoError(t, r.db.DeleteChannel(channel.ID))
	}

}
func TestPrepareAttachment(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		memberRole      string
		totalWorklogs   int
		projectWorklogs int
		commits         int
		isNonReporter   bool
		collectorError  error
		fieldValue      string
		points          int
	}{
		{"developer", 0, 0, 0, true, nil, " worklogs: 0:00 :angry: | commits: 0 :shit: | standup :x: |\n", 0},
		{"developer", 4000, 4000, 20, false, nil, " worklogs: 1:06 :angry: | commits: 20 :tada: | standup :heavy_check_mark: |\n", 2},
		{"developer", 14400, 4000, 20, false, nil, " worklogs: 1:06 out of 4:00 :disappointed: | commits: 20 :tada: | standup :heavy_check_mark: |\n", 2},
		{"developer", 28800, 28800, 20, false, nil, " worklogs: 8:00 :wink: | commits: 20 :tada: | standup :heavy_check_mark: |\n", 3},
		{"pm", 40000, 28800, 0, false, nil, " worklogs: 8:00 out of 11:06 :sunglasses: | standup :heavy_check_mark: |\n", 3},
		{"pm", 40000, 28800, 20, false, errors.New("anyErr"), " standup :heavy_check_mark: \n", 4},
	}

	for _, tt := range testCases {

		channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
			UserID:        "testUserID",
			ChannelID:     "testChannelID",
			RoleInChannel: tt.memberRole,
		})
		assert.NoError(t, err)

		userData := teammonitoring.CollectorData{tt.commits, tt.totalWorklogs}
		userInProjectData := teammonitoring.CollectorData{tt.commits, tt.projectWorklogs}

		fieldValue, points := r.PrepareAttachment(channelMember, userData, userInProjectData, tt.isNonReporter, tt.collectorError)
		assert.Equal(t, tt.fieldValue, fieldValue)
		assert.Equal(t, tt.points, points)

		r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID)

	}

}

func TestPrepareWeeklyAttachment(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		memberRole      string
		totalWorklogs   int
		projectWorklogs int
		commits         int
		isNonReporter   bool
		collectorError  error
		fieldValue      string
		points          int
	}{
		{"developer", 0, 0, 0, true, nil, " worklogs: 0:00 :disappointed: | commits: 0 :shit: |\n", 0},
		{"developer", 115200, 114200, 0, false, nil, " worklogs: 31:43 out of 32:00 :wink: | commits: 0 :shit: |\n", 1},
		{"developer", 129600, 129600, 20, false, nil, " worklogs: 36:00 :sunglasses: | commits: 20 :tada: |\n", 2},
		{"pm", 40000, 28800, 20, false, errors.New("anyErr"), "", 2},
	}

	for _, tt := range testCases {

		channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
			UserID:        "testUserID",
			ChannelID:     "testChannelID",
			RoleInChannel: tt.memberRole,
		})
		assert.NoError(t, err)

		userData := teammonitoring.CollectorData{tt.commits, tt.totalWorklogs}
		userInProjectData := teammonitoring.CollectorData{tt.commits, tt.projectWorklogs}

		fieldValue, points := r.PrepareWeeklyAttachment(channelMember, userData, userInProjectData, tt.collectorError)
		assert.Equal(t, tt.fieldValue, fieldValue)
		assert.Equal(t, tt.points, points)

		r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID)
	}

}
func TestGenerateAttachment(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	d := time.Date(2018, 11, 9, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases := []struct {
		fieldValue string
		points     int
		color      string
	}{
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 3, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 2, "warning"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 1, "warning"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 0, "danger"},
		{"", 0, "danger"},
	}

	for _, tt := range testCases {
		attachment := r.GenerateAttachment(tt.fieldValue, tt.points)
		assert.Equal(t, tt.color, attachment.Color)
		if len(attachment.Fields) != 0 {
			assert.Equal(t, tt.fieldValue, attachment.Fields[0].Value)
		} else {
			assert.Equal(t, 0, len(attachment.Fields))
		}
	}

	d = time.Date(2018, 11, 4, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases = []struct {
		fieldValue string
		points     int
		color      string
	}{
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 3, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 2, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 1, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 0, "good"},
		{"", 0, "good"},
	}

	for _, tt := range testCases {
		attachment := r.GenerateAttachment(tt.fieldValue, tt.points)
		assert.Equal(t, tt.color, attachment.Color)
		if len(attachment.Fields) != 0 {
			assert.Equal(t, tt.fieldValue, attachment.Fields[0].Value)
		} else {
			assert.Equal(t, 0, len(attachment.Fields))
		}
	}

	d = time.Date(2018, 11, 5, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases = []struct {
		fieldValue string
		points     int
		color      string
	}{
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 3, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 2, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 1, "good"},
		{"worklogs: 4:30 :disappointed: | commits: 2 :tada: | standup :heavy_check_mark: |", 0, "good"},
		{"", 0, "good"},
	}

	for _, tt := range testCases {
		attachment := r.GenerateAttachment(tt.fieldValue, tt.points)
		assert.Equal(t, tt.color, attachment.Color)
		if len(attachment.Fields) != 0 {
			assert.Equal(t, tt.fieldValue, attachment.Fields[0].Value)
		} else {
			assert.Equal(t, 0, len(attachment.Fields))
		}
	}
}

func TestGenerateReportAttachment(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	d := time.Date(2018, 11, 9, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserID",
		ChannelID:     "testChannelID",
		RoleInChannel: "",
	})
	assert.NoError(t, err)

	attachment := r.generateReportAttachment(channelMember, channel)
	assert.Equal(t, "", attachment.Text)
	assert.Equal(t, "warning", attachment.Color)
	if len(attachment.Fields) != 0 {
		assert.Equal(t, " standup :heavy_check_mark: \n", attachment.Fields[0].Value)
	}

	d = time.Date(2018, 11, 11, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	attachment = r.generateReportAttachment(channelMember, channel)
	assert.Equal(t, "", attachment.Text)
	assert.Equal(t, "", attachment.Color)
	if len(attachment.Fields) != 0 {
		assert.Equal(t, " standup :heavy_check_mark: \n", attachment.Fields[0].Value)
	}

	r.db.DeleteChannel(channel.ID)
	r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID)
}

func TestGenerateWeeklyReportAttachment(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	d := time.Date(2018, 11, 4, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	channelMember, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserID",
		ChannelID:     "testChannelID",
		RoleInChannel: "",
	})
	assert.NoError(t, err)

	attachment := r.generateWeeklyReportAttachment(channelMember, channel)
	assert.Equal(t, "", attachment.Text)
	assert.Equal(t, "", attachment.Color)
	if len(attachment.Fields) != 0 {
		assert.Equal(t, " standup :heavy_check_mark: \n", attachment.Fields[0].Value)
	}

	r.db.DeleteChannel(channel.ID)
	r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID)
}
