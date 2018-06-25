package chat

import (
	"fmt"

	"sync"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"

	"strings"
)

var (
	typeMessage     = ""
	typeEditMessage = "message_changed"
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
				s.api.PostMessage("CBAP453GV", "Hey! I am alive!", slack.PostMessageParameters{})

			case *slack.MessageEvent:
				s.handleMessage(ev)
			case *slack.PresenceChangeEvent:
				s.logger.Infof("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				s.logger.Infof("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				s.logger.Errorf(ev.Error())

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
func (s *Slack) handleMessage(msg *slack.MessageEvent) {

	userName := s.rtm.GetInfo().GetUserByID(msg.Msg.User)
	// TODO: check if channel message (not direct)
	switch msg.SubType {
	case typeMessage:
		if cleanMsg, ok := s.cleanMessage(msg.Msg.Text); ok {
			channelName := s.rtm.GetInfo().GetChannelByID(msg.Msg.Channel)
			_, err := s.db.CreateStandup(model.Standup{
				ChannelID:  msg.Msg.Channel,
				Channel:    channelName.Name,
				UsernameID: msg.Msg.User,
				Username:   userName.Name,
				FullName:   fmt.Sprintf("%s %s", userName.Profile.FirstName, userName.Profile.LastName),
				Comment:    cleanMsg,
				MessageTS:  msg.Msg.Timestamp,
			})
			if err != nil {
				s.logger.Error(err)
			}
			msg := "Good job! Standup accepted! Keep it up!"
			s.logger.Info(msg)
			s.SendMessage(channelName.Name, msg)

		}
	case typeEditMessage:
		standup, err := s.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			s.logger.Error(err)
		}
		_, err = s.db.AddToStandupHistory(model.StandupEditHistory{
			StandupID:   standup.ID,
			StandupText: standup.Comment})
		if err != nil {
			s.logger.Error(err)
		}
		if cleanMsg, ok := s.cleanMessage(msg.SubMessage.Text); ok {
			standup.Comment = cleanMsg

			_, err = s.db.UpdateStandup(standup)
			if err != nil {
				s.logger.Error(err)
			}
			s.logger.Info("Edited")
		}
	}
}

func (s *Slack) cleanMessage(message string) (string, bool) {
	if strings.Contains(message, fmt.Sprintf("<@%s>", s.myUsername)) {
		msg := strings.Replace(message, fmt.Sprintf("<@%s>", s.myUsername), "", -1)
		return strings.TrimSpace(msg), true
	}
	return message, false
}

// SendMessage posts a message in a specified channel
func (s *Slack) SendMessage(channel, message string) error {
	_, _, err := s.api.PostMessage(channel, message, slack.PostMessageParameters{})
	return err
}
