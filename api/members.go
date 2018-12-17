package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (r *REST) addCommand(accessLevel int, channelID, params string) string {
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
			return r.slack.Translate.AccessAtLeastAdmin
		}
		return r.addAdmins(members)
	case "developer", "разработчик", "":
		if accessLevel > 3 {
			return r.slack.Translate.AccessAtLeastPM
		}
		return r.addMembers(members, "developer", channelID)
	case "pm", "пм":
		if accessLevel > 2 {
			return r.slack.Translate.AccessAtLeastAdmin
		}
		return r.addMembers(members, "pm", channelID)
	default:
		return r.displayHelpText("add")
	}
}

func (r *REST) listCommand(channelID, params string) string {
	switch params {
	case "admin", "админ":
		return r.listAdmins()
	case "developer", "разработчик", "":
		return r.listMembers(channelID, "developer")
	case "pm", "пм":
		return r.listMembers(channelID, "pm")
	default:
		return r.displayHelpText("show")
	}
}

func (r *REST) deleteCommand(accessLevel int, channelID, params string) string {
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
			return r.slack.Translate.AccessAtLeastAdmin
		}
		return r.deleteAdmins(members)
	case "developer", "разработчик", "pm", "пм", "":
		if accessLevel > 3 {
			return r.slack.Translate.AccessAtLeastPM
		}
		return r.deleteMembers(members, channelID)
	default:
		return r.displayHelpText("remove")
	}
}

func (r *REST) addMembers(users []string, role, channel string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			logrus.Errorf("Rest FindChannelMemberByUserID failed: %v", err)
			chanMember, _ := r.db.CreateChannelMember(model.ChannelMember{
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
			text += fmt.Sprintf(r.slack.Translate.AddPMsFailed, failed)
		} else {
			text += fmt.Sprintf(r.slack.Translate.AddMembersFailed, failed)
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			text += fmt.Sprintf(r.slack.Translate.AddPMsExist, exist)
		} else {
			text += fmt.Sprintf(r.slack.Translate.AddMembersExist, exist)
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			text += fmt.Sprintf(r.slack.Translate.AddPMsAdded, added)
		} else {
			text += fmt.Sprintf(r.slack.Translate.AddMembersAdded, added)
		}

	}
	return text
}

func (r *REST) addAdmins(users []string) string {
	var failed, exist, added, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role == "admin" {
			exist += u
			continue
		}
		user.Role = "admin"
		r.db.UpdateUser(user)
		message := r.slack.Translate.PMAssigned
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(r.slack.Translate.AddAdminsFailed, failed)
	}
	if len(exist) != 0 {
		text += fmt.Sprintf(r.slack.Translate.AddAdminsExist, exist)
	}
	if len(added) != 0 {
		text += fmt.Sprintf(r.slack.Translate.AddAdminsAdded, added)
	}

	return text
}

func (r *REST) listMembers(channelID, role string) string {
	members, err := r.db.ListChannelMembersByRole(channelID, role)
	if err != nil {
		return fmt.Sprintf("failed to list members :%v\n", err)
	}
	var userIDs []string
	for _, user := range members {
		userIDs = append(userIDs, "<@"+user.UserID+">")
	}
	if role == "pm" {
		if len(userIDs) < 1 {
			return r.slack.Translate.ListNoPMs
		}
		return fmt.Sprintf(r.slack.Translate.ListPMs, strings.Join(userIDs, ", "))
	}
	if len(userIDs) < 1 {
		return r.slack.Translate.ListNoStandupers
	}
	return fmt.Sprintf(r.slack.Translate.ListStandupers, strings.Join(userIDs, ", "))
}

func (r *REST) listAdmins() string {
	admins, err := r.db.ListAdmins()
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		return r.slack.Translate.ListNoAdmins
	}
	return fmt.Sprintf(r.slack.Translate.ListAdmins, strings.Join(userNames, ", "))
}

func (r *REST) deleteMembers(members []string, channelID string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range members {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed += u
			continue
		}
		r.db.DeleteChannelMember(user.UserID, channelID)
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

func (r *REST) deleteAdmins(users []string) string {
	var failed, deleted, text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed += u
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := r.db.SelectUser(userID)
		if err != nil {
			failed += u
			continue
		}
		if user.Role != "admin" {
			failed += u
			continue
		}
		user.Role = ""
		r.db.UpdateUser(user)
		message := fmt.Sprintf(r.slack.Translate.PMRemoved)
		err = r.slack.SendUserMessage(userID, message)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted += u
	}

	if len(failed) != 0 {
		text += fmt.Sprintf(r.slack.Translate.DeleteAdminsFailed, failed)
	}
	if len(deleted) != 0 {
		text += fmt.Sprintf(r.slack.Translate.DeleteAdminsSucceed, deleted)
	}

	return text
}
