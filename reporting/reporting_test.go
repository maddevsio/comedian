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
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestCallDisplayYesterdayTeamReport(t *testing.T) {
	d := time.Date(2018, 11, 10, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

	testCase := []struct {
		reportTime string
	}{
		{""},
		{"10:00"},
	}
	for _, test := range testCase {
		bot.CP.ReportTime = test.reportTime
		r.CallDisplayYesterdayTeamReport()
	}
}

func TestCallDisplayWeeklyTeamReport(t *testing.T) {
	d := time.Date(2018, 12, 30, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

	testCase := []struct {
		reportTime string
	}{
		{""},
		{"10:00"},
	}
	for _, test := range testCase {
		bot.CP.ReportTime = test.reportTime
		r.CallDisplayWeeklyTeamReport()
	}
}

func TestDisplayYesterdayTeamReport(t *testing.T) {
	d := time.Date(2018, 11, 10, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.Language = "en_US"
	bot.CP.CollectorEnabled = true

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
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "chanId1",
	})
	assert.NoError(t, err)
	//creates channel without channel members
	channel2, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "chanId2",
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember1 doesn't obey send standup
	channelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user1, err := bot.DB.CreateUser(model.User{
		UserName: "username1",
		RealName: "realname1",
		UserID:   channelMember1.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
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
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	//creates channel members channel1
	//channelMember2 hasn't timetable and must sends standup
	channelMember2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user2, err := bot.DB.CreateUser(model.User{
		UserName: "username2",
		RealName: "realname2",
		UserID:   channelMember2.UserID,
		Role:     "",
	})
	assert.NoError(t, err)

	//create channel member channel1 without raw in users table
	channelMember3, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember4 doesn't obey send standup
	channelMember4, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid4",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user4, err := bot.DB.CreateUser(model.User{
		UserName: "username4",
		RealName: "realname4",
		UserID:   channelMember4.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable2, err := bot.DB.CreateTimeTable(model.TimeTable{
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
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCase := []struct {
		ExpectedReport string
	}{
		{"Yesterday report%!(EXTRA []slack.Attachment=[{good   0         realname2 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   } {good   0         realname4 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   } {warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }])"},
		//{good   0         realname2 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   }
		//channelMember2 must sends standup but IsNonReporter() return error
		//{warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }]
		//channelMember1 doesn't obey sends standup but he has not enough worklogs
		//{good   0         realname4 in #testChannel1   [{  worklogs: 7:46 :wink: | commits: 1 :wink: | false}] [] []   }
		//channelMember4 doen't obey send standup and he has enough worklogs
	}
	for _, test := range testCase {
		actualReport, _ := r.displayYesterdayTeamReport()
		assert.Equal(t, test.ExpectedReport, actualReport)
	}
	httpmock.DeactivateAndReset()
	//delete user
	err = bot.DB.DeleteUser(user1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(user2.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(user4.ID)
	assert.NoError(t, err)
	//delete timetables
	err = bot.DB.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteTimeTable(timeTable2.ID)
	assert.NoError(t, err)
	//delete channel members
	err = bot.DB.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember2.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember3.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember4.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = bot.DB.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
}
func TestDisplayWeeklyTeamReport(t *testing.T) {
	d := time.Date(2018, 11, 10, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

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
	channel1, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "chanId1",
	})
	assert.NoError(t, err)
	//creates channel without channel members
	channel2, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel2",
		ChannelID:   "chanId2",
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember1 doesn't obey send standup
	channelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user1, err := bot.DB.CreateUser(model.User{
		UserName: "username1",
		RealName: "realname1",
		UserID:   channelMember1.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
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
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	//creates channel members channel1
	//channelMember2 hasn't timetable and must sends standup
	channelMember2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user2, err := bot.DB.CreateUser(model.User{
		UserName: "username2",
		RealName: "realname2",
		UserID:   channelMember2.UserID,
		Role:     "",
	})
	assert.NoError(t, err)

	//create channel member channel1 without raw in users table
	channelMember3, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid3",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	//creates channel members of channel1
	//with timetable where Monday=0,Tuesday=0...Sunday=0
	//channelMember4 doesn't obey send standup
	channelMember4, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid4",
		ChannelID:     channel1.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	user4, err := bot.DB.CreateUser(model.User{
		UserName: "username4",
		RealName: "realname4",
		UserID:   channelMember4.UserID,
		Role:     "",
	})
	assert.NoError(t, err)
	timeTable2, err := bot.DB.CreateTimeTable(model.TimeTable{
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
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCase := []struct {
		ExpectedReport string
	}{
		{"Weekly report%!(EXTRA []slack.Attachment=[{good   0         realname2 in #testChannel1   [{  worklogs: 40:00 :sunglasses: | commits: 1 :wink: | false}] [] []   } {good   0         realname4 in #testChannel1   [{  worklogs: 40:00 :sunglasses: | commits: 1 :wink: | false}] [] []   } {warning   0         <@uid1> in #testChannel1   [{  worklogs: 5:33 :disappointed: | commits: 1 :wink: | false}] [] []   }])"},
	}
	for _, test := range testCase {
		actualReport, _ := r.displayWeeklyTeamReport()
		assert.Equal(t, test.ExpectedReport, actualReport)
	}
	httpmock.DeactivateAndReset()
	//delete user
	err = bot.DB.DeleteUser(user1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(user2.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteUser(user4.ID)
	assert.NoError(t, err)
	//delete timetables
	err = bot.DB.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteTimeTable(timeTable2.ID)
	assert.NoError(t, err)
	//delete channel members
	err = bot.DB.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember2.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember3.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember4.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	//delete channels
	err = bot.DB.DeleteChannel(channel1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannel(channel2.ID)
	assert.NoError(t, err)
}

func TestGetCollectorDataOnMember(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	d := time.Date(2018, 9, 18, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelID:   "testChannelID",
		ChannelName: "testChannel",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	channelMember, err := bot.DB.CreateChannelMember(model.ChannelMember{
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

	linkURLUsers := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, bot.TeamDomain, "users", channelMember.UserID, dateOfRequest, dateOfRequest)
	linkURLUserInProject := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, bot.TeamDomain, "user-in-project", fmt.Sprintf("%v/%v", channelMember.UserID, channel.ChannelName), dateOfRequest, dateOfRequest)

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

	assert.NoError(t, bot.DB.DeleteChannelMember(channelMember.UserID, channelMember.ChannelID))
	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
}

func TestProcessWorklogs(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

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
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

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
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	bot.CP.CollectorEnabled = true

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
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)
	bot.CP.CollectorEnabled = true

	member, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "testUserID",
		ChannelID:     "testChannelID",
		RoleInChannel: "developer",
	})
	assert.NoError(t, err)

	text, points := r.processStandup(member)
	assert.Equal(t, "", text)
	assert.Equal(t, 1, points)

	bot.DB.DeleteChannelMember(member.UserID, member.ChannelID)

}

func TestSweep(t *testing.T) {
	attachment := slack.Attachment{}
	entries := []model.AttachmentItem{
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
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	r, err := NewReporter(bot)
	assert.NoError(t, err)

	attachment := slack.Attachment{}
	entries := []model.AttachmentItem{
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
