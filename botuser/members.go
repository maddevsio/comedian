package botuser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/translation"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (bot *Bot) addCommand(accessLevel int, channelID, params string) string {

	accessAtLeastAdmin, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastAdmin", 0, nil)
	accessAtLeastPM, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastPM", 0, nil)

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
		if accessLevel > adminAccess {
			return accessAtLeastAdmin
		}
		return bot.addAdmins(members)
	case "developer", "разработчик", "":
		if accessLevel > pmAccess {
			return accessAtLeastPM
		}
		return bot.addMembers(members, "developer", channelID)
	case "pm", "пм":
		if accessLevel > adminAccess {
			return accessAtLeastAdmin
		}
		return bot.addMembers(members, "pm", channelID)
	default:
		return bot.DisplayHelpText("add")
	}
}

func (bot *Bot) showCommand(channelID, params string) string {
	switch params {
	case "admin", "админ":
		return bot.listAdmins()
	case "developer", "разработчик", "":
		return bot.listMembers(channelID, "developer")
	case "pm", "пм":
		return bot.listMembers(channelID, "pm")
	default:
		return bot.DisplayHelpText("show")
	}
}

func (bot *Bot) deleteCommand(accessLevel int, channelID, params string) string {
	accessAtLeastAdmin, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastAdmin", 0, nil)
	accessAtLeastPM, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AccessAtLeastPM", 0, nil)

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
		if accessLevel > adminAccess {
			return accessAtLeastAdmin
		}
		return bot.deleteAdmins(members)
	case "developer", "разработчик", "pm", "пм", "":
		if accessLevel > pmAccess {
			return accessAtLeastPM
		}
		return bot.deleteMembers(members, channelID)
	default:
		return bot.DisplayHelpText("remove")
	}
}

func (bot *Bot) addMembers(users []string, role, channel string) string {

	var failed, exist, added []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.db.FindChannelMemberByUserID(userID, channel)

		if user.UserID == userID && user.ChannelID == channel {
			exist = append(exist, u)
			continue
		}

		if err != nil {
			chanMember, _ := bot.db.CreateChannelMember(model.ChannelMember{
				TeamID:        bot.Properties.TeamID,
				UserID:        userID,
				ChannelID:     channel,
				RoleInChannel: role,
			})
			logrus.Infof("ChannelMember created! ID:%v", chanMember.ID)
		}

		added = append(added, u)
	}

	if len(failed) != 0 {
		if role == "pm" {
			addPMsFailed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddPMsFailed", len(failed), map[string]interface{}{
				"PM":  failed[0],
				"PMs": strings.Join(failed, ", "),
			})
			text += addPMsFailed

		} else {
			addMembersFailed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddMembersFailed", len(failed), map[string]interface{}{
				"user":  failed[0],
				"users": strings.Join(failed, ", "),
			})
			text += addMembersFailed
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			addPMsExist, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddPMsExist", len(exist), map[string]interface{}{
				"PM":  exist[0],
				"PMs": strings.Join(exist, ", "),
			})
			text += addPMsExist
		} else {
			addMembersExist, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddMembersExist", len(exist), map[string]interface{}{
				"user":  exist[0],
				"users": strings.Join(exist, ", "),
			})
			text += addMembersExist
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			addPMsAdded, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddPMsAdded", len(added), map[string]interface{}{
				"PM":  added[0],
				"PMs": strings.Join(added, ", "),
			})
			text += addPMsAdded

		} else {
			addMembersAdded, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddMembersAdded", len(added), map[string]interface{}{
				"user":  added[0],
				"users": strings.Join(added, ", "),
			})
			text += addMembersAdded
		}

	}
	return text
}

