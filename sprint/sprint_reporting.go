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
	"github.com/nlopes/slack"
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
	Name                  string
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
	activeSprint.Name = collectorInfo.SprintName
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
			if !bot.InList(task.AssigneeFullName, activeSprint.HasNotInProgressTasks) {
				activeSprint.HasNotInProgressTasks = append(activeSprint.HasNotInProgressTasks, task.AssigneeFullName)
			}
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
func MakeMessage(bot *bot.Bot, activeSprint ActiveSprint, project string, r reporting.Reporter) (string, []slack.Attachment, error) {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.CP.Language)
	//progressSprint=resolved_tasks/totalCountOfTasks*100
	var attachments []slack.Attachment
	var attachment1 slack.Attachment
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
	attachment1.Title = sprintProgressPercent
	sprintTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SprintTitle",
			Description: "Displays name of sprint",
			Other:       "*{{.name}} report*",
		},
		TemplateData: map[string]interface{}{
			"name": activeSprint.Name,
		},
	})
	attachment1.Pretext = sprintTitle
	attachment1.MarkdownIn = append(attachment1.MarkdownIn, attachment1.Pretext)
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
			Other:       "Sprint lasts {{.totalDays}} {{.td}}, {{.passedDays}} {{.pd}} passed, {{.leftDays}} {{.ld}} left.",
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
	attachment1.Text = sprintDays
	urlTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "UrlTitle",
			Description: "Displays url of sprint",
			Other:       "{{.url}}",
		},
		TemplateData: map[string]interface{}{
			"url": activeSprint.URL,
		},
	})
	attachment1.Text += urlTitle
	attachments = append(attachments, attachment1)
	var attachment2 slack.Attachment
	inProgressTasksTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "InProgressTasksTitle",
			Description: "Displays title of message about in progress issues",
			One:         "*In progress task:* ",
			Other:       "*In progress tasks:* ",
		},
		PluralCount: len(activeSprint.InProgressTasks),
	})
	attachment2.Pretext = inProgressTasksTitle
	attachment2.MarkdownIn = append(attachment2.MarkdownIn, attachment2.Pretext)
	attachments = append(attachments, attachment2)
	//sorts inprogress tasks by alphabet
	sort.Slice(activeSprint.InProgressTasks, func(i, j int) bool {
		return activeSprint.InProgressTasks[i].AssigneeFullName < activeSprint.InProgressTasks[j].AssigneeFullName
	})
	for _, task := range activeSprint.InProgressTasks {
		var attachment slack.Attachment
		if task.AssigneeFullName != "" {
			inProgressTaskTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTaskTitle",
					Description: "Displays asignee name as attachment title",
					Other:       "{{.assignee}}",
				},
				TemplateData: map[string]interface{}{
					"assignee": task.AssigneeFullName,
				},
			})
			attachment.Title = inProgressTaskTitle
			attachment.Color = "#2eb886"
			inProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTask",
					Description: "Displays message about in progress tasks",
					Other:       "{{.issue}};",
				},
				TemplateData: map[string]interface{}{
					"issue": task.Issue,
				},
			})
			inProgressTask += "\n"
			attachment.Text = inProgressTask
			attachment.Text += task.Link
			attachments = append(attachments, attachment)
		} else if task.AssigneeFullName == "" {
			//if full name from jira empty, uses name
			inProgressTaskTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTaskTitle",
					Description: "Displays asignee name as attachment title",
					Other:       "{{.assignee}}",
				},
				TemplateData: map[string]interface{}{
					"assignee": task.Assignee,
				},
			})
			attachment.Title = inProgressTaskTitle
			attachment.Color = "#2eb886"
			inProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "InProgressTask",
					Description: "Displays message about in progress tasks",
					Other:       "{{.issue}};",
				},
				TemplateData: map[string]interface{}{
					"issue": task.Issue,
				},
			})
			inProgressTask += "\n"
			attachment.Text = inProgressTask
			attachment.Text += task.Link
			attachments = append(attachments, attachment)
		}
	}
	var attachment3 slack.Attachment
	notInProgressTaskTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "notInProgressTaskTitle",
			Description: "Displays title of message about members that has not inprogress tasks",
			Other:       "*No tasks In Progress:*",
		},
	})
	attachment3.Pretext = notInProgressTaskTitle
	attachment3.MarkdownIn = append(attachment3.MarkdownIn, attachment3.Pretext)
	attachments = append(attachments, attachment3)
	if len(activeSprint.HasNotInProgressTasks) > 0 {
		var attachment slack.Attachment
		for _, task := range activeSprint.HasNotInProgressTasks {
			//members that has not inprogress tasks
			notInProgressTask := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "notInProgressTask",
					Description: "Displays members that has not in progress tasks",
					Other:       "{{.name}}",
				},
				TemplateData: map[string]interface{}{
					"name": task,
				},
			})
			attachment.Title = notInProgressTask
			attachment.Color = "danger"
			attachments = append(attachments, attachment)
		}
	}
	//calculate total worklogs of team members
	var totalTeamWorlogs int
	channelID, err := bot.DB.GetChannelID(project)
	if err != nil {
		logrus.Infof("sprint_reporting:GetChannelID failed. ChannelName: %v", project)
		return "", attachments, err
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
	var attachment4 slack.Attachment
	totalSprintWorklogsOfTeam := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "totalSprintWorklogsOfTeam",
			Description: "Displays message about total worklogs of team members",
			Other:       "*Total worklogs of team on period from {{.start}} to {{.now}}:*  {{.totalWorklogs}} (h)",
		},
		TemplateData: map[string]interface{}{
			"totalWorklogs": worklogs,
			"start":         MakeDate(activeSprint.StartDate),
			"now":           MakeDate(time.Now()),
		},
	})
	attachment4.Pretext = totalSprintWorklogsOfTeam
	attachment4.MarkdownIn = append(attachment3.MarkdownIn, attachment4.Pretext)
	attachments = append(attachments, attachment4)
	return "", attachments, nil
}

//MakeDate convert date to string
func MakeDate(date time.Time) string {
	y := date.Year()
	m := int(date.Month())
	d := date.Day()
	var sdate string
	if m < 10 {
		sdate = fmt.Sprintf("%v.0%v.%v", d, m, y)
	} else {
		sdate = fmt.Sprintf("%v.%v.%v", d, m, y)
	}
	return sdate
}
