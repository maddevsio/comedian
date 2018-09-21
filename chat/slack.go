package chat

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
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
	api  *slack.Client
	rtm  *slack.RTM
	wg   sync.WaitGroup
	db   *storage.MySQL
	Conf config.Config
}

// NewSlack creates a new copy of slack handler
func NewSlack(conf config.Config) (*Slack, error) {
	m, err := storage.NewMySQL(conf)

	if err != nil {
		logrus.Errorf("slack: NewMySQL failed: %v\n", err)
		return nil, err
	}

	s := &Slack{}
	s.Conf = conf
	s.api = slack.New(conf.SlackToken)
	s.rtm = s.api.NewRTM()
	s.db = m
	return s, nil
}

// Run runs a listener loop for slack
func (s *Slack) Run() {
	s.wg.Add(1)
	go s.rtm.ManageConnection()
	s.wg.Done()

	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			fmt.Printf("Team monitoring enabled: %v\n", s.Conf.TeamMonitoringEnabled)
			s.UpdateUsersList()
			s.UpdateChannelsList()
			s.handleConnection()
			gocron.Every(1).Day().At("23:55").Do(s.FillStandupsForNonReporters)
			gocron.Start()
		case *slack.MessageEvent:
			s.handleMessage(ev)
		case *slack.PresenceChangeEvent:
			logrus.Infof("slack: Presence Change: %v\n", ev)
		case *slack.RTMError:
			logrus.Errorf("slack: RTME: %v\n", ev)
		case *slack.InvalidAuthEvent:
			logrus.Info("slack: Invalid credentials")
			return
		}
	}
}

func (s *Slack) handleConnection() error {
	s.SendUserMessage(s.Conf.ManagerSlackUserID, s.Conf.Translate.HelloManager)
	return nil
}

func (s *Slack) handleMessage(msg *slack.MessageEvent) error {
	logrus.Infof("New message! %v", msg)
	switch msg.SubType {
	case typeMessage:
		if standupText, ok := s.isStandup(msg.Msg.Text); ok {
			standup, err := s.db.CreateStandup(model.Standup{
				ChannelID: msg.Channel,
				UserID:    msg.User,
				Comment:   standupText,
				MessageTS: msg.Msg.Timestamp,
			})
			if err != nil {
				logrus.Errorf("slack: CreateStandup failed: %v\n", err)
				return err
			}
			logrus.Infof("slack: Standup created: %v\n", standup)
			item := slack.ItemRef{msg.Channel, msg.Msg.Timestamp, "", ""}
			err = s.api.AddReaction("heavy_check_mark", item)
			if err != nil {
				logrus.Errorf("slack: AddReaction failed: %v", err)
				return err
			}
			return nil
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
	problemKeys := []string{s.Conf.Translate.P1, s.Conf.Translate.P2, s.Conf.Translate.P3}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{s.Conf.Translate.Y1, s.Conf.Translate.Y2, s.Conf.Translate.Y3, s.Conf.Translate.Y4}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{s.Conf.Translate.T1, s.Conf.Translate.T2, s.Conf.Translate.T3}
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
	return err
}

// SendUserMessage posts a message to a specific user
func (s *Slack) SendUserMessage(userID, message string) error {
	_, _, channelID, err := s.api.OpenIMChannel(userID)
	if err != nil {
		logrus.Errorf("slack: OpenIMChannel failed: %v\n", err)
		return err
	}
	err = s.SendMessage(channelID, message)
	if err != nil {
		logrus.Errorf("slack: SendMessage failed: %v\n", err)
		return err
	}
	return err
}

//UpdateUsersList updates users in workspace
func (s *Slack) UpdateUsersList() {
	users, _ := s.api.GetUsers()
	for _, user := range users {
		if user.IsBot || user.Name == "slackbot" {
			continue
		}

		u, err := s.db.SelectUser(user.ID)
		if err != nil {
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				s.db.CreateUser(model.User{
					UserName: user.Name,
					UserID:   user.ID,
					Role:     "admin",
				})
				continue
			}
			s.db.CreateUser(model.User{
				UserName: user.Name,
				UserID:   user.ID,
				Role:     "",
			})
			continue
		}
		if user.Deleted {
			s.db.DeleteUser(u.ID)
		}
		continue
	}
}

//UpdateChannelsList updates users in workspace
func (s *Slack) UpdateChannelsList() {
	params := slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
	}
	chans, _, _ := s.api.GetConversations(&params)
	for _, c := range chans {
		_, err := s.db.SelectChannel(c.ID)
		if err != nil {
			s.db.CreateChannel(model.Channel{
				ChannelName: c.Name,
				ChannelID:   c.ID,
				StandupTime: int64(0),
			})
		}
		continue
	}
}

//FillStandupsForNonReporters fills standup entries with empty standups to later recognize
//non reporters vs those who did not have to write standups
func (s *Slack) FillStandupsForNonReporters() {
	logrus.Println("FillStandupsForNonReporters!")
	//check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	timeFrom := time.Now().AddDate(0, 0, -1)
	allUsers, err := s.db.ListAllChannelMembers()
	if err != nil {
		logrus.Errorf("notifier: s.GetCurrentDayNonReporters failed: %v\n", err)
		return
	}
	for _, user := range allUsers {
		isNonReporter, err := s.db.IsNonReporter(user.UserID, user.ChannelID, timeFrom, time.Now())
		if err != nil {
			logrus.Errorf("notifier: IsNonReporter failed: %v\n", err)
			return
		}
		if isNonReporter {
			standup, err := s.db.CreateStandup(model.Standup{
				ChannelID: user.ChannelID,
				UserID:    user.UserID,
				Comment:   "",
				MessageTS: strconv.Itoa(int(time.Now().Unix())),
			})
			if err != nil {
				logrus.Errorf("notifier: CreateStandup failed: %v\n", err)
				return
			}
			logrus.Infof("notifier: Empty Standup created: %v\n", standup)
		}
	}
}
