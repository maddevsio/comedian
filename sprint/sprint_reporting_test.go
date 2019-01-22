package sprint

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// func TestMakeActiveSprint(t *testing.T) {
// 	d := time.Date(2019, 01, 06, 0, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })
// 	activeSprint := ActiveSprint{
// 		TotalNumberOfTasks: 3,
// 		ResolvedTasksCount: 2,
// 		SprintDaysCount:    10,
// 		PassedDays:         5,
// 		URL:                "url_1",
// 	}
// 	task1 := Task{Issue: "issue_1", Status: "indeterminate", Assignee: "user_1"}
// 	var InProgressTasks []Task
// 	InProgressTasks = append(InProgressTasks, task1)
// 	activeSprint.InProgressTasks = InProgressTasks
// 	testCase := []struct {
// 		collectorInfo CollectorInfo
// 		activeSprint  ActiveSprint
// 	}{
// 		{CollectorInfo{
// 			SprintURL:   "url_1",
// 			SprintStart: "2019-01-01T16:29:02.432+06:00",
// 			SprintEnd:   "2019-01-11T16:29:00.000+06:00",
// 			Tasks: []struct {
// 				Issue    string `json:"title"`
// 				Status   string `json:"status"`
// 				Assignee string `json:"assignee"`
// 			}{
// 				{"issue_1", "indeterminate", "user_1"},
// 				{"issue_2", "done", "user_1"},
// 				{"issue_3", "done", "user_3"},
// 			},
// 		}, activeSprint},
// 	}
// 	for _, test := range testCase {
// 		actualS := MakeActiveSprint(test.collectorInfo)
// 		assert.Equal(t, test.activeSprint, actualS)
// 		log.Println("actualS: ", actualS.ResolvedTasksCount)
// 	}
// }

// func TestMakeMessage(t *testing.T) {
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	bot, err := bot.NewBot(c)
// 	assert.NoError(t, err)
// 	bot.CP.Language = "en_US"

// 	activeS1 := ActiveSprint{
// 		TotalNumberOfTasks: 10,
// 		ResolvedTasksCount: 5,
// 		SprintDaysCount:    10,
// 		PassedDays:         5,
// 		URL:                "/url/to/sprint",
// 	}
// 	var inProgressTasks1 []Task
// 	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue1", Status: "indeterminate", Assignee: "user1"})
// 	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue2", Status: "indeterminate", Assignee: "user1"})
// 	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue3", Status: "indeterminate", Assignee: "user2"})
// 	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue4", Status: "indeterminate", Assignee: "user2"})
// 	inProgressTasks1 = append(inProgressTasks1, Task{Issue: "issue5", Status: "indeterminate", Assignee: "user1"})
// 	activeS1.InProgressTasks = inProgressTasks1

// 	testCase := []struct {
// 		activeSprint ActiveSprint
// 		expected     string
// 	}{
// 		{activeS1, "Sprint completed: 50%\nSprint lasts 10 days,5 days passed, 5 days left.\nIn progress tasks: \nissue1 - user1;\nissue2 - user1;\nissue3 - user2;\nissue4 - user2;\nissue5 - user1;\n\nURL: /url/to/sprint"},
// 	}
// 	for _, test := range testCase {
// 		actual := MakeMessage(bot, test.activeSprint)
// 		assert.Equal(t, test.expected, actual)
// 	}
// }

// func TestMakeDate(t *testing.T) {

// }
