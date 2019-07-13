package botuser

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

var (
	typeMessage       = ""
	typeEditMessage   = "message_changed"
	typeDeleteMessage = "message_deleted"
)

var problemKeys = []string{"issue", "мешает"}
var todayPlansKeys = []string{"today", "сегодня"}
var yesterdayWorkKeys = []string{"yesterday", "friday", "вчера", "пятниц"}

// Bot struct used for storing and communicating with slack api
type Bot struct {
	slack           *slack.Client
	properties      model.BotSettings
	db              *storage.DB
	localizer       *i18n.Localizer
	wg              sync.WaitGroup
	QuitChan        chan struct{}
	notifierThreads []*NotifierThread
	conf            *config.Config
}

//New creates new Bot instance
func New(config *config.Config, bundle *i18n.Bundle, settings model.BotSettings, db *storage.DB) *Bot {
	bot := &Bot{}
	bot.slack = slack.New(settings.AccessToken)
	bot.properties = settings
	bot.db = db
	bot.localizer = i18n.NewLocalizer(bundle, settings.Language)
	bot.QuitChan = make(chan struct{})
	bot.conf = config
	return bot
}

//Start updates Users list and launches notifications
func (bot *Bot) Start() {
	log.Info("Bot started for ", bot.properties.TeamName)

	err := bot.UpdateUsersList()
	if err != nil {
		log.Errorf("UpdateUsersList failed: %v", err)
	}
	bot.wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second * 60).C
		for {
			select {
			case <-ticker:
				bot.NotifyChannels(time.Now())
				err = bot.CallDisplayYesterdayTeamReport()
				if err != nil {
					log.Error("CallDisplayYesterdayTeamReport failed: ", err)
				}
				err = bot.CallDisplayWeeklyTeamReport()
				if err != nil {
					log.Error("CallDisplayWeeklyTeamReport failed: ", err)
				}
				err = bot.remindAboutWorklogs()
				if err != nil {
					log.Error("remindAboutWorklogs failed: ", err)
				}
			case <-bot.QuitChan:
				bot.wg.Done()
				return
			}
		}
	}()
}

//Stop closes bot QuitChan making bot goroutine to exit
func (bot *Bot) Stop() {
	close(bot.QuitChan)
}

//HandleCallBackEvent handles different callback events from Slack Event Subscription list
func (bot *Bot) HandleCallBackEvent(eventType string, eventData []byte) error {
	switch eventType {
	case "message":
		message := &slack.MessageEvent{}
		if err := json.Unmarshal(eventData, message); err != nil {
			return err
		}
		return bot.HandleMessage(message)
	case "app_mention":
		message := &slack.MessageEvent{}
		if err := json.Unmarshal(eventData, message); err != nil {
			return err
		}
		return bot.HandleAppMention(message)
	case "member_joined_channel":
		join := &slack.MemberJoinedChannelEvent{}
		if err := json.Unmarshal(eventData, join); err != nil {
			return err
		}
		_, err := bot.HandleJoin(join.Channel, join.Team)
		return err
	case "team_join":
		join := &slack.TeamJoinEvent{}
		if err := json.Unmarshal(eventData, join); err != nil {
			return err
		}
		_, err := bot.HandleJoinNewUser(join.User)
		return err
	case "user_change":
		event := &slack.UserChangeEvent{}
		if err := json.Unmarshal(eventData, event); err != nil {
			return err
		}
		return bot.updateUser(event.User)
	case "app_uninstalled":
		bot.Stop()
		err := bot.db.DeleteBotSettings(bot.properties.TeamID)
		if err != nil {
			return err
		}
		return nil
	default:
		log.WithFields(log.Fields{"event": string(eventData)}).Warning("unrecognized event!")
		return nil
	}
}

