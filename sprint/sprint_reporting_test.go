package sprint

import (
	"errors"
	"testing"
	"time"

	"gitlab.com/team-monitoring/comedian/model"

	"github.com/bouk/monkey"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/reporting"
)

func TestPrepareTime(t *testing.T) {
	testCase := []struct {
		timestring string
		expected   time.Time
		err        error
	}{
		{"2019-01-14T16:29:02.432+06:00", time.Date(2019, 01, 14, 0, 0, 0, 0, time.UTC), nil},
		{"2019-01-15T16:29:02.432+06:00", time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC), nil},
		{"2019-01T16:29:02.432+06:00", time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC), errors.New("Error parsing time")},
	}
	for _, test := range testCase {
		actual, err := prepareTime(test.timestring)
		assert.Equal(t, test.expected, actual)
		assert.Equal(t, test.err, err)
	}
}

func TestCountDays(t *testing.T) {
	testCase := []struct {
		start    time.Time
		end      time.Time
		expCount int
	}{
		{time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), time.Date(2019, 01, 11, 0, 0, 0, 0, time.UTC), 10},
		{time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), time.Date(2019, 01, 21, 0, 0, 0, 0, time.UTC), 20},
		{time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), time.Date(2019, 01, 02, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), 0},
	}
	for _, test := range testCase {
		actCount := countDays(test.start, test.end)
		assert.Equal(t, test.expCount, actCount)
	}
}

func TestMakeActiveSprint(t *testing.T) {
	d := time.Date(2019, 01, 06, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	startDate, err := prepareTime("2019-01-01T16:29:02.432+06:00")
	assert.NoError(t, err)
	activeSprint := ActiveSprint{
		TotalNumberOfTasks: 3,
		ResolvedTasksCount: 2,
		SprintDaysCount:    10,
		PassedDays:         5,
		URL:                "url_1",
		StartDate:          startDate,
	}
	task1 := Task{Issue: "issue_1", Status: "indeterminate", Assignee: "user_1", AssigneeFullName: "fullName1", Link: "link to task1"}
	var InProgressTasks []Task
	InProgressTasks = append(InProgressTasks, task1)
	activeSprint.InProgressTasks = InProgressTasks
	testCase := []struct {
		collectorInfo CollectorInfo
		activeSprint  ActiveSprint
	}{
		{CollectorInfo{
			SprintURL:   "url_1",
			SprintStart: "2019-01-01T16:29:02.432+06:00",
			SprintEnd:   "2019-01-11T16:29:00.000+06:00",
			Tasks: []struct {
				Issue            string `json:"title"`
				Status           string `json:"status"`
				Assignee         string `json:"assignee"`
				AssigneeFullName string `json:"assignee_name"`
				Link             string `json:"link"`
			}{
				{"issue_1", "indeterminate", "user_1", "fullName1", "link to task1"},
				{"issue_2", "done", "user_1", "fullName1", "link to task2"},
				{"issue_3", "done", "user_3", "fullName3", "link to task3"},
			},
		}, activeSprint},
	}
	for _, test := range testCase {
		actualS := MakeActiveSprint(test.collectorInfo)
		assert.Equal(t, test.activeSprint, actualS)
	}
}

func TestMakeMessage(t *testing.T) {
	d := time.Date(2019, 01, 05, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	bot.CP.Language = "en_US"
	bot.CP.CollectorEnabled = true

	r, err := reporting.NewReporter(bot)
	assert.NoError(t, err)

	startDate, err := prepareTime("2019-01-01T16:29:02.432+06:00")
	assert.NoError(t, err)
	activeS1 := ActiveSprint{
		TotalNumberOfTasks: 10,
		ResolvedTasksCount: 5,
		SprintDaysCount:    10,
		PassedDays:         5,
		URL:                "/url/to/sprint",
		StartDate:          startDate,
	}
	var inProgressTasks1 []Task
	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue1", Status: "indeterminate", Assignee: "user1", AssigneeFullName: "Bob", Link: "link_to_task1"})
	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue2", Status: "indeterminate", Assignee: "user1", AssigneeFullName: "John", Link: "link_to_task2"})
	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue3", Status: "indeterminate", Assignee: "user2", AssigneeFullName: "Frank", Link: "link_to_task3"})
	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue4", Status: "indeterminate", Assignee: "user2", AssigneeFullName: "Clark", Link: "link_to_task4"})
	//if fullname is empty than name must used instead
	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue5", Status: "indeterminate", Assignee: "user1", Link: "link_to_task5"})
	activeS1.InProgressTasks = inProgressTasks1

	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelID:   "cid2",
		ChannelName: "project",
	})
	assert.NoError(t, err)
	channelMember1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "uid",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)
	channelMember2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:    "userid",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)
	httpmock.Activate()
	url1 := "www.collector.some/rest/api/v1/logger//users/uid/2019-01-01/2019-01-05/"
	httpmock.RegisterResponder("GET", url1, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28800}`))
	url2 := "www.collector.some/rest/api/v1/logger//user-in-project/uid/project/2019-01-01/2019-01-05/"
	httpmock.RegisterResponder("GET", url2, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28800}`))
	url3 := "www.collector.some/rest/api/v1/logger//users/userid/2019-01-01/2019-01-05/"
	httpmock.RegisterResponder("GET", url3, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28800}`))
	url4 := "www.collector.some/rest/api/v1/logger//user-in-project/userid/project/2019-01-01/2019-01-05/"
	httpmock.RegisterResponder("GET", url4, httpmock.NewStringResponder(200, `{"total_commits":1,"worklogs":28800}`))

	defer httpmock.DeactivateAndReset()
	testCase := []struct {
		activeSprint ActiveSprint
		expected     string
	}{
		{activeS1, "Sprint completed: 50%\nSprint lasts 10 days,5 days passed, 5 days left.\nIn progress tasks: \n- issue5 - user1;\nlink_to_task5\n- issue1 - Bob;\nlink_to_task1\n- issue4 - Clark;\nlink_to_task4\n- issue3 - Frank;\nlink_to_task3\n- issue2 - John;\nlink_to_task2\n\nTotal worklogs of team on period from 1 January 2019 to 5 January 2019:  0:00 (h)\nLink to sprint: /url/to/sprint"},
	}
	for _, test := range testCase {
		actual, err := MakeMessage(bot, test.activeSprint, "project", r)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}

	//delete channel members
	err = bot.DB.DeleteChannelMember(channelMember1.UserID, channelMember1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(channelMember2.UserID, channelMember2.ChannelID)
	assert.NoError(t, err)
	//delete channel
	err = bot.DB.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestMakeDate(t *testing.T) {
	testCase := []struct {
		date     time.Time
		expected string
	}{
		{time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC), "1 January 2019"},
		{time.Date(2019, 02, 20, 0, 0, 0, 0, time.UTC), "20 February 2019"},
	}
	for _, test := range testCase {
		actual := MakeDate(test.date)
		assert.Equal(t, test.expected, actual)
	}
}
