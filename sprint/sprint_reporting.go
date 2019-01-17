package sprint

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
)

type Task struct {
	Issue    string
	Status   string
	Assignee string
}

type ActiveSprint struct {
	TotalNumberOfTasks int
	InProgressTasks    []Task
	ResolvedTasksCount int
	SprintDaysCount    int
	PassedDays         int
	URL                string
}

func prepareTime(timestring string) (date time.Time, err error) {
	//2019-01-14T16:29:02.432+06:00"
	parts := strings.Split(timestring, "T")
	var year, month, day int
	if len(parts) == 2 {
		//get year,month,day from time string
		ymt := strings.Split(parts[0], "-")
		if len(ymt) == 3 {
			year, _ = strconv.Atoi(ymt[0])
			month, _ = strconv.Atoi(ymt[1])
			day, _ = strconv.Atoi(ymt[2])
			date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			return date, err
		}
	}
	return date, errors.New("Error parsing time")
}

//countDays return count days between periods
func countDays(startDate, endDate time.Time) (countDays int) {
	countDays = int(endDate.Sub(startDate).Hours() / 24)
	return countDays
}

//MakeActiveSprint make ActiveSprint struct from CollectorInfo
func MakeActiveSprint(collectorInfo CollectorInfo) ActiveSprint {
	var activeSprint ActiveSprint
	activeSprint.URL = collectorInfo.SprintURL
	//calculate TotalNumberOfTasks
	activeSprint.TotalNumberOfTasks = len(collectorInfo.Tasks)
	for _, task := range collectorInfo.Tasks {
		//collect inprogress tasks
		var inProgressTask Task
		if task.Status == "indeterminate" {
			inProgressTask.Issue = task.Issue
			inProgressTask.Status = task.Status
			inProgressTask.Assignee = task.Assignee
			activeSprint.InProgressTasks = append(activeSprint.InProgressTasks, inProgressTask)
		}
		//collect done
		if task.Status == "done" {
			//increase count of resolved tasks
			activeSprint.ResolvedTasksCount++
		}
	}
	startDate, err := prepareTime(collectorInfo.SprintStart)
	if err != nil {
		logrus.Errorf("sprint.prepareTime failed: %v", err)
	}
	endDate, err := prepareTime(collectorInfo.SprintEnd)
	if err != nil {
		logrus.Errorf("sprint.prepareTime failed: %v", err)
	}
	activeSprint.SprintDaysCount = countDays(startDate, endDate)
	currentDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	activeSprint.PassedDays = countDays(startDate, currentDate)
	return activeSprint
}

//MakeMessage make message about sprint
func MakeMessage(bot *bot.Bot, activeSprint ActiveSprint) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.CP.Language)
	//progressSprint=resolved_tasks/totalCountOfTasks*100
	var totalMessage string
	progressSprint := float32(activeSprint.ResolvedTasksCount) / float32(activeSprint.TotalNumberOfTasks) * 100
	sprintProgressPercent := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SprintProgressPercent",
			Description: "Displays header message about sprint progress",
			Other:       "Sprint completed: {{.percent}}%",
		},
		TemplateData: map[string]interface{}{
			"percent": int(progressSprint),
		},
	})
	sprintProgressPercent += "\n"
	totalMessage += sprintProgressPercent
	leftDays := activeSprint.SprintDaysCount - activeSprint.PassedDays
	sprintTotalDays := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "sprintTotalDays",
			Description: "Displays total days of sprint",
			One:         "day",
			Other:       "days",
		},
		PluralCount: activeSprint.SprintDaysCount,
	})
	sprintPassedDays := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "sprintPassedDays",
			Description: "Displays passed days of sprint",
			One:         "day",
			Other:       "days",
		},
		PluralCount: activeSprint.PassedDays,
	})
	sprintLeftDays := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "sprintLeftDays",
			Description: "Displays left days of sprint",
			One:         "day",
			Other:       "days",
		},
		PluralCount: leftDays,
	})
	sprintDays := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SprintDays",
			Description: "Displays message about sprint duration",
			Other:       "Sprint lasts {{.totalDays}} {{.td}},{{.passedDays}} {{.pd}} passed, {{.leftDays}} {{.ld}} left.",
		},
		TemplateData: map[string]interface{}{
			"totalDays":  activeSprint.SprintDaysCount,
			"td":         sprintTotalDays,
			"passedDays": activeSprint.PassedDays,
			"pd":         sprintPassedDays,
			"leftDays":   leftDays,
			"ld":         sprintLeftDays,
		},
	})
	sprintDays += "\n"
	totalMessage += sprintDays
	inProgressTasksTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "InProgressTasksTitle",
			Description: "Displays title of message about in progress issues",
			One:         "In progress task: ",
			Other:       "In progress tasks: ",
		},
		PluralCount: len(activeSprint.InProgressTasks),
	})
	inProgressTasksTitle += "\n"
	totalMessage += inProgressTasksTitle
	for _, task := range activeSprint.InProgressTasks {
		inProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "InProgressTask",
				Description: "Displays message about in progress tasks",
				Other:       "{{.issue}} - {{.assignee}};",
			},
			TemplateData: map[string]interface{}{
				"issue":    task.Issue,
				"assignee": task.Assignee,
			},
		})
		inProgressTask += "\n"
		totalMessage += inProgressTask
	}
	totalMessage += "\n"
	urlTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "UrlTitle",
			Description: "Displays url of sprint",
			Other:       "URL: {{.url}}",
		},
		TemplateData: map[string]interface{}{
			"url": activeSprint.URL,
		},
	})
	totalMessage += urlTitle
	return totalMessage
}