//HandleMessage handles slack message event
func (bot *Bot) HandleMessage(msg *slack.MessageEvent) error {
	msg.Team = bot.properties.TeamID
	switch msg.SubType {
	case typeMessage:
		return bot.handleNewMessage(msg)
	case typeEditMessage:
		return bot.handleEditMessage(msg)
	case typeDeleteMessage:
		return bot.handleDeleteMessage(msg)
	case "bot_message":
		return nil
	}
	return nil
}

func (bot *Bot) handleNewMessage(msg *slack.MessageEvent) error {
	if !strings.Contains(msg.Msg.Text, bot.properties.UserID) {
		return nil
	}

	problem := bot.analizeStandup(msg.Msg.Text)
	if problem != "" {
		return bot.SendEphemeralMessage(msg.Channel, msg.User, problem)
	}

	_, err := bot.db.CreateStandup(model.Standup{
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.User,
		Comment:   msg.Msg.Text,
		MessageTS: msg.Msg.Timestamp,
	})
	if err != nil {
		bot.SendEphemeralMessage(msg.Channel, msg.User, "I could not save your standup due to some technical issues, please, report ths to your PM or directly to Comedian development team")
		return err
	}
	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.Msg.Timestamp,
		File:      "",
		Comment:   "",
	}
	err = bot.slack.AddReaction("heavy_check_mark", item)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName": bot.properties.TeamName,
			"Item":     item,
		}).Error("Failed to AddReaction!")
	}

	return nil
}

func (bot *Bot) handleEditMessage(msg *slack.MessageEvent) error {
	if !strings.Contains(msg.SubMessage.Text, bot.properties.UserID) {
		return nil
	}

	problem := bot.analizeStandup(msg.SubMessage.Text)
	if problem != "" {
		return bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
	}

	standup, err := bot.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
	if err == nil {
		standup.Comment = msg.SubMessage.Text
		_, err := bot.db.UpdateStandup(standup)
		if err != nil {
			return err
		}
		return nil
	}

	standup, err = bot.db.CreateStandup(model.Standup{
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.SubMessage.User,
		Comment:   msg.SubMessage.Text,
		MessageTS: msg.SubMessage.Timestamp,
	})
	if err != nil {
		return err
	}

	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.SubMessage.Timestamp,
		File:      "",
		Comment:   "",
	}
	err = bot.slack.AddReaction("heavy_check_mark", item)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName": bot.properties.TeamName,
			"Item":     item,
		}).Error("Failed to AddReaction!")
	}

	return nil
}

func (bot *Bot) handleDeleteMessage(msg *slack.MessageEvent) error {
	standup, err := bot.db.SelectStandupByMessageTS(msg.DeletedTimestamp)
	if err != nil {
		return nil
	}
	return bot.db.DeleteStandup(standup.ID)
}

func (bot *Bot) submittedStandupToday(userID, channelID string) bool {
	standup, err := bot.db.SelectLatestStandupByUser(userID, channelID)
	if err != nil {
		return false
	}

	user, err := bot.db.SelectUser(userID)
	if err != nil {
		return false
	}

	loc := time.FixedZone(user.TZ, user.TZOffset)

	if standup.Created.In(loc).Day() == time.Now().UTC().In(loc).Day() {
		return true
	}
	return false
}

//HandleAppMention функция которая работает точно так же как и HandleMessage
//но при этом не говорит пользователю что он уже написал стендап.
func (bot *Bot) HandleAppMention(msg *slack.MessageEvent) error {
	msg.Team = bot.properties.TeamID
	if !strings.Contains(msg.Msg.Text, bot.properties.UserID) {
		return nil
	}

	problem := bot.analizeStandup(msg.Msg.Text)
	if problem != "" {
		return nil
	}

	_, err := bot.db.CreateStandup(model.Standup{
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.User,
		Comment:   msg.Msg.Text,
		MessageTS: msg.Msg.Timestamp,
	})
	if err != nil {
		return err
	}

	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.Msg.Timestamp,
		File:      "",
		Comment:   "",
	}
	err = bot.slack.AddReaction("heavy_check_mark", item)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName": bot.properties.TeamName,
			"Item":     item,
		}).Error("Failed to AddReaction!")
	}

	return nil
}

