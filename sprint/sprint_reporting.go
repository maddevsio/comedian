package sprint

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.com/team-monitoring/comedian/utils"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/reporting"
)

type Task struct {
	Issue            string
	Status           string
	Assignee         string
	AssigneeFullName string
	Link             string
}

type ActiveSprint struct {
	TotalNumberOfTasks    int
	InProgressTasks       []Task
	HasNotInProgressTasks []string
	ResolvedTasksCount    int
	SprintDaysCount       int
	PassedDays            int
	URL                   string
	StartDate             time.Time
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
	//members that have inprogress tasks
	var hasInProgressTasks []string
	for _, task := range collectorInfo.Tasks {
		//collect inprogress tasks
		var inProgressTask Task
		//inprogress tasks
		if task.Status == "indeterminate" {
			//add in hasTasks member that has task without doubling
			if !bot.InList(task.AssigneeFullName, hasInProgressTasks) {
				hasInProgressTasks = append(hasInProgressTasks, task.AssigneeFullName)
			}
			inProgressTask.Issue = task.Issue
			inProgressTask.Status = task.Status
			inProgressTask.Assignee = task.Assignee
			inProgressTask.AssigneeFullName = task.AssigneeFullName
			inProgressTask.Link = task.Link
			activeSprint.InProgressTasks = append(activeSprint.InProgressTasks, inProgressTask)
		}
		//collect done
		if task.Status == "done" {
			//increase count of resolved tasks
			activeSprint.ResolvedTasksCount++
		}
	}
	//check members to find members that hasn't inprogress tasks
	for _, task := range collectorInfo.Tasks {
		if !bot.InList(task.AssigneeFullName, hasInProgressTasks) {
			activeSprint.HasNotInProgressTasks = append(activeSprint.HasNotInProgressTasks, task.AssigneeFullName)
		}
	}
	startDate, err := prepareTime(collectorInfo.SprintStart)
	if err != nil {
		logrus.Errorf("sprint.prepareTime failed: %v", err)
	}
	activeSprint.StartDate = startDate
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
func MakeMessage(bot *bot.Bot, activeSprint ActiveSprint, project string, r reporting.Reporter) (string, error) {
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
	//sorts inprogress tasks by alphabet
	sort.Slice(activeSprint.InProgressTasks, func(i, j int) bool {
		return activeSprint.InProgressTasks[i].AssigneeFullName < activeSprint.InProgressTasks[j].AssigneeFullName
	})
	for _, task := range activeSprint.InProgressTasks {
		if task.AssigneeFullName != "" {
			inProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTask",
					Description: "Displays message about in progress tasks",
					Other:       "{{.i}} {{.issue}} - {{.assignee}};",
				},
				TemplateData: map[string]interface{}{
					"i":        "-",
					"issue":    task.Issue,
					"assignee": task.AssigneeFullName,
				},
			})
			inProgressTask += "\n"
			totalMessage += inProgressTask
			totalMessage += task.Link + "\n"
		} else if task.AssigneeFullName == "" {
			//if full name from jira empty, uses name
			inProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTask",
					Description: "Displays message about in progress tasks",
					Other:       "{{.i}} {{.issue}} - {{.assignee}};",
				},
				TemplateData: map[string]interface{}{
					"i":        "-",
					"issue":    task.Issue,
					"assignee": task.Assignee,
				},
			})
			inProgressTask += "\n"
			totalMessage += inProgressTask
			totalMessage += task.Link + "\n"
		}
	}
	totalMessage += "\n"
	//calculate total worklogs of team members
	var totalTeamWorlogs int
	channelID, err := bot.DB.GetChannelID(project)
	if err != nil {
		logrus.Infof("sprint_reporting:GetChannelID failed. ChannelName: %v", project)
		return "", err
	}
	channelMembers, err := bot.DB.ListChannelMembers(channelID)
	if err != nil {
		logrus.Errorf("sprint_reporting:ListChannelMembers failed: %v", err)
	}
	for _, channelMember := range channelMembers {
		_, worklogs, err := r.GetCollectorDataOnMember(channelMember, activeSprint.StartDate, time.Now())
		if err != nil {
			logrus.Infof("sprint_reporting:GetCollectorDataOnMember failed: %v. Member's user_id: %v", err, channelMember.UserID)
			continue
		}
		totalTeamWorlogs += worklogs.Worklogs
	}
	worklogs := utils.SecondsToHuman(totalTeamWorlogs)
	totalSprintWorklogsOfTeam := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "totalSprintWorklogsOfTeam",
			Description: "Displays message about total worklogs of team members",
			Other:       "Total worklogs of team on period from {{.start}} to {{.now}}:  {{.totalWorklogs}} (h)",
		},
		TemplateData: map[string]interface{}{
			"totalWorklogs": worklogs,
			"start":         MakeDate(activeSprint.StartDate),
			"now":           MakeDate(time.Now()),
		},
	})
	totalMessage += totalSprintWorklogsOfTeam
	totalMessage += "\n"
	urlTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "UrlTitle",
			Description: "Displays url of sprint",
			Other:       "Link to sprint: {{.url}}",
		},
		TemplateData: map[string]interface{}{
			"url": activeSprint.URL,
		},
	})
	totalMessage += urlTitle
	return totalMessage, nil
}

//MakeDate convert date to string
func MakeDate(date time.Time) string {
	y := date.Year()
	m := date.Month()
	d := date.Day()

	sdate := fmt.Sprintf("%v %v %v", d, m, y)
	return sdate
}