func (bot *Bot) addAdmins(users []string) string {

	var failed, exist, added []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.db.SelectUser(userID)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		if user.Role == "admin" {
			exist = append(exist, u)
			continue
		}
		user.Role = "admin"
		bot.db.UpdateUser(user)

		adminAssigned, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AdminAssigned", 0, nil)
		err = bot.SendUserMessage(userID, adminAssigned)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added = append(added, u)
	}

	if len(failed) != 0 {
		addAdminsFailed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddAdminsFailed", len(failed), map[string]interface{}{
			"admin":  failed[0],
			"admins": strings.Join(failed, ", "),
		})
		text += addAdminsFailed
	}
	if len(exist) != 0 {
		addAdminsExist, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddAdminsExist", len(exist), map[string]interface{}{
			"admin":  exist[0],
			"admins": strings.Join(exist, ", "),
		})
		text += addAdminsExist
	}
	if len(added) != 0 {
		addAdminsAdded, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AddAdminsAdded", len(added), map[string]interface{}{
			"admin":  added[0],
			"admins": strings.Join(added, ", "),
		})
		text += addAdminsAdded
	}

	return text
}

func (bot *Bot) listMembers(channelID, role string) string {

	members, err := bot.db.ListChannelMembersByRole(channelID, role)
	if err != nil {
		return fmt.Sprintf("failed to list members :%v\n", err)
	}
	var userIDs []string
	for _, user := range members {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if role == "pm" {
		if len(userIDs) < 1 {
			listNoPMs, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListNoPMs", 0, nil)
			return listNoPMs
		}

		listPMs, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListPMs", len(userIDs), map[string]interface{}{
			"pm":  userIDs[0],
			"pms": strings.Join(userIDs, ", "),
		})

		return listPMs

	}
	if len(userIDs) < 1 {
		listNoStandupers, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListNoStandupers", 0, nil)
		return listNoStandupers

	}
	listStandupers, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListStandupers", len(userIDs), map[string]interface{}{
		"standuper":  userIDs[0],
		"standupers": strings.Join(userIDs, ", "),
	})
	return listStandupers
}

func (bot *Bot) listAdmins() string {

	admins, err := bot.db.ListAdmins()
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		listNoAdmins, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListNoAdmins", 0, nil)
		return listNoAdmins
	}
	listAdmins, _ := translation.Translate(bot.bundle, bot.Properties.Language, "ListAdmins", len(userNames), map[string]interface{}{
		"admin":  userNames[0],
		"admins": strings.Join(userNames, ", "),
	})
	return listAdmins

}

func (bot *Bot) deleteMembers(members []string, channelID string) string {

	var failed, deleted []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range members {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)

		user, err := bot.db.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed = append(failed, u)
			continue
		}

		bot.db.DeleteChannelMember(user.UserID, channelID)
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		deleteMembersFailed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "DeleteMembersFailed", len(failed), map[string]interface{}{
			"user":  failed[0],
			"users": strings.Join(failed, ", "),
		})
		text += deleteMembersFailed
	}
	if len(deleted) != 0 {
		deleteMembersSucceed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "DeleteMembersSucceed", len(deleted), map[string]interface{}{
			"user":  deleted[0],
			"users": strings.Join(deleted, ", "),
		})
		text += deleteMembersSucceed
	}

	return text
}

func (bot *Bot) deleteAdmins(users []string) string {

	var failed, deleted []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.db.SelectUser(userID)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		if user.Role != "admin" {
			failed = append(failed, u)
			continue
		}
		user.Role = ""
		bot.db.UpdateUser(user)
		adminRemoved, _ := translation.Translate(bot.bundle, bot.Properties.Language, "AdminRemoved", 0, nil)

		err = bot.SendUserMessage(userID, adminRemoved)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		deleteAdminsFailed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "DeleteAdminsFailed", len(failed), map[string]interface{}{
			"admin":  failed[0],
			"admins": strings.Join(failed, ", "),
		})
		text += deleteAdminsFailed
	}
	if len(deleted) != 0 {
		deleteAdminsSucceed, _ := translation.Translate(bot.bundle, bot.Properties.Language, "DeleteAdminsSucceed", len(deleted), map[string]interface{}{
			"admin":  deleted[0],
			"admins": strings.Join(deleted, ", "),
		})
		text += deleteAdminsSucceed
	}

	return text
}
