package botuser

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/translation"
	"github.com/maddevsio/comedian/utils"
)

func (bot *Bot) addCommand(accessLevel int, channelID, params string) string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
	accessAtLeastPM := translation.Translate(payload)

	if accessLevel > pmAccess {
		return accessAtLeastPM
	}

	var role string
	var members []string
	if strings.Contains(params, "/") {
		dividedText := strings.Split(params, "/")
		if len(dividedText) != 2 {
			return bot.DisplayHelpText("add")
		}
		role = strings.TrimSpace(dividedText[1])
		members = strings.Fields(dividedText[0])
	} else {
		role = "developer"
		members = strings.Fields(params)
	}

	switch role {
	case "developer", "разработчик", "":
		return bot.addMembers(members, "developer", channelID)
	case "designer", "дизайнер":
		return bot.addMembers(members, "designer", channelID)
	case "pm", "пм":
		return bot.addMembers(members, "pm", channelID)
	case "tester", "тестер":
		return bot.addMembers(members, "tester", channelID)
	default:
		return bot.DisplayHelpText("add")
	}
}

func (bot *Bot) showCommand(channelID, params string) string {
	switch params {
	case "admin", "админ":
		return bot.listAdmins()
	default:
		return bot.listMembers(channelID)
	}
}

func (bot *Bot) deleteCommand(accessLevel int, channelID, params string) string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
	accessAtLeastPM := translation.Translate(payload)
	if accessLevel > pmAccess {
		return accessAtLeastPM
	}
	return bot.deleteMembers(strings.Fields(params), channelID)
}

func (bot *Bot) addMembers(users []string, role, channel string) string {

	var failed, added []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			log.WithFields(log.Fields(map[string]interface{}{"place": "if !rg.MatchString(u)", "user": u, "error": "could not match user to template"})).Error("Error in AddMembers")
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)

		var user model.User
		var err error

		user, err = bot.db.SelectUser(userID)
		if err != nil {
			err := bot.AddNewSlackUser(userID)
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.AddNewSlackUser(u)", "user": u, "error": err})).Error("Error in AddMembers")
				failed = append(failed, u)
				continue
			}
			user, err = bot.db.SelectUser(userID)
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.SelectUser(userID)", "user": u, "error": err})).Error("Error in AddMembers")
				failed = append(failed, u)
				continue
			}
		}

		standuper, err := bot.db.FindStansuperByUserID(userID, channel)

		if standuper.UserID == userID && standuper.ChannelID == channel {
			standuper.RoleInChannel = role
			_, err := bot.db.UpdateStanduper(standuper)
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.UpdateStanduper(standuper)", "user": u, "error": err})).Error("Error in AddMembers")
				failed = append(failed, u)
				continue
			}
			added = append(added, u)
			continue
		}

		if err != nil {
			ch, err := bot.db.SelectChannel(channel)
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.SelectChannel(channel)", "user": u, "error": err})).Error("Error in AddMembers")
				failed = append(failed, u)
				continue
			}
			_, err = bot.db.CreateStanduper(model.Standuper{
				TeamID:                bot.properties.TeamID,
				UserID:                userID,
				ChannelID:             channel,
				ChannelName:           ch.ChannelName,
				RoleInChannel:         role,
				SubmittedStandupToday: false,
				RealName:              user.RealName,
			})
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.CreateStanduper", "user": u, "error": err})).Error("Error in AddMembers")
				failed = append(failed, u)
				continue
			}

			payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "MemberAssigned", 0, map[string]interface{}{"role": role, "channelID": ch.ChannelID, "channelName": ch.ChannelName}}
			err = bot.SendUserMessage(userID, translation.Translate(payload))
			if err != nil {
				log.WithFields(log.Fields(map[string]interface{}{"place": "bot.SendUserMessage", "user": u, "error": err})).Error("Error in AddMembers")
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

func (bot *Bot) listMembers(channelID string) (result string) {
	members, err := bot.db.ListChannelStandupers(channelID)
	if err != nil || len(members) == 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoStandupers", 0, nil}
		return translation.Translate(payload)
	}

	var managers []string
	var developers []string
	var designers []string
	var testers []string

	for _, member := range members {
		switch member.RoleInChannel {
		case "pm":
			managers = append(managers, "<@"+member.UserID+">")
		case "developer":
			developers = append(developers, "<@"+member.UserID+">")
		case "designer":
			designers = append(designers, "<@"+member.UserID+">")
		case "tester":
			testers = append(testers, "<@"+member.UserID+">")
		}
	}

	if len(managers) == 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoPMs", 0, nil}
		result += translation.Translate(payload) + "\n"
	} else {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListPMs", len(managers), map[string]interface{}{"pm": managers[0], "pms": strings.Join(managers, ", ")}}
		result += translation.Translate(payload) + "\n"
	}

	if len(developers) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListDevelopers", len(developers), map[string]interface{}{"developer": developers[0], "developers": strings.Join(developers, ", ")}}
		result += translation.Translate(payload) + "\n"
	}

	if len(designers) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListDesigners", len(designers), map[string]interface{}{"designer": designers[0], "designers": strings.Join(designers, ", ")}}
		result += translation.Translate(payload) + "\n"
	}

	if len(testers) != 0 {
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListTesters", len(testers), map[string]interface{}{"tester": testers[0], "testers": strings.Join(testers, ", ")}}
		result += translation.Translate(payload) + "\n"
	}
	return result
}

func (bot *Bot) listAdmins() string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListNoAdmins", 0, nil}
	listNoAdmins := translation.Translate(payload)

	users, err := bot.db.ListUsers()
	if err != nil {
		return listNoAdmins
	}
	var userNames []string
	for _, user := range users {
		if user.IsAdmin() && user.TeamID == bot.properties.TeamID {
			userNames = append(userNames, "<@"+user.UserName+">")
		}
	}

	if len(userNames) < 1 {
		return listNoAdmins
	}
	payload = translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "ListAdmins", len(userNames), map[string]interface{}{"admin": userNames[0], "admins": strings.Join(userNames, ", ")}}
	return translation.Translate(payload)
}

func (bot *Bot) deleteMembers(members []string, channelID string) string {

	var failed, deleted []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range members {
		if !rg.MatchString(u) {
			log.WithFields(log.Fields(map[string]interface{}{"place": "bot.SendUserMessage", "user": u, "error": "could not match user to template"})).Error("Error in deleteMembers")
			failed = append(failed, u)
			continue
		}

		userID, _ := utils.SplitUser(u)

		member, err := bot.db.FindStansuperByUserID(userID, channelID)
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.FindStansuperByUserID", "user": u, "error": err})).Error("Error in deleteMembers")
			failed = append(failed, u)
			continue
		}

		ch, err := bot.db.SelectChannel(channelID)
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.SelectChannel(channelID)", "user": u, "error": err})).Error("Error in deleteMembers")
			failed = append(failed, u)
			continue
		}

		err = bot.db.DeleteStanduper(member.ID)
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"place": "bot.db.DeleteStanduper(member.ID)", "user": u, "error": err})).Error("Error in deleteMembers")
			failed = append(failed, u)
			continue
		}
		payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, "MemberRemoved", 0, map[string]interface{}{"role": member.RoleInChannel, "channelID": ch.ChannelID, "channelName": ch.ChannelName}}
		err = bot.SendUserMessage(userID, translation.Translate(payload))
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"place": "bot.SendUserMessage", "user": u, "error": err})).Error("Error in deleteMembers")
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
