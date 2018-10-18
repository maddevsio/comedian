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
	typeMessage       = ""
	typeEditMessage   = "message_changed"
	typeDeleteMessage = "message_deleted"
)

// Slack struct used for storing and communicating with slack api
type Slack struct {
	API  *slack.Client
	RTM  *slack.RTM
	WG   sync.WaitGroup
	DB   *storage.MySQL
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
	s.API = slack.New(conf.SlackToken)
	s.RTM = s.API.NewRTM()
	s.DB = m
	return s, nil
}

// Run runs a listener loop for slack
func (s *Slack) Run() {

	fmt.Printf("Team monitoring enabled: %v\n", s.Conf.TeamMonitoringEnabled)
	s.UpdateUsersList()
	gocron.Every(1).Day().At("23:50").Do(s.FillStandupsForNonReporters)
	gocron.Every(1).Day().At("23:55").Do(s.UpdateUsersList)
	gocron.Start()

	s.WG.Add(1)
	go s.RTM.ManageConnection()
	s.WG.Done()

	for msg := range s.RTM.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			logrus.Info("Reconnected!")
			s.SendUserMessage(s.Conf.ManagerSlackUserID, s.Conf.Translate.HelloManager)
		case *slack.MessageEvent:
			botUserID := fmt.Sprintf("<@%s>", s.RTM.GetInfo().User.ID)
			s.handleMessage(ev, botUserID)
		case *slack.MemberJoinedChannelEvent:
			s.handleJoin(ev.Channel)
		case *slack.MemberLeftChannelEvent:
			logrus.Infof("type: %v, user: %v, channel: %v, channelType: %v, team: %v", ev.Type, ev.User, ev.Channel, ev.ChannelType, ev.Team)
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
	ch, err := s.DB.SelectChannel(channelID)
	if err != nil {
		logrus.Error("No such channel found! Will create one!")
		channel, err := s.API.GetConversationInfo(channelID, true)
		if err != nil {
			logrus.Errorf("GetConversationInfo failed: %v", err)
		}
		createdChannel, err := s.DB.CreateChannel(model.Channel{
			ChannelName: channel.Name,
			ChannelID:   channel.ID,
			StandupTime: int64(0),
		})
		if err != nil {
			logrus.Errorf("CreateChannel failed: %v", err)
			return
		}
		logrus.Infof("New Channel Created: %v", createdChannel)
		ch = createdChannel
	}
	logrus.Infof("Channel: %v", ch)
}

func (s *Slack) handleMessage(msg *slack.MessageEvent, botUserID string) error {
	switch msg.SubType {
	case typeMessage:
		if !strings.Contains(msg.Msg.Text, botUserID) && !strings.Contains(msg.Msg.Text, "#standup") && !strings.Contains(msg.Msg.Text, "#стэндап") {
			return nil
		}
		messageIsStandup, problem := s.analizeStandup(msg.Msg.Text)
		if problem != "" {
			return s.SendEphemeralMessage(msg.Channel, msg.User, problem)
		}
		if messageIsStandup {
			if s.DB.SubmittedStandupToday(msg.User, msg.Channel) {
				return s.SendEphemeralMessage(msg.Channel, msg.User, s.Conf.Translate.StandupHandleOneDayOneStandup)
			}
			standup, err := s.DB.CreateStandup(model.Standup{
				ChannelID: msg.Channel,
				UserID:    msg.User,
				Comment:   msg.Msg.Text,
				MessageTS: msg.Msg.Timestamp,
			})
			if err != nil {
				logrus.Errorf("CreateStandup failed: %v", err)
				errorReportToManager := fmt.Sprintf("I could not save standup for user %s in channel %s because of the following reasons: %v", msg.User, msg.Channel, err)
				s.SendUserMessage(s.Conf.ManagerSlackUserID, errorReportToManager)
				return s.SendEphemeralMessage(msg.Channel, msg.User, s.Conf.Translate.StandupHandleCouldNotSaveStandup)
			}
			logrus.Infof("Standup created #id:%v\n", standup.ID)
			item := slack.ItemRef{msg.Channel, msg.Msg.Timestamp, "", ""}
			time.Sleep(2 * time.Second)
			s.API.AddReaction("heavy_check_mark", item)
			return s.SendEphemeralMessage(msg.Channel, msg.User, s.Conf.Translate.StandupHandleCreatedStandup)
		}
	case typeEditMessage:
		if !strings.Contains(msg.SubMessage.Text, botUserID) && !strings.Contains(msg.SubMessage.Text, "#standup") && !strings.Contains(msg.SubMessage.Text, "#стэндап") {
			return nil
		}
		standup, err := s.DB.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			messageIsStandup, problem := s.analizeStandup(msg.SubMessage.Text)
			if problem != "" {
				return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
			}
			if messageIsStandup {
				if s.DB.SubmittedStandupToday(msg.SubMessage.User, msg.Channel) {
					return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, s.Conf.Translate.StandupHandleOneDayOneStandup)
				}
				standup, err := s.DB.CreateStandup(model.Standup{
					ChannelID: msg.Channel,
					UserID:    msg.SubMessage.User,
					Comment:   msg.SubMessage.Text,
					MessageTS: msg.SubMessage.Timestamp,
				})
				if err != nil {
					logrus.Errorf("CreateStandup while updating text failed: %v", err)
					errorReportToManager := fmt.Sprintf("I could not create standup while updating msg for user %s in channel %s because of the following reasons: %v", msg.SubMessage.User, msg.Channel, err)
					s.SendUserMessage(s.Conf.ManagerSlackUserID, errorReportToManager)
					return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, s.Conf.Translate.StandupHandleCouldNotSaveStandup)
				}
				logrus.Infof("Standup created #id:%v\n", standup.ID)
				item := slack.ItemRef{msg.Channel, msg.SubMessage.Timestamp, "", ""}
				time.Sleep(2 * time.Second)
				s.API.AddReaction("heavy_check_mark", item)
				return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, s.Conf.Translate.StandupHandleCreatedStandup)
			}
		}

		messageIsStandup, problem := s.analizeStandup(msg.SubMessage.Text)
		if problem != "" {
			return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
		}
		if messageIsStandup {
			standup.Comment = msg.SubMessage.Text
			_, err := s.DB.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("UpdateStandup failed: %v", err)
				errorReportToManager := fmt.Sprintf("I could not update standup for user %s in channel %s because of the following reasons: %v", msg.SubMessage.User, msg.Channel, err)
				s.SendUserMessage(s.Conf.ManagerSlackUserID, errorReportToManager)
				return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, s.Conf.Translate.StandupHandleCouldNotSaveStandup)
			}
			logrus.Infof("Standup updated #id:%v\n", standup.ID)
			time.Sleep(2 * time.Second)
			return s.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, s.Conf.Translate.StandupHandleUpdatedStandup)
		}

	case typeDeleteMessage:
		standup, err := s.DB.SelectStandupByMessageTS(msg.DeletedTimestamp)
		if err != nil {
			logrus.Errorf("SelectStandupByMessageTS failed: %v", err)
			return nil
		}
		s.DB.DeleteStandup(standup.ID)
		logrus.Infof("Standup deleted #id:%v\n", standup.ID)
	}
	return nil
}

