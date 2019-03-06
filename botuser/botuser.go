package botuser

import (
	"errors"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"gitlab.com/team-monitoring/comedian/translation"
	"gitlab.com/team-monitoring/comedian/utils"
)

var (
	typeMessage       = ""
	typeEditMessage   = "message_changed"
	typeDeleteMessage = "message_deleted"
)

const (
	adminAccess       = 2
	pmAccess          = 3
	regularUserAccess = 4
)

// Bot struct used for storing and communicating with slack api
type Bot struct {
	slack      *slack.Client
	rtm        *slack.RTM
	properties model.BotSettings
	db         *storage.MySQL
	bundle     *i18n.Bundle
}

func New(bundle *i18n.Bundle, settings model.BotSettings, db *storage.MySQL) *Bot {
	bot := &Bot{}
	bot.slack = slack.New(settings.AccessToken)
	bot.rtm = bot.slack.NewRTM()
	bot.properties = settings
	bot.db = db

	bot.bundle = bundle

	return bot
}

func (bot *Bot) Start() {
	bot.UpdateUsersList()

	go func(bot *Bot) {

		go bot.rtm.ManageConnection()
		for msg := range bot.rtm.IncomingEvents {
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				bot.HandleMessage(ev)
			case *slack.ConnectedEvent:
				payload := translation.Payload{bot.bundle, bot.properties.Language, "Reconnected", 0, nil}
				message, err := translation.Translate(payload)
				if err != nil {
					log.WithFields(log.Fields{
						"TeamName":     bot.properties.TeamName,
						"Language":     payload.Lang,
						"MessageID":    payload.MessageID,
						"Count":        payload.Count,
						"TemplateData": payload.TemplateData,
					}).Error("Failed to translate message!")
					continue
				}
				log.Info(message)
				bot.properties.UserID = bot.rtm.GetInfo().User.ID
			case *slack.MemberJoinedChannelEvent:
				bot.HandleJoin(ev.Channel, ev.Team)
			}
		}
	}(bot)

	go func(bot *Bot) {
		counter := time.NewTicker(time.Second * 60).C
		for time := range counter {
			bot.NotifyChannels(time)
		}
	}(bot)
}

func (bot *Bot) HandleMessage(msg *slack.MessageEvent) {

	switch msg.SubType {
	case typeMessage:
		err := bot.HandleNewMessage(msg)
		if err != nil {
			log.Error(err)
		}

	case typeEditMessage:
		err := bot.HandleEditMessage(msg)
		if err != nil {
			log.Error(err)
		}

	case typeDeleteMessage:
		err := bot.HandleDeleteMessage(msg)
		if err != nil {
			log.Error(err)
		}
	}
}

func (bot *Bot) HandleNewMessage(msg *slack.MessageEvent) error {
	if !strings.Contains(msg.Msg.Text, bot.properties.UserID) && !strings.Contains(msg.Msg.Text, "#standup") {
		return errors.New("bot is not mentioned and no #standup in the message body")
	}

	problem := bot.analizeStandup(msg.Msg.Text)
	if problem != "" {
		bot.SendEphemeralMessage(msg.Channel, msg.User, problem)
		return errors.New("Fail to save message as standup. Standup is not complete")
	}

	if bot.db.SubmittedStandupToday(msg.User, msg.Channel) {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "OneStandupPerDay", 0, nil}
		oneStandupPerDay, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		bot.SendEphemeralMessage(msg.Channel, msg.User, oneStandupPerDay)
		return errors.New("Fail to save message as standup. User already submitted standup today")
	}
	standup, err := bot.db.CreateStandup(model.Standup{
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.User,
		Comment:   msg.Msg.Text,
		MessageTS: msg.Msg.Timestamp,
	})
	if err != nil {
		return err
	}

	log.Infof("Standup created #id:%v\n", standup.ID)
	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.Msg.Timestamp,
		File:      "",
		Comment:   "",
	}
	bot.slack.AddReaction("heavy_check_mark", item)
	return nil
}

