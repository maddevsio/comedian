package reporting

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/chat"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestDisplayYesterdayTeamReport(t *testing.T) {
	d := time.Date(2018, 11, 10, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	httpmock.Activate()
	//user1's worklogs, commits
	url1_1 := "www.collector.some/rest/api/v1/logger//users/uid1/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url1_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":20000}`))
	url1_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid1/testChannel1/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url1_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":20000}`))
	//user2's worlogs, commits
	url2_1 := "www.collector.some/rest/api/v1/logger//users/uid2/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url2_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28000}`))
	url2_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid2/testChannel1/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url2_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28000}`))
	//user4's worklogs, commits
	url4_1 := "www.collector.some/rest/api/v1/logger//users/uid4/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url4_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28000}`))
	url4_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid4/testChannel1/2018-11-09/2018-11-09/"
	httpmock.RegisterResponder("GET", url4_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28000}`))

	//creates channel with channel members
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "chanId1",
	})
	assert.NoError(t, err)
	//creates channel without channel members
	channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "chanId2",
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember1 doesn't obey send standup
	channelMember1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user1, err := r.db.CreateUser(model.User{
		UserName: "username1",
		UserID:   channelMember1.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable1, err := r.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: channelMember1.ID,
		Created:         time.Now(),
		Modified:        time.Now(),
		Monday:          0,
		Tuesday:         0,
		Wednesday:       0,
		Thursday:        0,
		Friday:          0,
		Saturday:        0,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	//creates channel members channel1
	//channelMember2 hasn't timetable and must sends standup
	channelMember2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user2, err := r.db.CreateUser(model.User{
		UserName: "username2",
		UserID:   channelMember2.UserID,
		Role:     "",
	})
	assert.NoError(t, err)

	//create channel member channel1 without raw in users table
	channelMember3, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember4 doesn't obey send standup
	channelMember4, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid4",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user4, err := r.db.CreateUser(model.User{
		UserName: "username4",
		UserID:   channelMember4.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable2, err := r.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: channelMember4.ID,
		Created:         time.Now(),
		Modified:        time.Now(),
		Monday:          0,
		Tuesday:         0,
		Wednesday:       0,
		Thursday:        0,
		Friday:          0,
		Saturday:        0,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCase := []struct {
		ExpectedReport string
	}{
		{"Yesterday report%!(EXTRA []slack.Attachment=[{good   0         username2 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   } {good   0         username4 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   } {warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }])"},
		//{good   0         username2 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   }
		//channelMember2 must sends standup but IsNonReporter() return error
		//{warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }]
		//channelMember1 doesn't obey sends standup but he has not enough worklogs
		//{good   0         username4 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   }
		//channelMember4 doen't obey send standup and he has enough worklogs
	}
	for _, test := range testCase {
		actualReport, _ := r.displayYesterdayTeamReport()
		assert.Equal(t, test.ExpectedReport, actualReport)
	}
	httpmock.DeactivateAndReset()
	//delete user
	err = r.db.DeleteUser(user1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(user2.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(user4.ID)
	assert.NoError(t, err)
	//delete timetables
	err = r.db.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteTimeTable(timeTable2.ID)
	assert.NoError(t, err)
	//delete channel members
	err = r.db.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember2.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember3.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember4.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = r.db.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
}

func TestDisplayWeeklyTeamReport(t *testing.T) {
	d := time.Date(2018, 11, 10, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	httpmock.Activate()
	//user1's worklogs, commits
	url1_1 := "www.collector.some/rest/api/v1/logger//users/uid1/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url1_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":20000}`))
	url1_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid1/testChannel1/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url1_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":20000}`))
	//user2's worlogs, commits
	url2_1 := "www.collector.some/rest/api/v1/logger//users/uid2/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url2_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":144000}`))
	url2_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid2/testChannel1/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url2_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":144000}`))
	//user4's worklogs, commits
	url4_1 := "www.collector.some/rest/api/v1/logger//users/uid4/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url4_1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":144000}`))
	url4_2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid4/testChannel1/2018-11-03/2018-11-09/"
	httpmock.RegisterResponder("GET", url4_2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":144000}`))

	//creates channel with channel members
	channel1, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "chanId1",
	})
	assert.NoError(t, err)
	//creates channel without channel members
	channel2, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "chanId2",
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember1 doesn't obey send standup
	channelMember1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user1, err := r.db.CreateUser(model.User{
		UserName: "username1",
		UserID:   channelMember1.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable1, err := r.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: channelMember1.ID,
		Created:         time.Now(),
		Modified:        time.Now(),
		Monday:          0,
		Tuesday:         0,
		Wednesday:       0,
		Thursday:        0,
		Friday:          0,
		Saturday:        0,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	//creates channel members channel1
	//channelMember2 hasn't timetable and must sends standup
	channelMember2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user2, err := r.db.CreateUser(model.User{
		UserName: "username2",
		UserID:   channelMember2.UserID,
		Role:     "",
	})
	assert.NoError(t, err)

	//create channel member channel1 without raw in users table
	channelMember3, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember4 doesn't obey send standup
	channelMember4, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid4",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user4, err := r.db.CreateUser(model.User{
		UserName: "username4",
		UserID:   channelMember4.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable2, err := r.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: channelMember4.ID,
		Created:         time.Now(),
		Modified:        time.Now(),
		Monday:          0,
		Tuesday:         0,
		Wednesday:       0,
		Thursday:        0,
		Friday:          0,
		Saturday:        0,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCase := []struct {
		ExpectedReport string
	}{
		{"Weekly report%!(EXTRA []slack.Attachment=[{good   0         username2 in #testChannel1   [{  worklogs: 40:00 :sunglasses: | commits: 1 :wink: | false}] [] []   } {good   0         username4 in #testChannel1   [{  worklogs: 40:00 :sunglasses: | commits: 1 :wink: | false}] [] []   } {warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }])"},
	}
	for _, test := range testCase {
		actualReport, _ := r.displayWeeklyTeamReport()
		assert.Equal(t, test.ExpectedReport, actualReport)
	}
	httpmock.DeactivateAndReset()
	//delete user
	err = r.db.DeleteUser(user1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(user2.ID)
	assert.NoError(t, err)
	err = r.db.DeleteUser(user4.ID)
	assert.NoError(t, err)
	//delete timetables
	err = r.db.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteTimeTable(timeTable2.ID)
	assert.NoError(t, err)
	//delete channel members
	err = r.db.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember2.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember3.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(channelMember4.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = r.db.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = r.db.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
}

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

	d := time.Date(2018, 9, 18, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

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

		dataOnUser, dataOnUserInProject, err := r.GetCollectorDataOnMember(channelMember, startDate, endDate)
		assert.Equal(t, tt.totalWorklogs, dataOnUser.Worklogs)
		assert.Equal(t, tt.commits, dataOnUser.Commits)
		assert.Equal(t, tt.projectWorklogs, dataOnUserInProject.Worklogs)
		assert.Equal(t, tt.commits, dataOnUserInProject.Commits)
		assert.Equal(t, tt.collectorErr, err)
	}

	assert.NoError(t, r.db.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
}
func TestProcessWorklogs(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		totalWorklogs   int
		projectWorklogs int
		textOutput      string
		points          int
	}{
		{3600, 0, " worklogs: 0:00 out of 1:00 :angry: |", 0},
		{4 * 3600, 3600, " worklogs: 1:00 out of 4:00 :disappointed: |", 0},
		{8 * 3600, 3600, " worklogs: 1:00 out of 8:00 :wink: |", 1},
		{10 * 3600, 3600, " worklogs: 1:00 out of 10:00 :sunglasses: |", 1},
	}

	for _, tt := range testCases {
		text, points := r.processWorklogs(tt.totalWorklogs, tt.projectWorklogs)
		assert.Equal(t, tt.textOutput, text)
		assert.Equal(t, tt.points, points)
	}

	d := time.Date(2018, 11, 18, 10, 0, 0, 0, time.UTC)
	pg := monkey.Patch(time.Now, func() time.Time { return d })

	testCases = []struct {
		totalWorklogs   int
		projectWorklogs int
		textOutput      string
		points          int
	}{
		{3600, 0, "", 0},
	}

	for _, tt := range testCases {
		text, points := r.processWorklogs(tt.totalWorklogs, tt.projectWorklogs)
		assert.Equal(t, tt.textOutput, text)
		assert.Equal(t, tt.points, points)
	}

	pg.Unpatch()
}

func TestProcessWeeklyWorklogs(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		totalWorklogs   int
		projectWorklogs int
		textOutput      string
		points          int
	}{
		{30 * 3600, 30 * 3600, " worklogs: 30:00 :disappointed: |", 0},
		{33 * 3600, 33 * 3600, " worklogs: 33:00 :wink: |", 1},
		{36 * 3600, 36 * 3600, " worklogs: 36:00 :sunglasses: |", 1},
		{30 * 3600, 0, " worklogs: 0:00 out of 30:00 :disappointed: |", 0},
		{33 * 3600, 0, " worklogs: 0:00 out of 33:00 :wink: |", 1},
		{36 * 3600, 0, " worklogs: 0:00 out of 36:00 :sunglasses: |", 1},
	}

	for _, tt := range testCases {
		text, points := r.processWeeklyWorklogs(tt.totalWorklogs, tt.projectWorklogs)
		assert.Equal(t, tt.textOutput, text)
		assert.Equal(t, tt.points, points)
	}
}

func TestProcessCommits(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	testCases := []struct {
		totalCommits   int
		projectCommits int
		textOutput     string
		points         int
	}{
		{0, 0, " commits: 0 :shit: |", 0},
		{1, 1, " commits: 1 :wink: |", 1},
	}

	for _, tt := range testCases {
		text, points := r.processCommits(tt.totalCommits, tt.projectCommits)
		assert.Equal(t, tt.textOutput, text)
		assert.Equal(t, tt.points, points)
	}

	d := time.Date(2018, 11, 18, 10, 0, 0, 0, time.UTC)
	pg := monkey.Patch(time.Now, func() time.Time { return d })

	testCases = []struct {
		totalCommits   int
		projectCommits int
		textOutput     string
		points         int
	}{
		{0, 0, "", 0},
		{1, 1, " commits: 1  |", 1},
	}

	for _, tt := range testCases {
		text, points := r.processCommits(tt.totalCommits, tt.projectCommits)
		assert.Equal(t, tt.textOutput, text)
		assert.Equal(t, tt.points, points)
	}

	pg.Unpatch()
}

func TestProcessStandup(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	member, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserID",
		ChannelID:     "testChannelID",
		RoleInChannel: "developer",
	})
	assert.NoError(t, err)

	text, points := r.processStandup(member)
	assert.Equal(t, "", text)
	assert.Equal(t, 1, points)

	r.db.DeleteChannelMember(member.UserID, member.ChannelID)

}

func TestSweep(t *testing.T) {
	attachment := slack.Attachment{}
	entries := []AttachmentItem{
		{attachment, 0},
		{attachment, 3},
		{attachment, 1},
		{attachment, 20},
		{attachment, 21},
		{attachment, 50},
	}

	for i := 0; i < len(entries); i++ {
		if !sweep(entries, i) {
			fmt.Println(entries)
			break
		}
	}
}

func TestSortReportEntries(t *testing.T) {

	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r := NewReporter(s)

	attachment := slack.Attachment{}
	entries := []AttachmentItem{
		{attachment, 0},
		{attachment, 3},
		{attachment, 1},
		{attachment, 20},
		{attachment, 21},
		{attachment, 50},
	}

	sorted := r.sortReportEntries(entries)
	fmt.Println(sorted)
}
