package sprint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.com/team-monitoring/comedian/collector"
	"gitlab.com/team-monitoring/comedian/utils"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/model"
)

type Task struct {
	Issue            string
	Status           string
	Assignee         string
	AssigneeFullName string
	Link             string
}

type ActiveSprint struct {
	Name               string
	TotalNumberOfTasks int
	InProgressTasks    []Task
	NotInProgressTasks []string
	ResolvedTasksCount int
	SprintDaysCount    int
	PassedDays         int
	URL                string
	StartDate          time.Time
	LeftDays           int
	SprintProgress     float32
}

//SprintInfo used to parse sprint data from Collector
type SprintInfo struct {
	SprintName  string `json:"sprint_name"`
	SprintURL   string `json:"link_to_sprint"`
	SprintStart string `json:"started"`
	SprintEnd   string `json:"end"`
	Tasks       []struct {
		Issue            string `json:"title"`
		Status           string `json:"status"`
		Assignee         string `json:"assignee"`
		AssigneeFullName string `json:"assignee_name"`
		Link             string `json:"link"`
	} `json:"issues"`
}

//countDays return count days between periods
func countDays(startDate, endDate time.Time) (countDays int) {
	countDays = int(endDate.Sub(startDate).Hours() / 24)
	return countDays
}

//GetSprintInfo sends api request to collector service and returns Info object
func (r *SprintReporter) GetSprintInfo(project string) (sprintInfo SprintInfo, err error) {
	collectorURL := fmt.Sprintf("%v/rest/api/v1/projects/%v/%v/sprint/detail/", r.bot.Conf.CollectorURL, r.bot.TeamDomain, project)
	req, err := http.NewRequest("GET", collectorURL, nil)
	if err != nil {
		return sprintInfo, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", r.bot.Conf.CollectorToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sprintInfo, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err := fmt.Sprintf("could not get collector data. request [%v], responce [%v]", req.URL.Path, res.StatusCode)
		return sprintInfo, errors.New(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return sprintInfo, err
	}
	err = json.Unmarshal(body, &sprintInfo)
	if err != nil {
		return sprintInfo, err
	}
	return sprintInfo, err
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

//MakeActiveSprint make ActiveSprint struct from CollectorInfo
func MakeActiveSprint(sprintInfo SprintInfo) (ActiveSprint, error) {
	var activeSprint ActiveSprint
	activeSprint.Name = sprintInfo.SprintName
	activeSprint.URL = sprintInfo.SprintURL

	startDate, err := prepareTime(sprintInfo.SprintStart)
	if err != nil {
		return activeSprint, err
	}
	endDate, err := prepareTime(sprintInfo.SprintEnd)
	if err != nil {
		return activeSprint, err
	}

	activeSprint.SprintDaysCount = countDays(startDate, endDate)
	//this may recognize current date incorrectly since it does so with timeUTC
	currentDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	activeSprint.PassedDays = countDays(startDate, currentDate)

	activeSprint.StartDate = startDate
	activeSprint.SprintDaysCount = countDays(startDate, endDate)
	activeSprint.PassedDays = countDays(startDate, time.Now())

	activeSprint.TotalNumberOfTasks = len(sprintInfo.Tasks)

	var hasInProgressTasks []string
	for _, task := range sprintInfo.Tasks {
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
	for _, task := range sprintInfo.Tasks {
		if !bot.InList(task.AssigneeFullName, hasInProgressTasks) {
			if !bot.InList(task.AssigneeFullName, activeSprint.NotInProgressTasks) {
				activeSprint.NotInProgressTasks = append(activeSprint.NotInProgressTasks, task.AssigneeFullName)
			}
		}
	}

	sprintProgress := float32(activeSprint.ResolvedTasksCount) / float32(activeSprint.TotalNumberOfTasks) * 100
	leftDays := activeSprint.SprintDaysCount - activeSprint.PassedDays

	activeSprint.SprintProgress = sprintProgress
	activeSprint.LeftDays = leftDays

	return activeSprint, nil
}

//MakeMessage make message about sprint
//Need to divide this into several separate functions!
func (r *SprintReporter) MakeMessage(activeSprint ActiveSprint, worklogs string) (string, []slack.Attachment, error) {
	localizer := i18n.NewLocalizer(r.bot.Bundle, r.bot.CP.Language)
	var attachments []slack.Attachment
	var sprintBasicInfo slack.Attachment

	//if sprint's duration passed send message about sprint overdue on
	if activeSprint.LeftDays < 0 {
		sprintExpiredDays := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "SprintExpiredDays",
				Description: "Displays expired day or days of sprint",
				One:         "day",
				Other:       "days",
			},
			PluralCount: activeSprint.LeftDays,
		})

		sprintExpired := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "SprintExpired",
				Description: "Display message if sprint expired",
				Other:       "Deadline of sprint expired {{.days}} {{.day}} ago.",
			},
			TemplateData: map[string]interface{}{
				"days": activeSprint.LeftDays * (-1),
				"day":  sprintExpiredDays,
			},
		})
		message := sprintExpired
		message += "\n"
		message += activeSprint.URL
		return message, attachments, nil
	}

	sprintProgressPercent := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SprintProgressPercent",
			Description: "Displays header message about sprint progress",
			Other:       "Sprint completed: {{.percent}}%",
		},
		TemplateData: map[string]interface{}{
			"percent": int(activeSprint.SprintProgress),
		},
	})
	sprintBasicInfo.Title = sprintProgressPercent
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
	sprintBasicInfo.Pretext = sprintTitle
	sprintBasicInfo.MarkdownIn = append(sprintBasicInfo.MarkdownIn, sprintBasicInfo.Pretext)

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
		PluralCount: activeSprint.LeftDays,
	})
	sprintDays := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "SprintDays",
			Description: "Displays message about sprint duration",
			Other:       "Sprint lasts {{.totalDays}} {{.td}}, {{.passedDays}} {{.pd}} passed, {{.leftDays}} {{.ld}} left.\n",
		},
		TemplateData: map[string]interface{}{
			"totalDays":  activeSprint.SprintDaysCount,
			"td":         sprintTotalDays,
			"passedDays": activeSprint.PassedDays,
			"pd":         sprintPassedDays,
			"leftDays":   activeSprint.LeftDays,
			"ld":         sprintLeftDays,
		},
	})
	sprintDays += "\n"
	sprintBasicInfo.Text = sprintDays

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
	sprintBasicInfo.Text += urlTitle
	attachments = append(attachments, sprintBasicInfo)

	var inProgressTasts slack.Attachment
	inProgressTasksTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "InProgressTasksTitle",
			Description: "Displays title of message about in progress issues",
			One:         "*In progress task:* ",
			Other:       "*In progress tasks:* ",
		},
		PluralCount: len(activeSprint.InProgressTasks),
	})
	inProgressTasts.Pretext = inProgressTasksTitle
	inProgressTasts.MarkdownIn = append(inProgressTasts.MarkdownIn, inProgressTasts.Pretext)
	attachments = append(attachments, inProgressTasts)

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

	var notInProgressTasks slack.Attachment
	notInProgressTaskTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "notInProgressTaskTitle",
			Description: "Displays title of message about members that has not inprogress tasks",
			Other:       "*No tasks In Progress*",
		},
	})
	notInProgressTasks.Pretext = notInProgressTaskTitle
	notInProgressTasks.MarkdownIn = append(notInProgressTasks.MarkdownIn, notInProgressTasks.Pretext)
	attachments = append(attachments, notInProgressTasks)

	if len(activeSprint.NotInProgressTasks) > 0 {
		var attachment slack.Attachment
		for _, task := range activeSprint.NotInProgressTasks {
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

	var worklogsInfo slack.Attachment
	totalSprintWorklogsOfTeam := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "totalSprintWorklogsOfTeam",
			Description: "Displays message about total worklogs of team members",
			Other:       "*Total worklogs of team on period from {{.start}} to {{.now}}:*  {{.totalWorklogs}} (h)",
		},
		TemplateData: map[string]interface{}{
			"totalWorklogs": worklogs,
			"start":         activeSprint.StartDate.Format("2006-01-02"),
			"now":           time.Now().Format("2006-01-02"),
		},
	})
	worklogsInfo.Pretext = totalSprintWorklogsOfTeam
	worklogsInfo.MarkdownIn = append(worklogsInfo.MarkdownIn, worklogsInfo.Pretext)
	attachments = append(attachments, worklogsInfo)
	return "", attachments, nil
}