func (bot *Bot) analizeStandup(message string) string {
	errors := []string{}
	message = strings.ToLower(message)

	var mentionsYesterdayWork, mentionsTodayPlans, mentionsProblem bool

	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}

	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}

	if !mentionsYesterdayWork {
		warnings, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noYesterdayMention",
				Other: "- no 'yesterday' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(yesterdayWorkKeys, ", "),
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}
	if !mentionsTodayPlans {
		warnings, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noTodayMention",
				Other: "- no 'today' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(todayPlansKeys, ", "),
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}
	if !mentionsProblem {
		warnings, err := bot.localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "noProblemsMention",
				Other: "- no 'problems' keywords detected: {{.Keywords}}",
			},
			TemplateData: map[string]interface{}{
				"Keywords": strings.Join(problemKeys, ", "),
			},
		})
		if err != nil {
			log.Error(err)
		}
		errors = append(errors, warnings)
	}
	return strings.Join(errors, ", ")
}

// SendMessage posts a message in a specified channel visible for everyone
func (bot *Bot) SendMessage(channel, message string, attachments []slack.Attachment) error {
	_, _, err := bot.slack.PostMessage(channel, message, slack.PostMessageParameters{Attachments: attachments})
	return err
}

// SendEphemeralMessage posts a message in a specified channel which is visible only for selected user
func (bot *Bot) SendEphemeralMessage(channel, user, message string) error {
	_, err := bot.slack.PostEphemeral(channel, user, slack.MsgOptionText(message, true))
	return err
}

// SendUserMessage Direct Message specific user
func (bot *Bot) SendUserMessage(userID, message string) error {
	_, _, channelID, err := bot.slack.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	return bot.SendMessage(channelID, message, nil)
}

//HandleJoin handles comedian joining channel
func (bot *Bot) HandleJoin(channelID, teamID string) (model.Channel, error) {
	newChannel := model.Channel{}
	newChannel, err := bot.db.SelectChannel(channelID)
	if err == nil {
		return newChannel, nil
	}

	channel, err := bot.slack.GetConversationInfo(channelID, true)
	if err != nil {
		return newChannel, err
	}
	newChannel, err = bot.db.CreateChannel(model.Channel{
		TeamID:      teamID,
		ChannelName: channel.Name,
		ChannelID:   channel.ID,
		StandupTime: "",
	})
	if err != nil {
		return newChannel, err
	}
	return newChannel, nil
}

//HandleJoinNewUser handles comedian joining user
func (bot *Bot) HandleJoinNewUser(user slack.User) (model.User, error) {
	newUser, err := bot.db.SelectUser(user.ID)
	if err == nil {
		return newUser, nil
	}

	if user.IsBot || user.Name == "slackbot" {
		return newUser, nil
	}

	newUser, err = bot.db.CreateUser(model.User{
		TeamID:   user.TeamID,
		UserName: user.Name,
		UserID:   user.ID,
		RealName: user.RealName,
		TZ:       user.TZ,
		TZOffset: user.TZOffset,
	})
	if err != nil {
		return newUser, err
	}
	return newUser, nil
}

//ImplementCommands implements slash commands such as adding users and managing deadlines
func (bot *Bot) ImplementCommands(command slack.SlashCommand) string {
	switch command.Command {
	case "/start":
		return bot.joinCommand(command)
	case "/show":
		return bot.showCommand(command)
	case "/quit":
		return bot.quitCommand(command)
	case "/update_deadline":
		return bot.addDeadline(command)
	default:
		return ""
	}
}

//UpdateUsersList updates users in workspace
func (bot *Bot) UpdateUsersList() error {
	users, err := bot.slack.GetUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		err := bot.updateUser(user)
		if err != nil {
			log.WithFields(log.Fields{"user": user, "bot": bot, "error": err}).Error("updateUser failed")
		}
	}
	return nil
}

