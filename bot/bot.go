package bot

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jasonlvhit/gocron"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"

	"strings"
)

var (
	typeMessage       = ""
	typeEditMessage   = "message_changed"
	typeDeleteMessage = "message_deleted"
)

// Bot struct used for storing and communicating with slack api
type Bot struct {
	API        *slack.Client
	WG         sync.WaitGroup
	DB         *storage.MySQL
	Conf       config.Config
	CP         *model.ControllPannel
	Bundle     *i18n.Bundle
	TeamDomain string
}

// NewBot creates a new copy of bot handler
func NewBot(conf config.Config) (*Bot, error) {
	db, err := storage.NewMySQL(conf)
	if err != nil {
		logrus.Errorf("slack: NewMySQL failed: %v\n", err)
		return nil, err
	}

	cp, err := db.GetControllPannel()
	if err != nil {
		logrus.Errorf("slack: GetControllPannel failed: %v\n", err)
		cp, err = db.CreateControllPannel()
		if err != nil {
			return nil, err
		}
	}

	b := &Bot{}
	b.Conf = conf
	b.API = slack.New(conf.SlackToken)
	b.DB = db
	b.CP = &cp

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("bot/active.en.toml")
	bundle.MustLoadMessageFile("bot/active.ru.toml")

	b.Bundle = bundle

	return b, nil
}

// Run runs a listener loop for slack
func (b *Bot) Run() {

	b.UpdateUsersList()
	team, err := b.API.GetTeamInfo()
	if err != nil {
		logrus.Error(err)
	}
	if b.TeamDomain == "" {
		b.TeamDomain = team.Domain
	}

	localizer := i18n.NewLocalizer(b.Bundle, b.CP.Language)
	helloManager := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: "HelloManager",
		},
		TemplateData: map[string]string{
			"ID": b.CP.ManagerSlackUserID,
		},
	})

	b.SendUserMessage(b.CP.ManagerSlackUserID, helloManager)

	gocron.Every(1).Day().At("23:59").Do(b.FillStandupsForNonReporters)
	gocron.Every(1).Day().At("23:55").Do(b.UpdateUsersList)
	gocron.Start()

	rtm := b.API.NewRTM()

	b.WG.Add(1)
	go rtm.ManageConnection()
	b.WG.Done()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			botUserID := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
			b.handleMessage(ev, botUserID)
		case *slack.MemberJoinedChannelEvent:
			b.handleJoin(ev.Channel)
		case *slack.MemberLeftChannelEvent:
			b.handleLeft(ev.Channel, ev.User)
		case *slack.ChannelLeftEvent:
			b.handleBotRemovedFromChannel(ev.Channel)
		case *slack.InvalidAuthEvent:
			logrus.Error("Invalid Auth!")
			return
		case *slack.ConnectedEvent:
			logrus.Info("Reconnected!")
		}
	}
}

func (b *Bot) handleLeft(ChannelID, UserID string) {
	logrus.Infof("Member %v left channel %v", UserID, ChannelID)
	channelMember, err := b.DB.FindChannelMemberByUserID(UserID, ChannelID)
	if err != nil {
		logrus.Error("slack:handleLeft FindChannelMemberByUserID failed: ", err)
		return
	}
	timetable, err := b.DB.SelectTimeTable(channelMember.ID)
	if err != nil {
		logrus.Error("slack:handleLeft SelectTimeTable failed: ", err)
	}
	err = b.DB.DeleteTimeTable(timetable.ID)
	if err != nil {
		logrus.Error("slack:handleLeft DeleteTimeTable failed: ", err)
	}
	err = b.DB.DeleteChannelMember(UserID, ChannelID)
	if err != nil {
		logrus.Error("slack:handleLeft DeleteChannelMember failed: ", err)
	}
}

func (b *Bot) handleBotRemovedFromChannel(ChannelID string) {
	logrus.Infof("Bot removed from %v channel", ChannelID)
	channelMembers, err := b.DB.ListChannelMembers(ChannelID)
	if err != nil {
		logrus.Error("slack: ListChannelMembers failed: ", err)
		return
	}
	for _, chanMemb := range channelMembers {
		timetable, err := b.DB.SelectTimeTable(chanMemb.ID)
		if err != nil {
			logrus.Error("slack: SelectTimeTable failed: ", err)
		}
		err = b.DB.DeleteTimeTable(timetable.ID)
		if err != nil {
			logrus.Error("slack: DeleteTimeTable failed: ", err)
		}
		err = b.DB.DeleteChannelMember(chanMemb.UserID, chanMemb.ChannelID)
		if err != nil {
			logrus.Error("slack: DeleteChannelMember failed: ", err)
		}
	}
	err = b.DB.DeleteStandupTime(ChannelID)
	if err != nil {
		logrus.Error("slack: DeleteStandupTime failed: ", err)
	}
}

