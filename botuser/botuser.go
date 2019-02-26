package botuser

import (
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"gitlab.com/team-monitoring/comedian/utils"
)

var (
	typeMessage       = ""
	typeEditMessage   = "message_changed"
	typeDeleteMessage = "message_deleted"
)

// Bot struct used for storing and communicating with slack api
type Bot struct {
	API        *slack.Client
	Properties model.ControlPannel
	DB         *storage.MySQL
	Bundle     *i18n.Bundle
}

func (bot *Bot) HandleMessage(msg *slack.MessageEvent, botUserID string) {

	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
	oneStandupPerDay := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "OneStandupPerDay",
			Description: "Warning that only one standup per day is allowed",
			Other:       "You can submit only one standup per day. Please, edit today's standup or submit your next standup tomorrow!",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	couldNotSaveStandup := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "CouldNotSaveStandup",
			Description: "Displays a message when unexpected errors occur",
			Other:       "Something went wrong and I could not save your standup in database. Please, report this to your PM.",
		},
		TemplateData: map[string]string{
			"ID": msg.User,
		},
	})

	errorReportToManager := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrorReportToManager",
			Other: "I could not save standup for user <@{{.user}}> in channel <#{{.channel}}> because of the following reasons: %v",
		},
		TemplateData: map[string]string{
			"user":    msg.User,
			"channel": msg.Channel,
		},
	})

	switch msg.SubType {
	case typeMessage:

		if !strings.Contains(msg.Msg.Text, botUserID) && !strings.Contains(msg.Msg.Text, "#standup") {
			return
		}

		messageIsStandup, problem := bot.analizeStandup(msg.Msg.Text)
		if problem != "" {
			bot.SendEphemeralMessage(msg.Channel, msg.User, problem)
			return
		}
		if messageIsStandup {
			if bot.DB.SubmittedStandupToday(msg.User, msg.Channel) {
				bot.SendEphemeralMessage(msg.Channel, msg.User, oneStandupPerDay)
				return
			}
			standup, err := bot.DB.CreateStandup(model.Standup{
				TeamID:    msg.Team,
				ChannelID: msg.Channel,
				UserID:    msg.User,
				Comment:   msg.Msg.Text,
				MessageTS: msg.Msg.Timestamp,
			})
			if err != nil {
				logrus.Errorf("CreateStandup failed: %v", err)
				bot.SendUserMessage(bot.Properties.ManagerSlackUserID, fmt.Sprintf(errorReportToManager, err))
				bot.SendEphemeralMessage(msg.Channel, msg.User, couldNotSaveStandup)
				return
			}
			logrus.Infof("Standup created #id:%v\n", standup.ID)
			item := slack.ItemRef{
				Channel:   msg.Channel,
				Timestamp: msg.Msg.Timestamp,
				File:      "",
				Comment:   "",
			}
			bot.API.AddReaction("heavy_check_mark", item)
			return
		}
	case typeEditMessage:

		if !strings.Contains(msg.SubMessage.Text, botUserID) && !strings.Contains(msg.SubMessage.Text, "#standup") {
			return
		}

		standup, err := bot.DB.SelectStandupByMessageTS(msg.SubMessage.Timestamp)
		if err != nil {
			messageIsStandup, problem := bot.analizeStandup(msg.SubMessage.Text)
			if problem != "" {
				bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
				return
			}
			if messageIsStandup {
				if bot.DB.SubmittedStandupToday(msg.SubMessage.User, msg.Channel) {
					bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, oneStandupPerDay)
					return
				}
				logrus.Infof("CreateStandup while updating text ChannelID (%v), UserID (%v), Comment (%v), TimeStamp (%v)", msg.Channel, msg.SubMessage.User, msg.SubMessage.Text, msg.SubMessage.Timestamp)
				standup, err := bot.DB.CreateStandup(model.Standup{
					TeamID:    msg.Team,
					ChannelID: msg.Channel,
					UserID:    msg.SubMessage.User,
					Comment:   msg.SubMessage.Text,
					MessageTS: msg.SubMessage.Timestamp,
				})
				if err != nil {
					logrus.Errorf("CreateStandup while updating text failed: %v", err)
					bot.SendUserMessage(bot.Properties.ManagerSlackUserID, errorReportToManager)
					bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, couldNotSaveStandup)
					return
				}
				logrus.Infof("Standup created #id:%v\n", standup.ID)
				item := slack.ItemRef{
					Channel:   msg.Channel,
					Timestamp: msg.SubMessage.Timestamp,
					File:      "",
					Comment:   "",
				}
				bot.API.AddReaction("heavy_check_mark", item)
				return
			}
		}

		messageIsStandup, problem := bot.analizeStandup(msg.SubMessage.Text)
		if problem != "" {
			bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, problem)
			return
		}
		if messageIsStandup {
			standup.Comment = msg.SubMessage.Text
			st, err := bot.DB.UpdateStandup(standup)
			if err != nil {
				logrus.Errorf("UpdateStandup failed: %v", err)
				bot.SendEphemeralMessage(msg.Channel, msg.SubMessage.User, couldNotSaveStandup)
				return
			}
			logrus.Infof("Standup updated #id:%v\n", st.ID)
			return
		}

	case typeDeleteMessage:
		standup, err := bot.DB.SelectStandupByMessageTS(msg.DeletedTimestamp)
		if err != nil {
			logrus.Errorf("SelectStandupByMessageTS failed: %v", err)
			return
		}
		err = bot.DB.DeleteStandup(standup.ID)
		if err != nil {
			logrus.Errorf("DeleteStandup failed: %v", err)
			return
		}
		logrus.Infof("Standup deleted #id:%v\n", standup.ID)
	}
}

