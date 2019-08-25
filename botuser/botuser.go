package botuser

import (
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

//Message represent any message that can be send to Slack or any other destination
type Message struct {
	Type        string
	Channel     string
	User        string
	Text        string
	Attachments []slack.Attachment
}

// Bot struct used for storing and communicating with slack api
type Bot struct {
	conf        *config.Config
	db          *storage.DB
	localizer   *i18n.Localizer
	properties  *model.BotSettings
	slack       *slack.Client
	quitChan    chan struct{}
	messageChan chan Message
}

//New creates new Bot instance
func New(config *config.Config, bundle *i18n.Bundle, settings *model.BotSettings, db *storage.DB) *Bot {
	bot := &Bot{
		conf:       config,
		db:         db,
		slack:      slack.New(settings.AccessToken),
		properties: settings,
		localizer:  i18n.NewLocalizer(bundle, settings.Language),
	}

	bot.quitChan = make(chan struct{})
	bot.messageChan = make(chan Message)

	return bot
}

//Start updates Users list and launches notifications
func (bot *Bot) Start() {
	var wg sync.WaitGroup

	log.Info("Bot started for ", bot.properties.TeamName)

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second * 60).C
		for {
			select {
			case <-ticker:
				err := bot.notifyChannels()
				if err != nil {
					log.Error("notifyChannels failed: ", err)
				}
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
			case <-bot.quitChan:
				wg.Done()
				return
			case msg := <-bot.messageChan:
				err := bot.send(msg)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}()
}

func (bot *Bot) send(msg Message) error {
	if msg.Type == "message" {
		err := bot.SendMessage(msg.Channel, msg.Text, msg.Attachments)
		if err != nil {
			return err
		}
	}
	if msg.Type == "ephemeral" {
		err := bot.SendEphemeralMessage(msg.Channel, msg.User, msg.Text)
		if err != nil {
			return err
		}
	}
	if msg.Type == "direct" {
		err := bot.SendUserMessage(msg.User, msg.Text)
		if err != nil {
			return err
		}
	}

	return nil
}

//Stop closes bot quitChan making bot goroutine to exit
func (bot *Bot) Stop() {
	close(bot.quitChan)
}

//HandleMessage handles slack message event
func (bot *Bot) HandleMessage(msg *slack.MessageEvent) error {
	if !strings.Contains(msg.Msg.Text, bot.properties.UserID) {
		return nil
	}
	msg.Team = bot.properties.TeamID
	switch msg.SubType {
	case typeMessage:
		_, err := bot.handleNewMessage(msg)
		if err != nil {
			return err
		}
	case typeEditMessage:
		_, err := bot.handleEditMessage(msg)
		if err != nil {
			return err
		}
	case typeDeleteMessage:
		_, err := bot.handleDeleteMessage(msg)
		if err != nil {
			return err
		}
	case "bot_message":
		return nil
	}
	return nil
}

func (bot *Bot) handleNewMessage(msg *slack.MessageEvent) (string, error) {

	problem := bot.analizeStandup(msg.Msg.Text)
	if problem != "" {
		bot.messageChan <- Message{
			Type:    "ephemeral",
			Channel: msg.Channel,
			User:    msg.User,
			Text:    problem,
		}
		return problem, nil
	}

	_, err := bot.db.CreateStandup(model.Standup{
		Created:   time.Now().UTC(),
		Modified:  time.Now().UTC(),
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.User,
		Comment:   msg.Msg.Text,
		MessageTS: msg.Msg.Timestamp,
	})
	if err != nil {
		return "", err
	}
	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.Msg.Timestamp,
		File:      "",
		Comment:   "",
	}
	err = bot.slack.AddReaction("heavy_check_mark", item)
	if err != nil {
		return "", err
	}
	return "standup saved", nil
}

func (bot *Bot) handleEditMessage(msg *slack.MessageEvent) (string, error) {
	problem := bot.analizeStandup(msg.SubMessage.Text)
	if problem != "" {
		bot.messageChan <- Message{
			Type:    "ephemeral",
			Channel: msg.Channel,
			User:    msg.User,
			Text:    problem,
		}
		return problem, nil
	}

	standup, err := bot.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
	if err == nil {
		standup.Comment = msg.SubMessage.Text
		standup.Modified = time.Now().UTC()
		_, err := bot.db.UpdateStandup(standup)
		if err != nil {
			return "", err
		}
		return "standup updated", nil
	}

	standup, err = bot.db.CreateStandup(model.Standup{
		Created:   time.Now().UTC(),
		Modified:  time.Now().UTC(),
		TeamID:    msg.Team,
		ChannelID: msg.Channel,
		UserID:    msg.SubMessage.User,
		Comment:   msg.SubMessage.Text,
		MessageTS: msg.SubMessage.Timestamp,
	})
	if err != nil {
		return "", err
	}

	item := slack.ItemRef{
		Channel:   msg.Channel,
		Timestamp: msg.SubMessage.Timestamp,
		File:      "",
		Comment:   "",
	}
	err = bot.slack.AddReaction("heavy_check_mark", item)
	if err != nil {
		return "", err
	}

	return "standup created", nil
}

func (bot *Bot) handleDeleteMessage(msg *slack.MessageEvent) (string, error) {
	standup, err := bot.db.SelectStandupByMessageTS(msg.DeletedTimestamp)
	if err != nil {
		return "", nil
	}

	err = bot.db.DeleteStandup(standup.ID)
	if err != nil {
		return "", err
	}

	return "standup deleted", nil
}

func (bot *Bot) submittedStandupToday(userID, channelID string) bool {
	standup, err := bot.db.SelectLatestStandupByUser(userID, channelID)
	if err != nil {
		return false
	}

	userProfile, err := bot.slack.GetUserInfo(userID)
	if err != nil {
		log.Error(err)
		return false
	}

	loc := time.FixedZone(userProfile.TZ, userProfile.TZOffset)

	if standup.Created.In(loc).Day() == time.Now().UTC().In(loc).Day() {
		log.Info("not non reporter: ", userID)
		return true
	}
	return false
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
	msgParams := slack.NewPostMessageParameters()
	msgParams.Attachments = attachments
	_, _, err := bot.slack.PostMessage(channel, message, msgParams)
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

//Suits returns true if found desired bot properties
func (bot *Bot) Suits(team string) bool {
	return strings.ToLower(team) == strings.ToLower(bot.properties.TeamID) || strings.ToLower(team) == strings.ToLower(bot.properties.TeamName)
}

//Settings just returns bot settings
func (bot *Bot) Settings() *model.BotSettings {
	return bot.properties
}

//SetProperties updates bot settings
func (bot *Bot) SetProperties(settings *model.BotSettings) *model.BotSettings {
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

	users, err := bot.slack.GetUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.TeamID != bot.properties.TeamID {
			continue
		}

		standupers, err := bot.db.FindStansupersByUserID(user.ID)
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

		bot.messageChan <- Message{
			Type: "direct",
			User: user.ID,
			Text: message,
		}
	}

	return nil
}