func (b *Bot) handleJoin(channelID string) {
	_, err := b.DB.SelectChannel(channelID)
	if err != nil {
		logrus.Error("No such channel found! Will create one!")
		channel, err := b.API.GetConversationInfo(channelID, true)
		if err != nil {
			logrus.Errorf("GetConversationInfo failed: %v", err)
		}
		createdChannel, err := b.DB.CreateChannel(model.Channel{
			ChannelName: channel.Name,
			ChannelID:   channel.ID,
			StandupTime: int64(0),
		})
		if err != nil {
			logrus.Errorf("CreateChannel failed: %v", err)
			return
		}
		logrus.Infof("New Channel Created: %v", createdChannel)
	}
}

func (b *Bot) handleMessage(msg *slack.MessageEvent, botUserID string) {
	localizer := i18n.NewLocalizer(b.Bundle, b.CP.Language)
	oneStandupPerDay := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "OneStandupPerDay",
			Description: "Warning that only one standup per day is allowed",
			Other:       "<@{{.ID}}>, you can submit only one standup per day. Please, edit today's standup or submit your next standup tomorrow!",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	couldNotSaveStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "CouldNotSaveStandup",
			Other: "<@{{.ID}}>, something went wrong and I could not save your standup in database. Please, report this to your PM.",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	errorReportToManager := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrorReportToManager",
			Other: "I could not save standup for user {{.user}} in channel {{.channel}} because of the following reasons: %v",
		},
		TemplateData: map[string]string{
			"user":    msg.User,
			"channel": msg.Channel,
		},
	})

	switch msg.SubType {
	case typeMessage:

		if strings.Contains(msg.Msg.Text, "#bug") {
			b.recordBug(msg.Channel, msg.Msg.User, msg.Msg.Text)
		}

		if !strings.Contains(msg.Msg.Text, botUserID) && !strings.Contains(msg.Msg.Text, "#standup") {
			return
		}

		messageIsStandup, problem := b.analizeStandup(msg.Msg.Text)
		if problem != "" {
			b.SendEphemeralMessage(msg.Channel, msg.User, problem)
			return
		}
		if messageIsStandup {
			if b.DB.SubmittedStandupToday(msg.User, msg.Channel) {
				b.SendEphemeralMessage(msg.Channel, msg.User, oneStandupPerDay)
				return
			}
			standup, err := b.DB.CreateStandup(model.Standup{
				ChannelID: msg.Channel,
				UserID:    msg.User,
				Comment:   msg.Msg.Text,
				MessageTS: msg.Msg.Timestamp,
			})
			if err != nil {
				logrus.Errorf("CreateStandup failed: %v", err)
				b.SendUserMessage(b.CP.ManagerSlackUserID, errorReportToManager)
				b.SendEphemeralMessage(msg.Channel, msg.User, couldNotSaveStandup)
				return
			}
			logrus.Infof("Standup created #id:%v\n", standup.ID)
			item := slack.ItemRef{
				Channel:   msg.Channel,
				Timestamp: msg.Msg.Timestamp,
				File:      "",
				Comment:   "",
			}
			b.API.AddReaction("heavy_check_mark", item)
			return
		}
	case typeEditMessage:
		if strings.Contains(msg.SubMessage.Text, "#bug") {
			b.recordBug(msg.Channel, msg.SubMessage.User, msg.SubMessage.Text)
		}

		if !strings.Contains(msg.SubMessage.Text, botUserID) && !strings.Contains(msg.SubMessage.Text, "#standup") {
			return
		}
		standup, err := b.DB.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			messageIsStandup, problem := b.analizeStandup(msg.SubMessage.Text)
			if problem != "" {
				b.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
				return
			}
			if messageIsStandup {
				if b.DB.SubmittedStandupToday(msg.SubMessage.User, msg.Channel) {
					b.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, oneStandupPerDay)
					return
				}
				logrus.Infof("CreateStandup while updating text ChannelID (%v), UserID (%v), Comment (%v), TimeStamp (%v)", msg.Channel, msg.SubMessage.User, msg.SubMessage.Text, msg.SubMessage.Timestamp)
				standup, err := b.DB.CreateStandup(model.Standup{
					ChannelID: msg.Channel,
					UserID:    msg.SubMessage.User,
					Comment:   msg.SubMessage.Text,
					MessageTS: msg.SubMessage.Timestamp,
				})
				if err != nil {
					logrus.Errorf("CreateStandup while updating text failed: %v", err)
					b.SendUserMessage(b.CP.ManagerSlackUserID, errorReportToManager)
					b.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, couldNotSaveStandup)
					return
				}
				logrus.Infof("Standup created #id:%v\n", standup.ID)
				item := slack.ItemRef{
					Channel:   msg.Channel,
					Timestamp: msg.SubMessage.Timestamp,
					File:      "",
					Comment:   "",
				}
				b.API.AddReaction("heavy_check_mark", item)
				return
			}
		}

		messageIsStandup, problem := b.analizeStandup(msg.SubMessage.Text)
		if problem != "" {
			b.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
			return
		}
		if messageIsStandup {
			standup.Comment = msg.SubMessage.Text
			st, err := b.DB.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("UpdateStandup failed: %v", err)
				b.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, couldNotSaveStandup)
				return
			}
			logrus.Infof("Standup updated #id:%v\n", st.ID)
			return
		}

	case typeDeleteMessage:
		standup, err := b.DB.SelectStandupByMessageTS(msg.DeletedTimestamp)
		if err != nil {
			logrus.Errorf("SelectStandupByMessageTS failed: %v", err)
			return
		}
		err = b.DB.DeleteStandup(standup.ID)
		if err != nil {
			logrus.Errorf("DeleteStandup failed: %v", err)
			return
		}
		logrus.Infof("Standup deleted #id:%v\n", standup.ID)
	}
}

