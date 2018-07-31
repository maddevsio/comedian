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
)

// Chat inteface should be implemented for all messengers(facebook, slack, telegram, whatever)
type (
	// Slack struct used for storing and communicating with slack api
	Slack struct {
		Chat
		api        *slack.Client
		rtm        *slack.RTM
		wg         sync.WaitGroup
		myUsername string
		db         *storage.MySQL
		T          Translation
	}

	//Translation struct to get translation data
	Translation struct {
		HelloManager    string
		StandupAccepted string

		p1 string
		p2 string
		p3 string

		y1 string
		y2 string
		y3 string
		y4 string

		t1 string
		t2 string
		t3 string
	}
)

// NewSlack creates a new copy of slack handler
func NewSlack(conf config.Config) (*Slack, error) {
	m, err := storage.NewMySQL(conf)
	if err != nil {
		logrus.Errorf("slack: NewMySQL failed: %v\n", err)
		return nil, err
	}

	s := &Slack{}
	s.api = slack.New(conf.SlackToken)
	s.rtm = s.api.NewRTM()
	s.db = m
	t, err := getTranslation()
	if err != nil {
		logrus.Errorf("slack: getTranslation failed: %v\n", err)
		return nil, err
	}
	s.T = t

	logrus.Infof("slack: new Slack: %v\n", s)
	return s, nil
}

// Run runs a listener loop for slack
func (s *Slack) Run() error {

	s.wg.Add(1)
	go s.rtm.ManageConnection()
	s.wg.Done()

	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			s.handleConnection()
		case *slack.MessageEvent:
			s.handleMessage(ev)
		case *slack.PresenceChangeEvent:
			logrus.Infof("slack: Presence Change: %v\n", ev)
		case *slack.RTMError:
			logrus.Errorf("slack: RTME: %v\n", ev)
		case *slack.InvalidAuthEvent:
			logrus.Info("slack: Invalid credentials")
			return nil
		}
	}
	return nil
}

func (s *Slack) handleConnection() error {
	c, err := config.Get()
	if err != nil {
		logrus.Errorf("slack: GetConfig: %v\n", err)
		return err
	}
	s.SendUserMessage(c.ManagerSlackUserID, s.T.HelloManager)
	return nil
}

func (s *Slack) handleMessage(msg *slack.MessageEvent) error {
	switch msg.SubType {
	case typeMessage:
		if standupText, ok := s.isStandup(msg.Msg.Text); ok {
			standup, err := s.db.CreateStandup(model.Standup{
				ChannelID:  msg.Channel,
				UsernameID: msg.User,
				Comment:    standupText,
				MessageTS:  msg.Msg.Timestamp,
			})
			logrus.Infof("slack: Standup created: %v\n", standup)
			if err != nil {
				logrus.Errorf("slack: CreateStandup failed: %v\n", err)
				return err
			}
			return s.SendMessage(msg.Msg.Channel, s.T.StandupAccepted)
		}
	case typeEditMessage:
		standup, err := s.db.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			logrus.Errorf("slack: SelectStandupByMessageTS failed: %v\n", err)
			return err
		}
		standupHistory, err := s.db.AddToStandupHistory(model.StandupEditHistory{
			StandupID:   standup.ID,
			StandupText: standup.Comment})
		if err != nil {
			logrus.Errorf("slack: AddToStandupHistory failed: %v\n", err)
			return err
		}
		logrus.Infof("slack: Slack standup history: %v\n", standupHistory)
		if standupText, ok := s.isStandup(msg.SubMessage.Text); ok {
			standup.Comment = standupText

			standup, err = s.db.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("slack: UpdateStandup failed: %v\n", err)
				return err
			}
			logrus.Infof("slack: standup updated: %v\n", standup)
		}
	}
	return nil
}

func (s *Slack) isStandup(message string) (string, bool) {

	mentionsProblem := false
	problemKeys := []string{s.T.p1, s.T.p2, s.T.p3}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{s.T.y1, s.T.y2, s.T.y3, s.T.y4}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{s.T.t1, s.T.t2, s.T.t3}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}

	if mentionsProblem && mentionsYesterdayWork && mentionsTodayPlans {
		logrus.Infof("slack: this is a standup: %v\n", message)
		return strings.TrimSpace(message), true
	}

	logrus.Errorf("slack: This is not a standup: %v\n", message)
	return message, false
}

// SendMessage posts a message in a specified channel
func (s *Slack) SendMessage(channel, message string) error {
	_, _, err := s.api.PostMessage(channel, message, slack.PostMessageParameters{})
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	logrus.Infof("slack: Slack message sent: chan:%v, message:%v\n", channel, message)
	return err
}

// SendUserMessage posts a message to a specific user
func (s *Slack) SendUserMessage(userID, message string) error {
	_, _, channelID, err := s.api.OpenIMChannel(userID)
	if err != nil {
		logrus.Errorf("slack: OpenIMChannel failed: %v\n", err)
		return err
	}
	logrus.Infof("slack: Slack OpenIMChannel: %v\n", userID)
	err = s.SendMessage(channelID, message)
	if err != nil {
		logrus.Errorf("slack: SendMessage failed: %v\n", err)
	}
	logrus.Info("slack: Message sent\n")
	return err
}

func getTranslation() (Translation, error) {
	localizer, err := config.GetLocalizer()
	if err != nil {
		logrus.Errorf("slack: GetLocalizer failed: %v\n", err)
		return Translation{}, err
	}
	m := make(map[string]string)
	r := []string{
		"HelloManager", "StandupAccepted",
		"p1", "p2", "p3",
		"y1", "y2", "y3", "y4",
		"t1", "t2", "t3",
	}

	for _, t := range r {
		translated, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: t})
		if err != nil {
			logrus.Errorf("slack: Localize failed: %v\n", err)
			return Translation{}, err
		}
		m[t] = translated
	}

	t := Translation{
		HelloManager:    m["HelloManager"],
		StandupAccepted: m["StandupAccepted"],

		p1: m["p1"],
		p2: m["p2"],
		p3: m["p3"],

		y1: m["y1"],
		y2: m["y2"],
		y3: m["y3"],
		y4: m["y4"],

		t1: m["t1"],
		t2: m["t2"],
		t3: m["t3"],
	}
	return t, nil
}
