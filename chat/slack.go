package chat

import (
	"sync"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"

	"strings"
)

var (
	typeMessage     = ""
	typeEditMessage = "message_changed"
	localizer, _    = config.GetLocalizer()
)

// Slack struct used for storing and communicating with slack api
type Slack struct {
	Chat
	api        *slack.Client
	logger     *logrus.Logger
	rtm        *slack.RTM
	wg         sync.WaitGroup
	myUsername string
	db         *storage.MySQL
}

// NewSlack creates a new copy of slack handler
func NewSlack(conf config.Config) (*Slack, error) {
	m, err := storage.NewMySQL(conf)
	if err != nil {
		logrus.Errorf("ERROR: %s", err.Error())
		return nil, err
	}
	s := &Slack{}
	s.api = slack.New(conf.SlackToken)
	s.logger = logrus.New()
	s.rtm = s.api.NewRTM()
	s.db = m

	return s, nil
}

// Run runs a listener loop for slack
func (s *Slack) Run() error {

	s.ManageConnection()
	for {
		if s.myUsername == "" {
			info := s.rtm.GetInfo()
			if info != nil {
				s.myUsername = info.User.ID
			}
		}
		select {
		case msg := <-s.rtm.IncomingEvents:

			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				s.GetAllUsersToDB()
				//	s.api.PostMessage("CBAP453GV", "Hey! I am alive!", slack.PostMessageParameters{})
				s.SendUserMessage("UB9AE7CL9", localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "HelloManager"}))

			case *slack.MessageEvent:
				s.handleMessage(ev)
			case *slack.PresenceChangeEvent:
				s.logger.Infof("Presence Change: %v\n", ev)

			// case *slack.LatencyReport:
			// 	s.logger.Infof("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				logrus.Errorf("ERROR: %s", ev.Error())

			case *slack.InvalidAuthEvent:
				s.logger.Info("Invalid credentials")
				return nil
			}
		}
	}
}

// ManageConnection manages connection
func (s *Slack) ManageConnection() {
	s.wg.Add(1)
	go func() {
		s.rtm.ManageConnection()
		s.wg.Done()
	}()

}

func (s *Slack) handleMessage(msg *slack.MessageEvent) error {

	switch msg.SubType {
	case typeMessage:
		if standupText, ok := s.isStandup(msg.Msg.Text); ok {
			_, err := s.db.CreateStandup(model.Standup{
				Channel:    msg.Msg.Channel,
				UsernameID: msg.Msg.User,
				Username:   msg.Msg.Username,
				Comment:    standupText,
				MessageTS:  msg.Msg.Timestamp,
			})
			var text string
			if err != nil {
				logrus.Errorf("ERROR: %s", err.Error())
				text = err.Error()
				return err
			} else {
				text = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StandupAccepted"})
			}
			return s.SendMessage(msg.Msg.Channel, text)

		}
	case typeEditMessage:
		standup, err := s.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			logrus.Errorf("ERROR: %s", err.Error())
			return err
		}
		_, err = s.db.AddToStandupHistory(model.StandupEditHistory{
			StandupID:   standup.ID,
			StandupText: standup.Comment})
		if err != nil {
			logrus.Errorf("ERROR: %s", err.Error())
			return err
		}
		if standupText, ok := s.isStandup(msg.SubMessage.Text); ok {
			standup.Comment = standupText

			_, err = s.db.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("ERROR: %s", err.Error())
				return err
			}
			s.logger.Info("Edited")
		}
	}
	return nil
}

func (s *Slack) isStandup(message string) (string, bool) {
	if (strings.Contains(message, "роблем") || strings.Contains(message, "рудност") || strings.Contains(message, "атрдуднен")) && (strings.Contains(message, "чера") || strings.Contains(message, "ятницу") || strings.Contains(message, "делал") || strings.Contains(message, "делано")) && (strings.Contains(message, "егодн") || strings.Contains(message, "обираюс")) {
		logrus.Infof("This message is a standup: %v", message)
		return strings.TrimSpace(message), true

	}
	logrus.Errorf("This message is not a standup: %v", message)
	return message, false

}

// SendMessage posts a message in a specified channel
func (s *Slack) SendMessage(channel, message string) error {
	_, _, err := s.api.PostMessage(channel, message, slack.PostMessageParameters{})
	return err
}

// SendUserMessage posts a message to a specific user
func (s *Slack) SendUserMessage(userID, message string) error {
	_, _, channelID, err := s.api.OpenIMChannel(userID)
	logrus.Println(channelID)
	if err != nil {
		logrus.Errorf("ERROR: %s", err.Error())
		return err
	}
	err = s.SendMessage(channelID, message)
	return err
}

// GetAllUsersToDB selects all the users in the organization and sync them to db
func (s *Slack) GetAllUsersToDB() error {
	users, err := s.api.GetUsers()
	if err != nil {
		logrus.Errorf("ERROR: %s", err.Error())
		return err
	}
	chans, err := s.api.GetChannels(false)
	if err != nil {
		logrus.Error("ERROR: %s", err.Error())
		return err
	}
	var channelID string
	for _, channel := range chans {
		if channel.Name == "general" {
			channelID = channel.ID
		}
	}
	for _, user := range users {
		_, err := s.db.FindStandupUserInChannel(user.Name, channelID)
		if err != nil {
			s.db.CreateStandupUser(model.StandupUser{
				SlackUserID: user.ID,
				SlackName:   user.Name,
				ChannelID:   "",
				Channel:     "",
			})
		}

	}
	return nil
}
