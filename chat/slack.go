package chat

import (
	"fmt"
	"os"
	"sync"

	"log"

	"github.com/maddevsio/comedian/config"
	"github.com/nlopes/slack"
)

// Slack struct used for storing and communicating with slack api
type Slack struct {
	Chat
	api    *slack.Client
	logger *log.Logger
	rtm    *slack.RTM
	wg     sync.WaitGroup
}

// NewSlack creates a new copy of slack handler
func NewSlack(conf config.Config) (*Slack, error) {
	s := &Slack{}
	s.api = slack.New(conf.SlackToken)
	s.logger = log.New(os.Stdout, "comedian: ", log.Lshortfile|log.LstdFlags)
	s.rtm = s.api.NewRTM()
	slack.SetLogger(s.logger)
	return s, nil
}

// Run runs a listener loop for slack
func (s *Slack) Run() error {
	s.ManageConnection()
	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello
		case *slack.ConnectedEvent:
			s.api.PostMessage("#standup", "<!channel> Hello world", slack.PostMessageParameters{})

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)

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
	return nil
}

func (s *Slack) ManageConnection() {
	s.wg.Add(1)
	go func() {
		s.rtm.ManageConnection()
		s.wg.Done()
	}()

}