func (b *Bot) analizeStandup(message string) (bool, string) {
	localizer := i18n.NewLocalizer(b.Bundle, b.CP.Language)
	message = strings.ToLower(message)

	mentionsYesterdayWork := false
	yesterdayWorkKeys := []string{"yesterday", "friday", "monday", "tuesday", "wednesday", "thursday", "saturday", "sunday", "completed", "вчера", "пятниц", "делал", "сделано", "понедельник", "вторник", "сред", "четверг", "суббот", "воскресенье"}
	for _, work := range yesterdayWorkKeys {
		if strings.Contains(message, work) {
			mentionsYesterdayWork = true
		}
	}
	if !mentionsYesterdayWork {
		standupHandleNoYesterdayWorkMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoYesterdayWorkMentioned",
				Description: "No 'yesterday' keywords in standup",
				Other:       ":warning: No 'yesterday' related keywords detected! Please, use one of the following: 'yesterday' or weekdays such as 'friday' etc.",
			},
		})
		return false, standupHandleNoYesterdayWorkMentioned
	}

	mentionsTodayPlans := false
	todayPlansKeys := []string{"today", "going", "plan", "сегодня", "собираюсь", "план"}
	for _, plan := range todayPlansKeys {
		if strings.Contains(message, plan) {
			mentionsTodayPlans = true
		}
	}
	if !mentionsTodayPlans {
		standupHandleNoTodayPlansMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoTodayPlansMentioned",
				Description: "No 'today' keywords in standup",
				Other:       ":warning: No 'today' related keywords detected! Please, use one of the following: 'today', 'going', 'plan'",
			},
		})
		return false, standupHandleNoTodayPlansMentioned
	}

	mentionsProblem := false
	problemKeys := []string{"problem", "difficult", "stuck", "question", "issue", "block", "проблем", "трудност", "затрдуднени", "вопрос"}
	for _, problem := range problemKeys {
		if strings.Contains(message, problem) {
			mentionsProblem = true
		}
	}
	if !mentionsProblem {
		standupHandleNoProblemsMentioned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "StandupHandleNoProblemsMentioned",
				Description: "No 'problems' key in standup",
				Other:       ":warning: No 'problems' related keywords detected! Please, use one of the following: 'problem', 'difficult', 'stuck', 'question', 'issue'",
			},
		})
		return false, standupHandleNoProblemsMentioned
	}

	return true, ""
}

