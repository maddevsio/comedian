package botuser

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"gitlab.com/team-monitoring/comedian/utils"
	"golang.org/x/text/language"
)

var (
	typeMessage         = ""
	typeEditMessage     = "message_changed"
	typeDeleteMessage   = "message_deleted"
	oneStandupPerDay    string
	couldNotSaveStandup string
)

const (
	adminAccess       = 2
	pmAccess          = 3
	regularUserAccess = 4
)

// Bot struct used for storing and communicating with slack api
type Bot struct {
	slack      *slack.Client
	Properties model.ControlPannel
	db         *storage.MySQL
	bundle     *i18n.Bundle
}

func New(cp model.ControlPannel, db *storage.MySQL) *Bot {
	bot := &Bot{}

	bot.slack = slack.New(cp.AccessToken)
	bot.Properties = cp
	bot.db = db

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFile("botuser/active.en.toml")
	bundle.LoadMessageFile("botuser/active.ru.toml")

	bot.bundle = bundle

	return bot
}

func (bot *Bot) Start() {
	bot.UpdateUsersList()

	go func(bot *Bot) {
		rtm := bot.slack.NewRTM()
		go rtm.ManageConnection()
		for msg := range rtm.IncomingEvents {
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				botUserID := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
				bot.HandleMessage(ev, botUserID)
			case *slack.ConnectedEvent:
				log.Info("Reconnected!")
			case *slack.MemberJoinedChannelEvent:
				bot.HandleJoin(ev.Channel, ev.Team)
			}
		}
	}(bot)

	go func(bot *Bot) {
		notificationForChannels := time.NewTicker(time.Second * 60).C
		for {
			select {
			case <-notificationForChannels:
				bot.NotifyChannels()
			}
		}
	}(bot)
}

func (bot *Bot) HandleMessage(msg *slack.MessageEvent, botUserID string) {

	localizer := i18n.NewLocalizer(bot.bundle, bot.Properties.Language)
	oneStandupPerDay = localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "OneStandupPerDay",
			Description: "Warning that only one standup per day is allowed",
			Other:       "You can submit only one standup per day. Please, edit today's standup or submit your next standup tomorrow!",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	couldNotSaveStandup = localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "CouldNotSaveStandup",
			Description: "Displays a message when unexpected errors occur",
			Other:       "Something went wrong and I could not save your standup in database. Please, report this to your PM.",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	switch msg.SubType {
	case typeMessage:
		err := bot.HandleNewMessage(msg, botUserID)
		if err != nil {
			log.Error(err)
		}

	case typeEditMessage:
		err := bot.HandleEditMessage(msg, botUserID)
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

func (bot *Bot) HandleNewMessage(msg *slack.MessageEvent, botUserID string) error {
	if !strings.Contains(msg.Msg.Text, botUserID) && !strings.Contains(msg.Msg.Text, "#standup") {
		return errors.New("bot is not mentioned and no #standup in the message body")
	}

	problem := bot.analizeStandup(msg.Msg.Text)
	if problem != "" {
		bot.SendEphemeralMessage(msg.Channel, msg.User, problem)
		return errors.New("Fail to save message as standup. Standup is not complete")
	}

	if bot.db.SubmittedStandupToday(msg.User, msg.Channel) {
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

func (bot *Bot) HandleEditMessage(msg *slack.MessageEvent, botUserID string) error {
	if !strings.Contains(msg.SubMessage.Text, botUserID) && !strings.Contains(msg.SubMessage.Text, "#standup") {
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
	localizer := i18n.NewLocalizer(bot.bundle, bot.Properties.Language)
	message = strings.ToLower(message)

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{"yesterday", "friday", "monday", "tuesday", "wednesday", "thursday", "saturday", "sunday", "completed", "вчера", "пятниц", "делал", "сделано", "понедельник", "вторник", "сред", "четверг", "суббот", "воскресенье"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	if !mentionsYesterdayWork {
		standupHandleNoYesterdayWorkMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoYesterdayWorkMentioned",
				Description: "No 'yesterday' keywords in standup",
				Other:       ":warning: No 'yesterday' related keywords detected! Please, use one of the following: 'yesterday' or weekdays such as 'friday' etc.",
			},
		})
		return standupHandleNoYesterdayWorkMentioned
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{"today", "going", "plan", "сегодня", "собираюсь", "план"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}
	if !mentionsTodayPlans {
		standupHandleNoTodayPlansMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoTodayPlansMentioned",
				Description: "No 'today' keywords in standup",
				Other:       ":warning: No 'today' related keywords detected! Please, use one of the following: 'today', 'going', 'plan'",
			},
		})
		return standupHandleNoTodayPlansMentioned
	}

	mentionsProblem := false

	problemKeys := []string{"problem", "difficult", "stuck", "question", "issue", "block", "проблем", "трудност", "затруднени", "вопрос", "блок"}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}
	if !mentionsProblem {
		standupHandleNoProblemsMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoProblemsMentioned",
				Description: "No 'problems' key in standup",
				Other:       ":warning: No 'problems' related keywords detected! Please, use one of the following: 'problem', 'difficult', 'stuck', 'question', 'issue'",
			},
		})
		return standupHandleNoProblemsMentioned
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

	log.Error("No such channel found! Will create one!")
	channel, err := bot.slack.GetConversationInfo(channelID, true)
	if err != nil {
		log.Errorf("GetConversationInfo failed: %v", err)
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
				bot.db.DeleteChannelMember(member.UserID, member.ChannelID)
			}
		}
	}
	log.Info("Users list updated successfully")
}

func (bot *Bot) Suits(team string) bool {
	if team == bot.Properties.TeamID || team == bot.Properties.TeamName {
		return true
	}
	return false
}