//CountTotalWorklogs counts total worklogs
func (r *SprintReporter) CountTotalWorklogs(channelID string, startDate time.Time) (string, error) {
	var totalTeamWorlogs int
	channelMembers, err := r.bot.DB.ListChannelMembers(channelID)
	if err != nil {
		return "", err
	}
	for _, channelMember := range channelMembers {
		_, dataOnUserInProject, err := r.GetCollectorDataOnMember(channelMember, startDate, time.Now())
		if err != nil {
			// need to think if we still want to get sprint info if worklogs are incorrect.
			logrus.Error(err)
			continue
		}
		totalTeamWorlogs += dataOnUserInProject.Worklogs
	}
	return utils.SecondsToHuman(totalTeamWorlogs), nil
}

//GetCollectorDataOnMember sends API request to Collector endpoint and returns CollectorData type
func (r *SprintReporter) GetCollectorDataOnMember(member model.ChannelMember, startDate, endDate time.Time) (collector.Data, collector.Data, error) {
	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	project, err := r.bot.DB.GetChannelName(member.ChannelID)
	if err != nil {
		return collector.Data{}, collector.Data{}, err
	}

	dataOnUser, err := collector.GetCollectorData(r.bot, "users", member.UserID, dateFrom, dateTo)
	if err != nil {
		return collector.Data{}, collector.Data{}, err
	}

	userInProject := fmt.Sprintf("%v/%v", member.UserID, project)
	dataOnUserInProject, err := collector.GetCollectorData(r.bot, "user-in-project", userInProject, dateFrom, dateTo)
	if err != nil {
		return collector.Data{}, collector.Data{}, err
	}

	return dataOnUser, dataOnUserInProject, err
}
