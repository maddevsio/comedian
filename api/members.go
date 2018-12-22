package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (ba *BotAPI) addCommand(accessLevel int, channelID, params string) string {
	var role string
	var members []string
	if strings.Contains(params, "/") {
		dividedText := strings.Split(params, "/")
		role = strings.TrimSpace(dividedText[1])
		members = strings.Fields(dividedText[0])
	} else {
		role = "developer"
		members = strings.Fields(params)
	}

	switch role {
	case "admin", "админ":
		if accessLevel > 2 {
			return ba.Bot.Translate.AccessAtLeastAdmin
		}
		return ba.addAdmins(members)
	case "developer", "разработчик", "":
		if accessLevel > 3 {
			return ba.Bot.Translate.AccessAtLeastPM
		}
		return ba.addMembers(members, "developer", channelID)
	case "pm", "пм":
		if accessLevel > 2 {
			return ba.Bot.Translate.AccessAtLeastAdmin
		}
		return ba.addMembers(members, "pm", channelID)
	default:
		return DisplayHelpText("add")
	}
}

func (ba *BotAPI) showCommand(channelID, params string) string {
	switch params {
	case "admin", "админ":
		return ba.listAdmins()
	case "developer", "разработчик", "":
		return ba.listMembers(channelID, "developer")
	case "pm", "пм":
		return ba.listMembers(channelID, "pm")
	default:
		return DisplayHelpText("show")
	}
}

func (ba *BotAPI) deleteCommand(accessLevel int, channelID, params string) string {
	var role string
	var members []string
	if strings.Contains(params, "/") {
		dividedText := strings.Split(params, "/")
		role = strings.TrimSpace(dividedText[1])
		members = strings.Fields(dividedText[0])
	} else {
		role = "developer"
		members = strings.Fields(params)
	}

	switch role {
	case "admin", "админ":
		if accessLevel > 2 {
			return ba.Bot.Translate.AccessAtLeastAdmin
		}
		return ba.deleteAdmins(members)
	case "developer", "разработчик", "pm", "пм", "":
		if accessLevel > 3 {
			return ba.Bot.Translate.AccessAtLeastPM
		}
		return ba.deleteMembers(members, channelID)
	default:
		return DisplayHelpText("remove")
	}
}

func (ba *BotAPI) addMembers(users []string, role, channel string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			logrus.Errorf("Rest FindChannelMemberByUserID failed: %v", err)
			chanMember, _ := ba.Bot.DB.CreateChannelMember(model.ChannelMember{
				UserID:        userID,
				ChannelID:     channel,
				RoleInChannel: role,
			})
			logrus.Infof("ChannelMember created! ID:%v", chanMember.ID)
		}
		if user.UserID == userID && user.ChannelID == channel {
			exist += u
			continue
		}
		added += u
	}

	if len(failed) != 0 {
		if role == "pm" {
			text += fmt.Sprintf(ba.Bot.Translate.AddPMsFailed, failed)
		} else {
			text += fmt.Sprintf(ba.Bot.Translate.AddMembersFailed, failed)
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			text += fmt.Sprintf(ba.Bot.Translate.AddPMsExist, exist)
		} else {
			text += fmt.Sprintf(ba.Bot.Translate.AddMembersExist, exist)
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			text += fmt.Sprintf(ba.Bot.Translate.AddPMsAdded, added)
		} else {
			text += fmt.Sprintf(ba.Bot.Translate.AddMembersAdded, added)
		}

	}
	return text
}

func (ba *BotAPI) addAdmins(users []string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := ba.Bot.DB.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role == "admin" {
			exist += u
			continue
		}
		user.Role = "admin"
		ba.Bot.DB.UpdateUser(user)
		message := ba.Bot.Translate.PMAssigned
		err = ba.Bot.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(ba.Bot.Translate.AddAdminsFailed, failed)
	}
	if len(exist) != 0 {
		text += fmt.Sprintf(ba.Bot.Translate.AddAdminsExist, exist)
	}
	if len(added) != 0 {
		text += fmt.Sprintf(ba.Bot.Translate.AddAdminsAdded, added)
	}

	return text
}

func (ba *BotAPI) listMembers(channelID, role string) string {
	members, err := ba.Bot.DB.ListChannelMembersByRole(channelID, role)
	if err != nil {
		return fmt.Sprintf("failed to list members :%v\n", err)
	}
	var userIDs []string
	for _, user := range members {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if role == "pm" {
		if len(userIDs) < 1 {
			return ba.Bot.Translate.ListNoPMs
		}
		return fmt.Sprintf(ba.Bot.Translate.ListPMs, strings.Join(userIDs, ", "))
	}
	if len(userIDs) < 1 {
		localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)
		listNoStandupers := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID: "ListNoStandupers",
			},
		})

		return listNoStandupers
	}
	return fmt.Sprintf(ba.Bot.Translate.ListStandupers, strings.Join(userIDs, ", "))
}

func (ba *BotAPI) listAdmins() string {
	admins, err := ba.Bot.DB.ListAdmins()
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		return ba.Bot.Translate.ListNoAdmins
	}
	return fmt.Sprintf(ba.Bot.Translate.ListAdmins, strings.Join(userNames, ", "))
}

func (ba *BotAPI) deleteMembers(members []string, channelID string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range members {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := ba.Bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed += u
			continue
		}
		ba.Bot.DB.DeleteChannelMember(user.UserID, channelID)
		deleted += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf("Could not remove the following members: %v\n", failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf("The following members were removed: %v\n", deleted)
	}

	return text
}

func (ba *BotAPI) deleteAdmins(users []string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := ba.Bot.DB.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role != "admin" {
			failed += u
			continue
		}
		user.Role = ""
		ba.Bot.DB.UpdateUser(user)
		message := fmt.Sprintf(ba.Bot.Translate.PMRemoved)
		err = ba.Bot.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(ba.Bot.Translate.DeleteAdminsFailed, failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf(ba.Bot.Translate.DeleteAdminsSucceed, deleted)
	}

	return text
}
