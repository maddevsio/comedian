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
	localizer       *i18n.Localizer
)

// Slack struct used for storing and communicating with slack api
type Slack struct {
	Chat
	api        *slack.Client
	rtm        *slack.RTM
	wg         sync.WaitGroup
	myUsername string
	db         *storage.MySQL
}

// NewSlack creates a new copy of slack handler
func NewSlack(conf config.Config) (*Slack, error) {
	m, err := storage.NewMySQL(conf)
	if err != nil {
		logrus.Errorf("database connection: %v", err)
		return nil, err
	}
	s := &Slack{}
	s.api = slack.New(conf.SlackToken)
	s.rtm = s.api.NewRTM()
	s.db = m

	localizer, err = config.GetLocalizer()
	if err != nil {
		logrus.Errorf("get localizer: %v", err)
		return nil, err
	}
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
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "HelloManager"})
				if err != nil {
					logrus.Errorf("localize text: %v", err)
				}
				s.SendUserMessage("UB9AE7CL9", text)

			case *slack.MessageEvent:
				s.handleMessage(ev)
			case *slack.PresenceChangeEvent:
				logrus.Infof("Presence Change: %v\n", ev)

			case *slack.RTMError:
				logrus.Errorf("RTME: %v", ev)

			case *slack.InvalidAuthEvent:
				logrus.Info("Invalid credentials")
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
				ChannelID:  msg.Msg.Channel,
				UsernameID: msg.Msg.User,
				Username:   msg.Msg.Username,
				Comment:    standupText,
				MessageTS:  msg.Msg.Timestamp,
			})
			var text string
			if err != nil {
				logrus.Errorf("create standup: %v", err)
				return err
			}
			text, err = localizer.Localize(&i18n.LocalizeConfig{MessageID: "StandupAccepted"})
			if err != nil {
				logrus.Errorf("localize text: %v", err)
			}
			return s.SendMessage(msg.Msg.Channel, text)

		}
	case typeEditMessage:
		standup, err := s.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			logrus.Errorf("select standup by message TS: %v", err)
			return err
		}
		_, err = s.db.AddToStandupHistory(model.StandupEditHistory{
			StandupID:   standup.ID,
			StandupText: standup.Comment})
		if err != nil {
			logrus.Errorf("add to standup history: %v", err)
			return err
		}
		if standupText, ok := s.isStandup(msg.SubMessage.Text); ok {
			standup.Comment = standupText

			standup, err = s.db.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("update standup: %v", err)
				return err
			}
			logrus.Infof("standup updated: %v", standup)
		}
	}
	return nil
}

func (s *Slack) isStandup(message string) (string, bool) {

	p1, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p1"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	p2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p2"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	p3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p3"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}

	mentionsProblem := false
	problemKeys := []string{p1, p2, p3}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}

	y1, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y1"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	y2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y2"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	y3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y3"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	y4, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y4"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{y1, y2, y3, y4}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	t1, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "t1"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	t2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "t2"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	t3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "t3"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	mentionsTodayPlans := false
	todayPlansKeys := []string{t1, t2, t3}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}

	if mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans {
		logrus.Infof("this is a standup: %v", message)
		return strings.TrimSpace(message), true
	}

	logrus.Errorf("This is not a standup: %v", message)
	return message, false

}

// SendMessage posts a message in a specified channel
func (s *Slack) SendMessage(channel, message string) error {
	_, _, err := s.api.PostMessage(channel, message, slack.PostMessageParameters{})
	if err != nil {
		logrus.Errorf("post message: %v", err)
		return err
	}
	return err
}

// SendUserMessage posts a message to a specific user
func (s *Slack) SendUserMessage(userID, message string) error {
	_, _, channelID, err := s.api.OpenIMChannel(userID)
	logrus.Println(channelID)
	if err != nil {
		logrus.Errorf("open IM channel: %v", err)
		return err
	}
	err = s.SendMessage(channelID, message)
	if err != nil {
		logrus.Errorf("send message: %v", err)
	}
	return err
}
