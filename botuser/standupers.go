package botuser

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/translation"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (bot *Bot) addCommand(accessLevel int, channelID, params string) string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastAdmin", 0, nil}
	accessAtLeastAdmin := translation.Translate(payload)

	payload = translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
	accessAtLeastPM := translation.Translate(payload)

	if accessLevel > pmAccess {
		return accessAtLeastPM
	}

	var role string
	var members []string
	if strings.Contains(params, "/") {
		dividedText := strings.Split(params, "/")
		if len(dividedText) != 2 {
			return "wrong username. Try again with correct username"
		}
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
		return bot.addMembers(members, "developer", channelID)
	case "designer", "дизайнер":
		return bot.addMembers(members, "designer", channelID)
	case "pm", "пм":
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
	case "designer", "дизайнер":
		return bot.listMembers(channelID, "designer")
	case "pm", "пм":
		return bot.listMembers(channelID, "pm")
	default:
		return bot.DisplayHelpText("show")
	}
}

func (bot *Bot) deleteCommand(accessLevel int, channelID, params string) string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastAdmin", 0, nil}
	accessAtLeastAdmin := translation.Translate(payload)
	payload = translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
	accessAtLeastPM := translation.Translate(payload)

	if accessLevel > pmAccess {
		return accessAtLeastPM
	}

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
		user, err := bot.db.FindStansuperByUserID(userID, channel)

		if user.UserID == userID && user.ChannelID == channel {
			exist = append(exist, u)
			continue
		}

		if err != nil {
			ch, err := bot.db.SelectChannel(channel)
			if err != nil {
				log.Error(err)
				failed = append(failed, u)
				continue
			}
			standuper, err := bot.db.CreateStanduper(model.Standuper{
				TeamID:                bot.properties.TeamID,
				UserID:                userID,
				ChannelID:             channel,
				RoleInChannel:         role,
				SubmittedStandupToday: false,
			})
			if err != nil {
				log.Error(err)
				failed = append(failed, u)
				continue
			}
			log.Infof("ChannelMember created! ID:%v", standuper.ID)

			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "MemberAssigned", 0, map[string]interface{}{"role": role, "channelID": ch.ChannelID, "channelName": ch.ChannelName}}
			err = bot.SendUserMessage(userID, translation.Translate(payload))
			if err != nil {
				log.Errorf("rest: SendUserMessage failed: %v\n", err)
			}
		}

		added = append(added, u)
	}

	if len(failed) != 0 {
		if role == "pm" {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddPMsFailed", len(failed), map[string]interface{}{"PM": failed[0], "PMs": strings.Join(failed, ", ")}}
			addPMsFailed := translation.Translate(payload)
			text += addPMsFailed

		} else {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddMembersFailed", len(failed), map[string]interface{}{"user": failed[0], "users": strings.Join(failed, ", ")}}
			addMembersFailed := translation.Translate(payload)
			text += addMembersFailed
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddPMsExist", len(exist), map[string]interface{}{"PM": exist[0], "PMs": strings.Join(exist, ", ")}}
			addPMsExist := translation.Translate(payload)
			text += addPMsExist
		} else {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddMembersExist", len(exist), map[string]interface{}{"user": exist[0], "users": strings.Join(exist, ", ")}}
			addMembersExist := translation.Translate(payload)
			text += addMembersExist
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddPMsAdded", len(added), map[string]interface{}{"PM": added[0], "PMs": strings.Join(added, ", ")}}
			addPMsAdded := translation.Translate(payload)
			text += addPMsAdded

		} else {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddMembersAdded", len(added), map[string]interface{}{"user": added[0], "users": strings.Join(added, ", ")}}
			addMembersAdded := translation.Translate(payload)
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
		// <@foo> passes this check... need to fix it later
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

		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AdminAssigned", 0, nil}
		_, err = bot.db.UpdateUser(user)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		err = bot.SendUserMessage(userID, translation.Translate(payload))
		if err != nil {
			log.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added = append(added, u)
	}

	if len(failed) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddAdminsFailed", len(failed), map[string]interface{}{"admin": failed[0], "admins": strings.Join(failed, ", ")}}
		addAdminsFailed := translation.Translate(payload)
		text += addAdminsFailed
	}
	if len(exist) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddAdminsExist", len(exist), map[string]interface{}{"admin": exist[0], "admins": strings.Join(exist, ", ")}}
		addAdminsExist := translation.Translate(payload)
		text += addAdminsExist
	}
	if len(added) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AddAdminsAdded", len(added), map[string]interface{}{"admin": added[0], "admins": strings.Join(added, ", ")}}
		addAdminsAdded := translation.Translate(payload)
		text += addAdminsAdded
	}

	return text
}