func (bot *Bot) HandleEditMessage(msg *slack.MessageEvent) error {
	if !strings.Contains(msg.SubMessage.Text, bot.properties.UserID) && !strings.Contains(msg.SubMessage.Text, "#standup") {
		return errors.New("bot is not mentioned and no #standup in the message body")
	}

	standup, err := bot.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
	if err != nil {
		problem := bot.analizeStandup(msg.SubMessage.Text)
		if problem != "" {
			bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
			return errors.New("Fail to save edited message as standup. Standup is not complete")
		}

		if bot.db.SubmittedStandupToday(msg.SubMessage.User, msg.Channel) {
			payload := translation.Payload{bot.bundle, bot.properties.Language, "OneStandupPerDay", 0, nil}
			oneStandupPerDay, err := translation.Translate(payload)
			if err != nil {
				log.WithFields(log.Fields{
					"TeamName":     bot.properties.TeamName,
					"Language":     payload.Lang,
					"MessageID":    payload.MessageID,
					"Count":        payload.Count,
					"TemplateData": payload.TemplateData,
				}).Error("Failed to translate message!")
			}
			bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, oneStandupPerDay)
			return errors.New("Fail to save edited message as standup. User already submitted standup today")
		}

		standup, err := bot.db.CreateStandup(model.Standup{
			TeamID:    msg.Team,
			ChannelID: msg.Channel,
			UserID:    msg.SubMessage.User,
			Comment:   msg.SubMessage.Text,
			MessageTS: msg.SubMessage.Timestamp,
		})
		if err != nil {
			return err
		}

		log.Infof("Standup created #id:%v\n", standup.ID)

		item := slack.ItemRef{
			Channel:   msg.Channel,
			Timestamp: msg.SubMessage.Timestamp,
			File:      "",
			Comment:   "",
		}
		bot.slack.AddReaction("heavy_check_mark", item)
		return nil
	}

	problem := bot.analizeStandup(msg.SubMessage.Text)
	if problem != "" {
		bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
		return errors.New("Fail to save edited message as standup. Standup is not complete")
	}

	standup.Comment = msg.SubMessage.Text
	st, err := bot.db.UpdateStandup(standup)
	if err != nil {
		return err
	}
	log.Infof("Standup updated #id:%v\n", st.ID)
	return nil
}

func (bot *Bot) HandleDeleteMessage(msg *slack.MessageEvent) error {
	standup, err := bot.db.SelectStandupByMessageTS(msg.DeletedTimestamp)
	if err != nil {
		return err
	}
	err = bot.db.DeleteStandup(standup.ID)
	if err != nil {
		return err
	}
	log.Infof("Standup deleted #id:%v\n", standup.ID)
	return nil
}

func (bot *Bot) analizeStandup(message string) string {
	message = strings.ToLower(message)

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{"yesterday", "friday", "monday", "tuesday", "wednesday", "thursday", "saturday", "sunday", "completed", "вчера", "пятниц", "делал", "сделано", "понедельник", "вторник", "сред", "четверг", "суббот", "воскресенье"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	if !mentionsYesterdayWork {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "StandupHandleNoYesterdayWorkMentioned", 0, nil}
		problem, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")

		}
		return problem
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{"today", "going", "plan", "сегодня", "собираюсь", "план"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}
	if !mentionsTodayPlans {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "StandupHandleNoTodayPlansMentioned", 0, nil}
		problem, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")

		}
		return problem
	}

	mentionsProblem := false

	problemKeys := []string{"problem", "difficult", "stuck", "question", "issue", "block", "проблем", "трудност", "затруднени", "вопрос", "блок"}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}
	if !mentionsProblem {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "StandupHandleNoProblemsMentioned", 0, nil}
		problem, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")

		}
		return problem
	}

	return ""
}