// SendMessage posts a message in a specified channel visible for everyone
func (b *Bot) SendMessage(channel, message string, attachments []slack.Attachment) error {
	_, _, err := b.API.PostMessage(channel, message, slack.PostMessageParameters{
		Attachments: attachments,
	})
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

// SendEphemeralMessage posts a message in a specified channel which is visible only for selected user
func (b *Bot) SendEphemeralMessage(channel, user, message string) error {
	_, err := b.API.PostEphemeral(
		channel,
		user,
		slack.MsgOptionText(message, true),
	)
	if err != nil {
		logrus.Errorf("slack: PostEphemeral failed: %v\n", err)
		return err
	}
	return err
}

// SendUserMessage Direct Message specific user
func (b *Bot) SendUserMessage(userID, message string) error {
	_, _, channelID, err := b.API.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	err = b.SendMessage(channelID, message, nil)
	if err != nil {
		return err
	}
	return err
}

//UpdateUsersList updates users in workspace
func (b *Bot) UpdateUsersList() {
	users, err := b.API.GetUsers()
	if err != nil {
		logrus.Errorf("GetUsers failed: %v", err)
		return
	}
	for _, user := range users {
		if user.IsBot || user.Name == "slackbot" {
			continue
		}

		u, err := b.DB.SelectUser(user.ID)
		if err != nil && !user.Deleted {
			logrus.Errorf("SelectUser with ID [%v] failed %v", user.ID, err)
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				u, err = b.DB.CreateUser(model.User{
					UserName: user.Name,
					UserID:   user.ID,
					Role:     "admin",
					RealName: user.RealName,
				})
				if err != nil {
					logrus.Errorf("CreateUser failed %v", err)
					continue
				}
				continue
			}
			u, err = b.DB.CreateUser(model.User{
				UserName: user.Name,
				UserID:   user.ID,
				Role:     "",
				RealName: user.RealName,
			})
			if err != nil {
				logrus.Errorf("CreateUser with no role failed %v", err)
				continue
			}
		}
		if !user.Deleted {
			u.UserName = user.Name
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				u.Role = "admin"
			}
			u.RealName = user.RealName
			_, err = b.DB.UpdateUser(u)
			if err != nil {
				logrus.Errorf("Update User failed %v", err)
				continue
			}
		}

		if user.Deleted {
			b.DB.DeleteUser(u.ID)
			cm, err := b.DB.FindMembersByUserID(u.UserID)
			if err != nil {
				continue
			}
			for _, member := range cm {
				b.DB.DeleteChannelMember(member.UserID, member.ChannelID)
				tt, err := b.DB.SelectTimeTable(member.ID)
				if err != nil {
					continue
				}
				b.DB.DeleteTimeTable(tt.ID)
			}
		}
	}
	logrus.Info("Users list updated successfully")
}

//FillStandupsForNonReporters fills standup entries with empty standups to later recognize
//non reporters vs those who did not have to write standups
func (b *Bot) FillStandupsForNonReporters() {
	logrus.Info("FillStandupsForNonReporters begin")
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		logrus.Info("Weekends! Do not check!")
		return
	}
	allUsers, err := b.DB.ListAllChannelMembers()
	if err != nil {
		logrus.Errorf("ListAllChannelMembers while FillStandupsForNonReporters failed: %v", err)
		return
	}
	for _, user := range allUsers {
		hasStandup := b.DB.SubmittedStandupToday(user.UserID, user.ChannelID)
		shouldBeTracked := b.DB.MemberShouldBeTracked(user.ID, time.Now())
		logrus.Infof("User [%v] in [%v] should be tracked [%v] and has standup [%v]", user.UserID, user.ChannelID, shouldBeTracked, hasStandup)
		if !hasStandup && shouldBeTracked {
			_, err := b.DB.CreateStandup(model.Standup{
				ChannelID: user.ChannelID,
				UserID:    user.UserID,
				Comment:   "",
				MessageTS: strconv.Itoa(int(time.Now().Unix())),
			})
			if err != nil {
				logrus.Errorf("Could not create empty standup for user [%v] in [%v]", user.UserID, user.ChannelID)
				continue
			}
			logrus.Infof("Empty standup created for user [%v] in [%v]", user.UserID, user.ChannelID)
		}
	}
}

func (b *Bot) recordBug(channelID, userID, bug string) {
	var text string
	b.SendEphemeralMessage(channelID, userID, "Thank you! Bug Recorded!")
	user, err := b.DB.SelectUser(userID)
	if err != nil {
		logrus.Error(err)
		return
	}
	channel, err := b.API.GetChannelInfo(channelID)

	if err != nil {
		logrus.Error(err)
		text = fmt.Sprintf("%v in %v reported a bug! \n %v", user, channel, bug)
		b.SendUserMessage(b.CP.ManagerSlackUserID, text)
		return
	}

	text = fmt.Sprintf("%v in %v reported a bug! \n %v", user.RealName, channel.Name, bug)
	b.SendUserMessage(b.CP.ManagerSlackUserID, text)

}