func (bot *Bot) listMembers(channelID, role string) string {

	standupers, err := bot.db.ListChannelStandupers(channelID)
	if err != nil {
		return "could not list members"
	}
	var userIDs []string
	for _, standuper := range standupers {
		if standuper.RoleInChannel == role {
			userIDs = append(userIDs, "<@"+standuper.UserID+">")
		}
	}
	if role == "pm" {
		if len(userIDs) < 1 {
			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoPMs", 0, nil}
			listNoPMs := translation.Translate(payload)
			return listNoPMs
		}

		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListPMs", len(userIDs), map[string]interface{}{"pm": userIDs[0], "pms": strings.Join(userIDs, ", ")}}
		listPMs := translation.Translate(payload)

		return listPMs

	}

	if len(userIDs) < 1 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoStandupers", 0, nil}
		listNoStandupers := translation.Translate(payload)
		return listNoStandupers

	}

	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListStandupers", len(userIDs), map[string]interface{}{"standuper": userIDs[0], "standupers": strings.Join(userIDs, ", ")}}
	listStandupers := translation.Translate(payload)
	return listStandupers
}

func (bot *Bot) listAdmins() string {

	users, err := bot.db.ListUsers()
	if err != nil {
		return "could not list users"
	}
	var userNames []string
	for _, user := range users {
		if user.IsAdmin() && user.TeamID == bot.properties.TeamID {
			userNames = append(userNames, "<@"+user.UserName+">")
		}
	}
	if len(userNames) < 1 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoAdmins", 0, nil}
		listNoAdmins := translation.Translate(payload)
		return listNoAdmins
	}
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListAdmins", len(userNames), map[string]interface{}{"admin": userNames[0], "admins": strings.Join(userNames, ", ")}}
	listAdmins := translation.Translate(payload)
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

		// need to make sure people have different userID across teams
		member, err := bot.db.FindStansuperByUserID(userID, channelID)
		if err != nil {
			log.Errorf("rest: FindStansuperByUserID failed: %v\n", err)
			failed = append(failed, u)
			continue
		}

		ch, err := bot.db.SelectChannel(channelID)
		if err != nil {
			log.Error(err)
			failed = append(failed, u)
			continue
		}

		bot.db.DeleteStanduper(member.ID)
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "MemberRemoved", 0, map[string]interface{}{"role": member.RoleInChannel, "channelID": ch.ChannelID, "channelName": ch.ChannelName}}
		err = bot.SendUserMessage(userID, translation.Translate(payload))
		if err != nil {
			log.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "DeleteMembersFailed", len(failed), map[string]interface{}{"user": failed[0], "users": strings.Join(failed, ", ")}}
		deleteMembersFailed := translation.Translate(payload)
		text += deleteMembersFailed
	}
	if len(deleted) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "DeleteMembersSucceed", len(deleted), map[string]interface{}{"user": deleted[0], "users": strings.Join(deleted, ", ")}}
		deleteMembersSucceed := translation.Translate(payload)
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
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AdminRemoved", 0, nil}
		adminRemoved := translation.Translate(payload)
		_, err = bot.db.UpdateUser(user)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		err = bot.SendUserMessage(userID, adminRemoved)
		if err != nil {
			log.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "DeleteAdminsFailed", len(failed), map[string]interface{}{"admin": failed[0], "admins": strings.Join(failed, ", ")}}
		deleteAdminsFailed := translation.Translate(payload)
		text += deleteAdminsFailed
	}
	if len(deleted) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "DeleteAdminsSucceed", len(deleted), map[string]interface{}{"admin": deleted[0], "admins": strings.Join(deleted, ", ")}}
		deleteAdminsSucceed := translation.Translate(payload)
		text += deleteAdminsSucceed
	}

	return text
}
