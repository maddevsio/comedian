package botuser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
)

//CollectorData used to parse data on user from Collector
type CollectorData struct {
	Commits  int `json:"total_commits"`
	Worklogs int `json:"worklogs"`
}

//AttachmentItem is needed to sort attachments
type AttachmentItem struct {
	SlackAttachment slack.Attachment
	Points          int
}

// CallDisplayYesterdayTeamReport calls displayYesterdayTeamReport
func (bot *Bot) CallDisplayYesterdayTeamReport() error {
	if bot.properties.ReportingTime == "" {
		return nil
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, err := w.Parse(bot.properties.ReportingTime, time.Now())
	if err != nil {
		return err
	}

	if time.Now().Hour() == r.Time.Hour() && time.Now().Minute() == r.Time.Minute() {
		_, err := bot.displayYesterdayTeamReport()
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// CallDisplayWeeklyTeamReport calls displayWeeklyTeamReport
func (bot *Bot) CallDisplayWeeklyTeamReport() error {
	if int(time.Now().Weekday()) != 0 {
		return nil
	}

	if bot.properties.ReportingTime == "" {
		return nil
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	r, err := w.Parse(bot.properties.ReportingTime, time.Now())

	if time.Now().Hour() == r.Time.Hour() && time.Now().Minute() == r.Time.Minute() {
		_, err = bot.displayWeeklyTeamReport()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// displayYesterdayTeamReport generates report on users who submit standups
func (bot *Bot) displayYesterdayTeamReport() (FinalReport string, err error) {
	var allReports []slack.Attachment

	channels, err := bot.db.ListChannels()
	if err != nil {
		log.Errorf("GetAllChannels failed: %v", err)
		return FinalReport, err
	}

	reportHeader, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "reportHeader",
			Other: "",
		},
	})
	if err != nil {
		log.Error(err)
	}

	for _, channel := range channels {
		if channel.TeamID != bot.properties.TeamID {
			continue
		}
		var attachments []slack.Attachment
		var attachmentsPull []AttachmentItem

		Standupers, err := bot.db.ListChannelStandupers(channel.ChannelID)
		if err != nil {
			log.Errorf("ListChannelStandupers failed for channel %v: %v", channel.ChannelName, err)
			continue
		}

		if len(Standupers) == 0 {
			continue
		}

		for _, member := range Standupers {
			var attachment slack.Attachment
			var attachmentFields []slack.AttachmentField
			var worklogs, commits, standup string
			var worklogsPoints, commitsPoints, standupPoints int

			userInfo, err := bot.slack.GetUserInfo(member.UserID)
			if err != nil {
				continue
			}

			dataOnUser, dataOnUserInProject, collectorError := bot.GetCollectorDataOnMember(member, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))

			if collectorError == nil {
				worklogs, worklogsPoints = bot.processWorklogs(dataOnUser.Worklogs, dataOnUserInProject.Worklogs)
				commits, commitsPoints = bot.processCommits(dataOnUser.Commits, dataOnUserInProject.Commits)
			}

			if member.RoleInChannel == "pm" || member.RoleInChannel == "designer" {
				commits = ""
				commitsPoints++
			}

			if collectorError != nil {
				worklogs = ""
				worklogsPoints++
				commits = ""
				commitsPoints++
			}

			standup, standupPoints = bot.processStandup(member)

			fieldValue := worklogs + commits + standup

			//if there is nothing to show, do not create attachment
			if fieldValue == "" {
				log.Warningf("Nothing to show... skip member! %v", member)
				continue
			}

			attachmentFields = append(attachmentFields, slack.AttachmentField{
				Value: fieldValue,
				Short: false,
			})

			points := worklogsPoints + commitsPoints + standupPoints

			//attachment text will be depend on worklogsPoints,commitsPoints and standupPoints
			if points >= 3 {
				notTagStanduper, err := bot.localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "notTagStanduper",
						Other: "",
					},
					TemplateData: map[string]interface{}{"user": userInfo.RealName, "channel": channel.ChannelName},
				})
				if err != nil {
					log.Error(err)
				}
				attachment.Text = notTagStanduper
			} else {
				tagStanduper, err := bot.localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "tagStanduper",
						Other: "",
					},
					TemplateData: map[string]interface{}{"user": member.UserID, "channel": channel.ChannelName},
				})
				if err != nil {
					log.Error(err)
				}
				attachment.Text = tagStanduper
			}

			switch points {
			case 0:
				attachment.Color = "danger"
			case 1, 2:
				attachment.Color = "warning"
			case 3:
				attachment.Color = "good"
			}

			if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
				attachment.Color = "good"
			}

			attachment.Fields = attachmentFields

			item := AttachmentItem{
				SlackAttachment: attachment,
				Points:          dataOnUserInProject.Worklogs,
			}

			attachmentsPull = append(attachmentsPull, item)
		}

		if len(attachmentsPull) == 0 {
			continue
		}

		attachments = bot.sortReportEntries(attachmentsPull)
		if bot.properties.IndividualReportsOn {
			bot.messageChan <- Message{
				Type:        "message",
				Channel:     channel.ChannelID,
				Text:        reportHeader,
				Attachments: attachments,
			}
		}

		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return
	}

	reportingChannelID := ""
	for _, ch := range channels {
		if (ch.ChannelName == bot.properties.ReportingChannel && ch.TeamID == bot.properties.TeamID) || (ch.ChannelID == bot.properties.ReportingChannel && ch.TeamID == bot.properties.TeamID) {
			reportingChannelID = ch.ChannelID
		}
	}

	bot.messageChan <- Message{
		Type:        "message",
		Channel:     reportingChannelID,
		Text:        reportHeader,
		Attachments: allReports,
	}
	FinalReport = fmt.Sprintf(reportHeader, allReports)
	return FinalReport, nil
}