func (bot *Bot) analizeStandup(message string) (bool, string) {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
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

	problemKeys := []string{"problem", "difficult", "stuck", "question", "issue", "block", "проблем", "трудност", "затруднени", "вопрос", "блок"}
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
func (bot *Bot) SendMessage(channel, message string, attachments []slack.Attachment) error {
	_, _, err := bot.API.PostMessage(channel, message, slack.PostMessageParameters{
		Attachments: attachments,
	})
	if err != nil {
		logrus.Errorf("slack: PostMessage failed: %v\n", err)
		return err
	}
	return err
}

// SendEphemeralMessage posts a message in a specified channel which is visible only for selected user
func (bot *Bot) SendEphemeralMessage(channel, user, message string) error {
	_, err := bot.API.PostEphemeral(
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
func (bot *Bot) SendUserMessage(userID, message string) error {
	_, _, channelID, err := bot.API.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	err = bot.SendMessage(channelID, message, nil)
	if err != nil {
		return err
	}
	return err
}

//HandleJoin handles comedian joining channel
func (bot *Bot) HandleJoin(channelID, teamID string) {
	_, err := bot.DB.SelectChannel(channelID)
	if err == nil {
		return
	}

	logrus.Error("No such channel found! Will create one!")
	channel, err := bot.API.GetConversationInfo(channelID, true)
	if err != nil {
		logrus.Errorf("GetConversationInfo failed: %v", err)
	}
	createdChannel, err := bot.DB.CreateChannel(model.Channel{
		TeamID:      teamID,
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

func (bot *Bot) ImplementCommands(form model.FullSlackForm) string {

	accessLevel, err := bot.getAccessLevel(form.UserID, form.ChannelID)
	if err != nil {
		return err.Error()
	}

	command, params := utils.CommandParsing(form.Text)

	switch command {
	case "add":
		return bot.addCommand(accessLevel, form.ChannelID, params)
	case "show":
		return bot.showCommand(form.ChannelID, params)
	case "remove":
		return bot.deleteCommand(accessLevel, form.ChannelID, params)
	case "add_deadline":
		return bot.addTime(accessLevel, form.ChannelID, params)
	case "remove_deadline":
		return bot.removeTime(accessLevel, form.ChannelID)
	case "show_deadline":
		return bot.showTime(form.ChannelID)
	default:
		return bot.DisplayHelpText("")
	}
}

func (bot *Bot) getAccessLevel(userID, channelID string) (int, error) {
	user, err := bot.DB.SelectUser(userID)
	if err != nil {
		return 0, err
	}
	if user.IsAdmin() {
		return 2, nil
	}
	if bot.DB.UserIsPMForProject(userID, channelID) {
		return 3, nil
	}
	return 4, nil
}

//UpdateUsersList updates users in workspace
func (bot *Bot) UpdateUsersList() {
	users, err := bot.API.GetUsers()
	if err != nil {
		logrus.Errorf("GetUsers failed: %v", err)
		return
	}
	for _, user := range users {
		if user.IsBot || user.Name == "slackbot" {
			continue
		}

		u, err := bot.DB.SelectUser(user.ID)
		if err != nil && !user.Deleted {
			if user.IsAdmin || user.IsOwner || user.IsPrimaryOwner {
				u, err = bot.DB.CreateUser(model.User{
					TeamID:   user.TeamID,
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
			u, err = bot.DB.CreateUser(model.User{
				TeamID:   user.TeamID,
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
			u.TeamID = user.TeamID
			_, err = bot.DB.UpdateUser(u)
			if err != nil {
				logrus.Errorf("Update User failed %v", err)
				continue
			}
		}

		if user.Deleted {
			bot.DB.DeleteUser(u.ID)
			cm, err := bot.DB.FindMembersByUserID(u.UserID)
			if err != nil {
				continue
			}
			for _, member := range cm {
				bot.DB.DeleteChannelMember(member.UserID, member.ChannelID)
			}
		}
	}
	logrus.Info("Users list updated successfully")
}
