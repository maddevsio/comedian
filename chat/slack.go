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

	fmt.Printf("Team monitoring enabled: %v\n", s.Conf.TeamMonitoringEnabled)
	s.UpdateUsersList()
	gocron.Every(1).Day().At("23:50").Do(s.FillStandupsForNonReporters)
	gocron.Start()

	s.wg.Add(1)
	go s.rtm.ManageConnection()
	s.wg.Done()

	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			fmt.Println("Reconnected!")
			s.SendUserMessage(s.Conf.ManagerSlackUserID, s.Conf.Translate.HelloManager)
		case *slack.MessageEvent:
			s.handleMessage(ev)
		case *slack.MemberJoinedChannelEvent:
			s.handleJoin(ev.Channel)
		case *slack.MemberLeftChannelEvent:
			logrus.Infof("type: %v, user: %v, channel: %v, channelType: %v, team: %v", ev.Type, ev.User, ev.Channel, ev.ChannelType, ev.Team)
			user, err := s.api.GetUserInfo(ev.User)
			if err != nil {
				logrus.Errorf("GetUserInfo failed: %v", err)
			}
			if user.IsBot && user.Name == "comedian" {
				logrus.Info("Comedian left the chat!")
			}
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

func (s *Slack) handleJoin(channelID string) {
	ch, err := s.db.SelectChannel(channelID)
	if err != nil {
		logrus.Error("No such channel found! Will create one!")
		channel, err := s.api.GetConversationInfo(channelID, true)
		if err != nil {
			logrus.Errorf("GetConversationInfo failed: %v", err)
		}
		createdChannel, err := s.db.CreateChannel(model.Channel{
			ChannelName: channel.Name,
			ChannelID:   channel.ID,
			StandupTime: int64(0),
		})
		if err != nil {
			logrus.Errorf("CreateChannel failed: %v", err)
		}
		logrus.Infof("New Channel Created: %v", createdChannel)
		ch = createdChannel
	}
	logrus.Infof("Channel: %v", ch)
}

func (s *Slack) handleMessage(msg *slack.MessageEvent) error {
	switch msg.SubType {
	case typeMessage:
		text := fmt.Sprintf("%v\n", msg.Msg.Text)
		logrus.Infof("Message text: %v", text)
		standupText, messageIsStandup := s.isStandup(text)
		if messageIsStandup {
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
			standupText, messageIsStandup := s.isStandup(msg.SubMessage.Text)
			if messageIsStandup {
				_, err := s.db.CreateStandup(model.Standup{
					ChannelID: msg.Channel,
					UserID:    msg.SubMessage.User,
					Comment:   standupText,
					MessageTS: msg.SubMessage.Timestamp,
				})
				if err != nil {
					logrus.Errorf("slack: CreateStandup failed: %v\n", err)
					return err
				}
				item := slack.ItemRef{msg.Channel, msg.SubMessage.Timestamp, "", ""}
				s.api.AddReaction("heavy_check_mark", item)
				return nil
			}
		}
		text, messageIsStandup := s.isStandup(msg.SubMessage.Text)
		if messageIsStandup {
			standup.Comment = text
			_, err := s.db.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("slack: UpdateStandup failed: %v\n", err)
				return err
			}
			item := slack.ItemRef{msg.Channel, msg.SubMessage.Timestamp, "", ""}
			s.api.RemoveReaction("exploding_head", item)
			s.api.AddReaction("heavy_check_mark", item)
		} else {
			item := slack.ItemRef{msg.Channel, msg.SubMessage.Timestamp, "", ""}
			s.api.RemoveReaction("heavy_check_mark", item)
			s.api.AddReaction("exploding_head", item)
		}
	}
	return nil
}

func (s *Slack) isStandup(message string) (string, bool) {

	mentionsProblem := false
	problemKeys := []string{s.Conf.Translate.P1, s.Conf.Translate.P2, s.Conf.Translate.P3, s.Conf.Translate.P4}
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
	logrus.Infof("UpdateUsersList start")
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

//FillStandupsForNonReporters fills standup entries with empty standups to later recognize
//non reporters vs those who did not have to write standups
func (s *Slack) FillStandupsForNonReporters() {
	logrus.Println("FillStandupsForNonReporters!")
	//check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	allUsers, err := s.db.ListAllChannelMembers()
	logrus.Infof("List all channel members: %v", allUsers)
	if err != nil {
		logrus.Errorf("notifier: s.GetCurrentDayNonReporters failed: %v\n", err)
		return
	}
	for _, user := range allUsers {

		if user.Created.Day() == time.Now().Day() {
			logrus.Infof("User %v, was created today. Skip!", user)
			continue
		}
		hasStandup := s.db.SubmittedStandupToday(user.UserID, user.ChannelID)
		logrus.Infof("User: %v hasStandup: %v", user.UserID, hasStandup)
		if !hasStandup {
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