func (s *Slack) analizeStandup(message string) (bool, string) {
	message = strings.ToLower(message)
	mentionsProblem := false
	problemKeys := []string{"problem", "difficult", "stuck", "question", "issue", "проблем", "трудност", "затрдуднени", "вопрос"}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}
	if !mentionsProblem {
		return false, s.Conf.Translate.StandupHandleNoProblemsMentioned
	}

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{"yesterday", "friday", "completed", "вчера", "пятниц", "делал", "сделано"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}
	if !mentionsYesterdayWork {
		return false, s.Conf.Translate.StandupHandleNoYesterdayWorkMentioned
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{"today", "going", "plan", "сегодня", "собираюсь", "план"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}
	if !mentionsTodayPlans {
		return false, s.Conf.Translate.StandupHandleNoTodayPlansMentioned
	}
	return true, ""
}

// SendMessage posts a message in a specified channel visible for everyone
func (s *Slack) SendMessage(channel, message string) error {
	_, _, err := s.API.PostMessage(channel, message, slack.PostMessageParameters{})
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

func (s *Slack) SendReportMessage(channel, message string, attachments []slack.Attachment) error {
	_, _, err := s.API.PostMessage(channel, message, slack.PostMessageParameters{
		Attachments: attachments,
	})
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

// SendEphemeralMessage posts a message in a specified channel which is visible only for selected user
func (s *Slack) SendEphemeralMessage(channel, user, message string) error {
	_, err := s.API.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(message, true),
	)
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

// SendUserMessage Direct Message specific user
func (s *Slack) SendUserMessage(userID, message string) error {
	_, _, channelID, err := s.API.OpenIMChannel(userID)
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
	users, err := s.API.GetUsers()
	if err != nil {
		logrus.Errorf("GetUsers failed: %v", err)
	}
	for _, user := range users {
		if user.IsBot || user.Name == "slackbot" {
			continue
		}

		u, err := s.DB.SelectUser(user.ID)
		if err != nil {
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				s.DB.CreateUser(model.User{
					UserName: user.Name,
					UserID:   user.ID,
					Role:     "admin",
				})
				continue
			}
			s.DB.CreateUser(model.User{
				UserName: user.Name,
				UserID:   user.ID,
				Role:     "",
			})
			continue
		}
		if user.Deleted {
			s.DB.DeleteUser(u.ID)
			cm, err := s.DB.FindMembersByUserID(u.UserID)
			if err != nil {
				continue
			}
			for _, member := range cm {
				s.DB.DeleteChannelMember(member.UserID, member.ChannelID)
				tt, err := s.DB.SelectTimeTable(member.ID)
				if err != nil {
					continue
				}
				s.DB.DeleteTimeTable(tt.ID)
			}
		}
	}
}

//FillStandupsForNonReporters fills standup entries with empty standups to later recognize
//non reporters vs those who did not have to write standups
func (s *Slack) FillStandupsForNonReporters() {
	logrus.Println("FillStandupsForNonReporters!")
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		return
	}
	allUsers, err := s.DB.ListAllChannelMembers()
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
		hasStandup := s.DB.SubmittedStandupToday(user.UserID, user.ChannelID)
		logrus.Infof("User: %v hasStandup: %v", user.UserID, hasStandup)
		if !hasStandup {
			standup, err := s.DB.CreateStandup(model.Standup{
				ChannelID: user.ChannelID,
				UserID:    user.UserID,
				Comment:   "",
				MessageTS: strconv.Itoa(int(time.Now().Unix())),
			})
			if err != nil {
				logrus.Errorf("notifier: CreateStandup failed: %v\n", err)
				errorReportToManager := fmt.Sprintf("I could not create empty standup for user %s in channel %s because of the following reasons: %v", user.UserID, user.ChannelID, err)
				s.SendUserMessage(s.Conf.ManagerSlackUserID, errorReportToManager)
				return
			}
			logrus.Infof("notifier: Empty Standup created: %v\n", standup.ID)
		}
	}
}
