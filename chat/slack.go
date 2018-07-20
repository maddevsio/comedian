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
		logrus.Errorf("slack: NewMySQL failed: %v\n", err)
		return nil, err
	}
	logrus.Infof("slack: mysql connection: %v\n", m)
	s := &Slack{}
	s.api = slack.New(conf.SlackToken)
	s.rtm = s.api.NewRTM()
	s.db = m

	localizer, err = config.GetLocalizer()
	if err != nil {
		logrus.Errorf("slack: GetLocalizer failed: %v\n", err)
		return nil, err
	}
	logrus.Infof("slack: new Slack: %v\n", s)
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
					logrus.Errorf("slack: Localize failed: %v\n", err)
				}
				c, err := config.Get()
				if err != nil {
					logrus.Errorf("slack: GetConfig: %v\n", err)
					return err
				}
				s.SendUserMessage(c.ManagerSlackUserID, text)

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
	user, err := s.db.FindStandupUserInChannelByUserID(msg.Msg.User, msg.Msg.Channel)
	if err != nil {
		logrus.Errorf("slack: FindStandupUserInChannelByUserID failed: %v,\n User:%v,\n channel:%v", err, msg.Msg.User, msg.Msg.Channel)
	}
	switch msg.SubType {
	case typeMessage:
		if standupText, ok := s.isStandup(msg.Msg.Text); ok {
			standup, err := s.db.CreateStandup(model.Standup{
				Channel:    user.Channel,
				ChannelID:  user.ChannelID,
				UsernameID: user.SlackUserID,
				Username:   user.SlackName,
				Comment:    standupText,
				MessageTS:  msg.Msg.Timestamp,
			})
			logrus.Infof("slack: Standup created: %v\n", standup)
			var text string
			if err != nil {
				logrus.Errorf("slack: CreateStandup failed: %v\n", err)
				return err
			}
			text, err = localizer.Localize(&i18n.LocalizeConfig{MessageID: "StandupAccepted"})
			if err != nil {
				logrus.Errorf("slack: Localize failed: %v\n", err)
			}
			return s.SendMessage(msg.Msg.Channel, text)
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

	p1, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p1"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	p2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p2"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	p3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "p3"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
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
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	y2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y2"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	y3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y3"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	y4, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "y4"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
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
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	t2, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "t2"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	t3, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "t3"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
	}
	mentionsTodayPlans := false
	todayPlansKeys := []string{t1, t2, t3}
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