// displayWeeklyTeamReport generates report on users who submit standups
func (bot *Bot) displayWeeklyTeamReport() (string, error) {
	var FinalReport string
	var allReports []slack.Attachment

	channels, err := bot.db.ListChannels()
	if err != nil {
		return FinalReport, err
	}

	reportHeaderWeekly, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "reportHeaderWeekly",
			Other: "",
		},
	})
	if err != nil {
		log.Error(err)
	}

	for _, channel := range channels {
		var attachmentsPull []AttachmentItem
		var attachments []slack.Attachment

		Standupers, err := bot.db.ListChannelStandupers(channel.ChannelID)
		if err != nil {
			log.Errorf("ListChannelStandupers failed for channel %v: %v", channel.ChannelName, err)
			continue
		}

		if len(Standupers) == 0 {
			continue
		}

		for _, member := range Standupers {
			var attachment slack.Attachment
			var attachmentFields []slack.AttachmentField
			var worklogs, commits string
			var worklogsPoints, commitsPoints int

			userInfo, err := bot.slack.GetUserInfo(member.UserID)
			if err != nil {
				continue
			}

			dataOnUser, dataOnUserInProject, collectorError := bot.GetCollectorDataOnMember(member, time.Now().AddDate(0, 0, -7), time.Now().AddDate(0, 0, -1))

			if collectorError == nil {
				worklogs, worklogsPoints = bot.processWeeklyWorklogs(dataOnUser.Worklogs, dataOnUserInProject.Worklogs)
				commits, commitsPoints = bot.processCommits(dataOnUser.Commits, dataOnUserInProject.Commits)
			}

			if member.RoleInChannel == "pm" || member.RoleInChannel == "designer" {
				commits = ""
				commitsPoints++
			}

			if collectorError != nil {
				worklogs = ""
				worklogsPoints++
				commits = ""
				commitsPoints++
			}

			fieldValue := worklogs + commits

			//if there is nothing to show, do not create attachment
			if fieldValue == "" {
				continue
			}

			attachmentFields = append(attachmentFields, slack.AttachmentField{
				Value: fieldValue,
				Short: false,
			})

			points := worklogsPoints + commitsPoints

			if points >= 2 {
				notTagStanduper, err := bot.localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "notTagStanduper",
						Other: "",
					},
					TemplateData: map[string]interface{}{"user": userInfo.RealName, "channel": channel.ChannelName},
				})
				if err != nil {
					log.Error(err)
				}
				attachment.Text = notTagStanduper
			} else {
				tagStanduper, err := bot.localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "tagStanduper",
						Other: "",
					},
					TemplateData: map[string]interface{}{"user": member.UserID, "channel": channel.ChannelName},
				})
				if err != nil {
					log.Error(err)
				}
				attachment.Text = tagStanduper
			}

			switch points {
			case 0:
				attachment.Color = "danger"
			case 1:
				attachment.Color = "warning"
			case 2:
				attachment.Color = "good"
			}

			attachment.Fields = attachmentFields

			item := AttachmentItem{
				SlackAttachment: attachment,
				Points:          dataOnUserInProject.Worklogs,
			}

			attachmentsPull = append(attachmentsPull, item)
		}

		if len(attachmentsPull) == 0 {
			continue
		}

		attachments = bot.sortReportEntries(attachmentsPull)

		if bot.properties.IndividualReportsOn {
			bot.messageChan <- Message{
				Type:        "message",
				Channel:     channel.ChannelID,
				Text:        reportHeaderWeekly,
				Attachments: attachments,
			}
		}
		allReports = append(allReports, attachments...)
	}

	if len(allReports) == 0 {
		return FinalReport, nil
	}

	reportingChannelID := ""
	for _, ch := range channels {
		if (ch.ChannelName == bot.properties.ReportingChannel && ch.TeamID == bot.properties.TeamID) || (ch.ChannelID == bot.properties.ReportingChannel && ch.TeamID == bot.properties.TeamID) {
			reportingChannelID = ch.ChannelID
		}
	}

	bot.messageChan <- Message{
		Type:        "message",
		Channel:     reportingChannelID,
		Text:        reportHeaderWeekly,
		Attachments: allReports,
	}
	FinalReport = fmt.Sprintf(reportHeaderWeekly, allReports)
	return FinalReport, nil
}

