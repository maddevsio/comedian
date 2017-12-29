package chat

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"log"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nlopes/slack"
)

var (
	typeMessage     = ""
	typeEditMessage = "message_changed"
)

// Slack struct used for storing and communicating with slack api
type Slack struct {
	Chat
	api        *slack.Client
	logger     *log.Logger
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
	s.logger = log.New(os.Stdout, "comedian: ", log.Lshortfile|log.LstdFlags)
	s.rtm = s.api.NewRTM()
	s.db = m
	slack.SetLogger(s.logger)
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
			case *slack.HelloEvent:
				// Ignore hello
			case *slack.ConnectedEvent:
				s.api.PostMessage("#general", "<!channel> Hello world", slack.PostMessageParameters{})

			case *slack.MessageEvent:
				s.handleMessage(ev)
			case *slack.PresenceChangeEvent:
				fmt.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				fmt.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
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
				fmt.Println(err)
			}
			fmt.Println("Standup accepted")
		}
	case typeEditMessage:
		standup, err := s.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			fmt.Println(err)
		}
		if cleanMsg, ok := s.cleanMessage(msg.SubMessage.Text); ok {
			standup.Comment = cleanMsg
			_, err := s.db.UpdateStandup(standup)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Edited")

		}

	}

}

func (s *Slack) cleanMessage(message string) (string, bool) {
	if strings.Contains(message, fmt.Sprintf("<@%s>", s.myUsername)) {
		msg := strings.Replace(message, fmt.Sprintf("<@%s>", s.myUsername), "", -1)

		return msg, true
	}
	return message, false
}

func (s *Slack) sendMessage(channel, message string) error {
	_, _, err := s.api.PostMessage(channel, message, slack.PostMessageParameters{})
	return err
}