// SendMessage posts a message in a specified channel visible for everyone
func (bot *Bot) SendMessage(channel, message string, attachments []slack.Attachment) error {
	_, _, err := bot.slack.PostMessage(channel, message, slack.PostMessageParameters{
		Attachments: attachments,
	})
	if err != nil {
		log.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

// SendEphemeralMessage posts a message in a specified channel which is visible only for selected user
func (bot *Bot) SendEphemeralMessage(channel, user, message string) error {
	_, err := bot.slack.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(message, true),
	)
	if err != nil {
		log.Errorf("slack: PostEphemeral failed: %v\n", err)
		return err
	}
	return err
}

// SendUserMessage Direct Message specific user
func (bot *Bot) SendUserMessage(userID, message string) error {
	_, _, channelID, err := bot.slack.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	err = bot.SendMessage(channelID, message, nil)
	if err != nil {
		return err
	}
	return err
}

//HandleJoin handles comedian joining channel
func (bot *Bot) HandleJoin(channelID, teamID string) {
	_, err := bot.db.SelectChannel(channelID)
	if err == nil {
		return
	}

	channel, err := bot.slack.GetConversationInfo(channelID, true)
	if err != nil {
		log.Errorf("GetConversationInfo failed: %v", err)
		return
	}
	createdChannel, err := bot.db.CreateChannel(model.Channel{
		TeamID:      teamID,
		ChannelName: channel.Name,
		ChannelID:   channel.ID,
		StandupTime: int64(0),
	})
	if err != nil {
		log.Errorf("CreateChannel failed: %v", err)
		return
	}
	log.Infof("New Channel Created: %v", createdChannel)
}

func (bot *Bot) ImplementCommands(form model.FullSlackForm) string {

	accessLevel, err := bot.getAccessLevel(form.UserID, form.ChannelID)
	if err != nil {
		return err.Error()
	}

	command, params := utils.CommandParsing(form.Text)

	switch command {
	case "add":
		return bot.addCommand(accessLevel, form.ChannelID, params)
	case "show":
		return bot.showCommand(form.ChannelID, params)
	case "remove":
		return bot.deleteCommand(accessLevel, form.ChannelID, params)
	case "add_deadline":
		return bot.addTime(accessLevel, form.ChannelID, params)
	case "remove_deadline":
		return bot.removeTime(accessLevel, form.ChannelID)
	case "show_deadline":
		return bot.showTime(form.ChannelID)
	default:
		return bot.DisplayHelpText("")
	}
}

func (bot *Bot) getAccessLevel(userID, channelID string) (int, error) {
	user, err := bot.db.SelectUser(userID)
	if err != nil {
		return 0, err
	}
	if user.IsAdmin() {
		return 2, nil
	}
	if bot.db.UserIsPMForProject(userID, channelID) {
		return 3, nil
	}
	return 4, nil
}

//UpdateUsersList updates users in workspace
func (bot *Bot) UpdateUsersList() {
	users, err := bot.slack.GetUsers()
	if err != nil {
		log.Errorf("GetUsers failed: %v", err)
		return
	}
	for _, user := range users {
		if user.IsBot || user.Name == "slackbot" {
			continue
		}

		u, err := bot.db.SelectUser(user.ID)
		if err != nil && !user.Deleted {
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				u, err = bot.db.CreateUser(model.User{
					TeamID:   user.TeamID,
					UserName: user.Name,
					UserID:   user.ID,
					Role:     "admin",
					RealName: user.RealName,
				})
				if err != nil {
					log.Errorf("CreateUser failed %v", err)
					continue
				}
				continue
			}
			u, err = bot.db.CreateUser(model.User{
				TeamID:   user.TeamID,
				UserName: user.Name,
				UserID:   user.ID,
				Role:     "",
				RealName: user.RealName,
			})
			if err != nil {
				log.Errorf("CreateUser with no role failed %v", err)
				continue
			}
		}
		if !user.Deleted {
			u.UserName = user.Name
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				u.Role = "admin"
			}
			u.RealName = user.RealName
			u.TeamID = user.TeamID
			_, err = bot.db.UpdateUser(u)
			if err != nil {
				log.Errorf("Update User failed %v", err)
				continue
			}
		}

		if user.Deleted {
			bot.db.DeleteUser(u.ID)
			cm, err := bot.db.FindMembersByUserID(u.UserID)
			if err != nil {
				continue
			}
			for _, member := range cm {
				bot.db.DeleteStanduper(member.ID)
			}
		}
	}
	log.Info("Users list updated successfully")
}

func (bot *Bot) Suits(team string) bool {
	if team == bot.properties.TeamID || team == bot.properties.TeamName {
		return true
	}
	return false
}

func (bot *Bot) SetProperties(settings model.BotSettings) model.BotSettings {
	bot.properties = settings
	return bot.properties
}