func (bot *Bot) processWorklogs(totalWorklogs, projectWorklogs int) (string, int) {

	var points int
	worklogsEmoji := ""

	w := totalWorklogs / 3600
	switch {
	case w < 3:
		worklogsEmoji = ":angry:"
	case w >= 3 && w < 7:
		worklogsEmoji = ":disappointed:"
	case w >= 7 && w < 9:
		worklogsEmoji = ":wink:"
		points++
	case w >= 9:
		worklogsEmoji = ":sunglasses:"
		points++
	}

	worklogsTime := SecondsToHuman(totalWorklogs)

	if totalWorklogs != projectWorklogs {
		var err error
		worklogsTime, err = bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "worklogsTime",
				Other: "",
			},
			TemplateData: map[string]interface{}{"projectWorklogs": SecondsToHuman(projectWorklogs), "totalWorklogs": SecondsToHuman(totalWorklogs)},
		})
		if err != nil {
			log.Error(err)
		}
	}

	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		worklogsEmoji = ""
		if projectWorklogs == 0 {
			return "", points
		}
	}

	worklogsTranslation, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "worklogsTranslation",
			Other: "",
		},
		TemplateData: map[string]interface{}{"worklogsTime": worklogsTime, "worklogsEmoji": worklogsEmoji},
	})
	if err != nil {
		log.Error(err)
	}
	return worklogsTranslation, points
}

func (bot *Bot) processWeeklyWorklogs(totalWorklogs, projectWorklogs int) (string, int) {
	var points int
	worklogsEmoji := ""

	w := totalWorklogs / 3600
	switch {
	case w < 31:
		worklogsEmoji = ":disappointed:"
	case w >= 31 && w < 35:
		worklogsEmoji = ":wink:"
		points++
	case w >= 35:
		worklogsEmoji = ":sunglasses:"
		points++
	}
	worklogsTime := SecondsToHuman(totalWorklogs)

	if totalWorklogs != projectWorklogs {
		var err error
		worklogsTime, err = bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "worklogsTime",
				Other: "",
			},
			TemplateData: map[string]interface{}{"projectWorklogs": SecondsToHuman(projectWorklogs), "totalWorklogs": SecondsToHuman(totalWorklogs)},
		})
		if err != nil {
			log.Error(err)
		}
	}

	worklogsTranslation, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "worklogsTranslation",
			Other: "",
		},
		TemplateData: map[string]interface{}{"worklogsTime": worklogsTime, "worklogsEmoji": worklogsEmoji},
	})
	if err != nil {
		log.Error(err)
	}

	return worklogsTranslation, points
}