//AddNewSlackUser add new added slack user to db
func (bot *Bot) AddNewSlackUser(userID string) error {
	users, err := bot.slack.GetUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.ID == userID {
			err := bot.updateUser(user)
			if err != nil {
				log.WithFields(log.Fields{"user": user, "bot": bot, "error": err}).Error("AddNewSlackUser.updateUser failed")
			}
		}
	}
	return nil
}

func (bot *Bot) updateUser(user slack.User) error {
	if user.IsBot || user.Name == "slackbot" {
		return nil
	}

	u, err := bot.db.SelectUser(user.ID)
	if err != nil && !user.Deleted {
		u, err = bot.db.CreateUser(model.User{
			TeamID:   user.TeamID,
			UserName: user.Name,
			UserID:   user.ID,
			RealName: user.RealName,
			TZ:       user.TZ,
			TZOffset: user.TZOffset,
			Status:   strings.ToLower(user.Profile.StatusText),
		})
		if err != nil {
			return err
		}
	}

	if !user.Deleted {
		u.UserName = user.Name
		u.RealName = user.RealName
		u.TeamID = user.TeamID
		u.TZ = user.TZ
		u.TZOffset = user.TZOffset
		u.Status = strings.ToLower(user.Profile.StatusText)
		_, err = bot.db.UpdateUser(u)
		if err != nil {
			return err
		}
	}

	if user.Deleted {
		err := bot.db.DeleteUser(u.ID)
		if err != nil {
			return err
		}

		standupers, err := bot.db.ListStandupers()
		if err != nil {
			return err
		}
		for _, standuper := range standupers {
			if u.UserID == standuper.UserID {
				err := bot.db.DeleteStanduper(standuper.ID)
				if err != nil {
					log.WithFields(log.Fields{"user": user, "bot": bot, "standuper": standuper, "error": err}).Error("DeleteStanduper failed")
				}
			}
		}
	}

	return nil
}

//Suits returns true if found desired bot properties
func (bot *Bot) Suits(team string) bool {
	return strings.ToLower(team) == strings.ToLower(bot.properties.TeamID) || strings.ToLower(team) == strings.ToLower(bot.properties.TeamName)
}

//Settings just returns bot settings
func (bot *Bot) Settings() model.BotSettings {
	return bot.properties
}

//SetProperties updates bot settings
func (bot *Bot) SetProperties(settings model.BotSettings) model.BotSettings {
	bot.properties = settings
	return bot.properties
}

func (bot *Bot) remindAboutWorklogs() error {
	if time.Now().AddDate(0, 0, 1).Day() != 1 {
		return nil
	}

	if time.Now().Hour() != 10 || time.Now().Minute() != 0 {
		return nil
	}

	users, err := bot.db.ListUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.TeamID != bot.properties.TeamID {
			continue
		}

		standupers, err := bot.db.FindStansupersByUserID(user.UserID)
		if err != nil {
			log.Error(err)
			continue
		}

		if len(standupers) < 1 {
			continue
		}

		_, _, err = bot.GetCollectorDataOnMember(standupers[0], time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local), time.Now())
		if err != nil {
			log.Error(err)
			continue
		}

		message := "Сегодня последний день месяца. Пожалуйста, перепроверьте ворклоги!\n"
		var total int

		for _, member := range standupers {
			user, userInProject, err := bot.GetCollectorDataOnMember(member, time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local), time.Now())
			if err != nil {
				log.Error(err)
				continue
			}

			message += fmt.Sprintf("%s залогано %.2f\n", member.ChannelName, float32(userInProject.Worklogs)/3600)
			total = user.Worklogs
		}

		message += fmt.Sprintf("В общем: %.2f", float32(total)/3600)

		err = bot.SendUserMessage(user.UserID, message)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}