func (bot *Bot) processCommits(totalCommits, projectCommits int) (string, int) {
	var points int
	commitsEmoji := ""

	c := projectCommits
	switch {
	case c == 0:
		commitsEmoji = ":shit:"
	case c > 0:
		commitsEmoji = ":wink:"
		points++
	}

	if int(time.Now().Weekday()) == 0 || int(time.Now().Weekday()) == 1 {
		commitsEmoji = ""
		if projectCommits == 0 {
			return "", points
		}
	}

	commitsTranslation, err := bot.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "commitsTranslation",
			Other: "",
		},
		TemplateData: map[string]interface{}{"projectCommits": projectCommits, "commitsEmoji": commitsEmoji},
	})
	if err != nil {
		log.Error(err)
	}
	return commitsTranslation, points
}

func (bot *Bot) processStandup(member model.Standuper) (string, int) {
	var text string
	var points int

	t := time.Now().AddDate(0, 0, -1)

	timeFrom := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	timeTo := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)

	standup, err := bot.db.GetStandupForPeriod(member.UserID, member.ChannelID, timeFrom, timeTo)
	if err != nil || standup == nil {
		if time.Now().Weekday() == 0 || time.Now().Weekday() == 1 {
			return "", points
		}
		if member.RoleInChannel == "pm" {
			return "", 1
		}
		noStandup, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noStandup",
				Other: "",
			},
		})
		if err != nil {
			log.Error(err)
		}
		text = noStandup
	} else {
		hasStandup, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "hasStandup",
				Other: "",
			},
		})
		if err != nil {
			log.Error(err)
		}
		text = hasStandup
		points++
	}

	return text, points
}

func (bot *Bot) sortReportEntries(entries []AttachmentItem) []slack.Attachment {
	var attachments []slack.Attachment

	for i := 0; i < len(entries); i++ {
		if !sweep(entries, i) {
			break
		}
	}

	for _, item := range entries {
		attachments = append(attachments, item.SlackAttachment)
	}

	return attachments
}

func sweep(entries []AttachmentItem, prevPasses int) bool {
	var N = len(entries)
	var didSwap = false
	var firstIndex = 0
	var secondIndex = 1

	for secondIndex < (N - prevPasses) {

		var firstItem = entries[firstIndex]
		var secondItem = entries[secondIndex]
		if entries[firstIndex].Points < entries[secondIndex].Points {
			entries[firstIndex] = secondItem
			entries[secondIndex] = firstItem
			didSwap = true
		}
		firstIndex++
		secondIndex++
	}

	return didSwap
}

//GetCollectorDataOnMember sends API request to Collector endpoint and returns CollectorData type
func (bot *Bot) GetCollectorDataOnMember(member model.Standuper, startDate, endDate time.Time) (CollectorData, CollectorData, error) {
	dateFrom := fmt.Sprintf("%d-%02d-%02d", startDate.Year(), startDate.Month(), startDate.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", endDate.Year(), endDate.Month(), endDate.Day())

	project, err := bot.db.SelectChannel(member.ChannelID)
	if err != nil {
		return CollectorData{}, CollectorData{}, err
	}

	dataOnUser, err := bot.GetCollectorData("users", member.UserID, dateFrom, dateTo)
	if err != nil {
		return CollectorData{}, CollectorData{}, err
	}

	userInProject := fmt.Sprintf("%v/%v", member.UserID, project.ChannelName)
	dataOnUserInProject, err := bot.GetCollectorData("user-in-project", userInProject, dateFrom, dateTo)
	if err != nil {
		return CollectorData{}, CollectorData{}, err
	}

	return dataOnUser, dataOnUserInProject, err
}

//GetCollectorData sends api request to collector servise and returns collector object
func (bot *Bot) GetCollectorData(getDataOn, data, dateFrom, dateTo string) (CollectorData, error) {
	var collectorData CollectorData
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", bot.conf.CollectorURL, bot.properties.TeamID, getDataOn, data, dateFrom, dateTo)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		return collectorData, err
	}
	token := bot.conf.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return collectorData, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		log.WithFields(log.Fields(map[string]interface{}{"body": string(body), "requestURL": linkURL, "res.StatusCode": res.StatusCode})).Warning("Failed to get collector data on member!")
		return collectorData, fmt.Errorf("failed to get collector data. %v", res.StatusCode)
	}
	json.Unmarshal(body, &collectorData)
	return collectorData, nil
}

//SecondsToHuman converts seconds (int) to HH:MM format
func SecondsToHuman(input int) string {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	return fmt.Sprintf("%v:%02d", int(hours), int(minutes))
}
