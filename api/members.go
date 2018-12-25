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
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)
	accessAtLeastAdmin := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastAdmin",
			Description: "Display warning that role must be at least admin",
			Other:       "Access Denied! You need to be at least admin in this slack to use this command!",
		},
	})

	accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastPM",
			Description: "Display warning that role must be at least pm",
			Other:       "Access Denied! You need to be at least PM in this project to use this command!",
		},
	})

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
			return accessAtLeastAdmin
		}
		return ba.addAdmins(members)
	case "developer", "разработчик", "":
		if accessLevel > 3 {
			return accessAtLeastPM
		}
		return ba.addMembers(members, "developer", channelID)
	case "pm", "пм":
		if accessLevel > 2 {
			return accessAtLeastAdmin
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
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)
	accessAtLeastAdmin := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastAdmin",
			Description: "Display warning that role must be at least admin",
			Other:       "Access Denied! You need to be at least admin in this slack to use this command!",
		},
	})

	accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastPM",
			Description: "Display warning that role must be at least pm",
			Other:       "Access Denied! You need to be at least PM in this project to use this command!",
		},
	})

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
			return accessAtLeastAdmin
		}
		return ba.deleteAdmins(members)
	case "developer", "разработчик", "pm", "пм", "":
		if accessLevel > 3 {
			return accessAtLeastPM
		}
		return ba.deleteMembers(members, channelID)
	default:
		return DisplayHelpText("remove")
	}
}

func (ba *BotAPI) addMembers(users []string, role, channel string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
			addPMsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsFailed",
					Description: "Displays a message when errors occur when assigning users as PM",
					Other:       "Could not assign users as PMs: {{.PMs}}",
				},
				TemplateData: map[string]interface{}{
					"PMs": failed,
				},
			})
			text += addPMsFailed
		} else {
			addMembersFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersFailed",
					Description: "Displays a message when errors occur when assigning users",
					Other:       "Could not assign members: {{.users}}",
				},
				TemplateData: map[string]interface{}{
					"users": failed,
				},
			})
			text += addMembersFailed
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			addPMsExist := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsExist",
					Description: "Displays a message when errors occur when assigning users, which has already PM-role",
					Other:       "Users already have roles: {{.PMs}}\n",
				},
				TemplateData: map[string]interface{}{
					"PMs": exist,
				},
			})
			text += addPMsExist
		} else {
			addMembersExist := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersExist",
					Description: "Displays a message when errors occur when assigning users, which already has role",
					Other:       "Members already have roles: {{.users}}\n",
				},
				TemplateData: map[string]interface{}{
					"users": exist,
				},
			})
			text += addMembersExist
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			addPMsAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsAdded",
					Description: "Display a message when users successfully assigning as PMs",
					Other:       "Users are assigned as PMs: {{.PMs}}\n",
				},
				TemplateData: map[string]interface{}{
					"PMs": added,
				},
			})
			text += addPMsAdded
		} else {
			addMembersAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersAdded",
					Description: "Display a message when users are successfully assigned",
					Other:       "Members are assigned: {{.users}}\n",
				},
				TemplateData: map[string]interface{}{
					"users": added,
				},
			})
			text += addMembersAdded
		}

	}
	return text
}

func (ba *BotAPI) addAdmins(users []string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		adminAssigned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "PMAssigned",
				Description: "Display message when user added as admin for Comedian",
				Other:       "You have been added as Admin for Comedian",
			},
		})

		err = ba.Bot.SendUserMessage(userID, adminAssigned)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added += u
	}

	if len(failed) != 0 {
		addAdminsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsFailed",
				Description: "Display message when user added as admin for Comedian",
				Other:       "Could not assign users as admins: {{.admins}}",
			},
			TemplateData: map[string]interface{}{
				"admins": failed,
			},
		})
		text += addAdminsFailed
	}
	if len(exist) != 0 {
		addAdminsExist := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsExist",
				Description: "",
				Other:       "Users were already assigned as admins: {{.admins}}\n",
			},
			TemplateData: map[string]interface{}{
				"admins": exist,
			},
		})
		text += addAdminsExist
	}
	if len(added) != 0 {
		addAdminsAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsAdded",
				Description: "",
				Other:       "Users are assigned as admins: {{.admins}}\n",
			},
			TemplateData: map[string]interface{}{
				"admins": added,
			},
		})
		text += addAdminsAdded
	}

	return text
}

func (ba *BotAPI) listMembers(channelID, role string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
			listNoPMs := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "ListNoPMs",
					Description: "Displays message about there are no pms in channel",
					Other:       "No PMs in this channel! To add one, please, use `/add` slash command",
				},
			})
			return listNoPMs
		}
		listPMs := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ListPMs",
				Description: "Display list of pms",
				One:         "PM in this channel: {{.pm}}",
				Other:       "PMs in this channel: {{.pms}}",
			},
			PluralCount: len(userIDs),
			TemplateData: map[string]interface{}{
				"pm":  userIDs[0],
				"pms": strings.Join(userIDs, ", "),
			},
		})
		return listPMs

	}
	if len(userIDs) < 1 {
		listNoStandupers := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ListNoStandupers",
				Description: "Displays message when there are no standupers in the channel",
				Other:       "No standupers in this channel! To add one, please, use `/add` slash command",
			},
		})
		return listNoStandupers

	}
	listStandupers := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ListStandupers",
			Description: "Displays list of standupers",
			One:         "Standuper in this channel: {{.standuper}}",
			Other:       "Standupers in this channel: {{.standupers}}",
		},
		PluralCount: len(userIDs),
		TemplateData: map[string]interface{}{
			"standuper":  userIDs[0],
			"standupers": strings.Join(userIDs, ", "),
		},
	})
	return listStandupers

}

func (ba *BotAPI) listAdmins() string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	admins, err := ba.Bot.DB.ListAdmins()
	if err != nil {
		return fmt.Sprintf("failed to list users :%v\n", err)
	}
	var userNames []string
	for _, admin := range admins {
		userNames = append(userNames, "<@"+admin.UserName+">")
	}
	if len(userNames) < 1 {
		listNoAdmins := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ListNoAdmins",
				Description: "Displays message when there are no admins in the channel",
				Other:       "No admins in this workspace! To add one, please, use `/add` slash command",
			},
		})
		return listNoAdmins

	}
	listAdmins := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ListAdmins",
			Description: "Displays list of admins",
			One:         "Admin in this workspace: {{.admin}}",
			Other:       "Admins in this workspace: {{.admins}}",
		},
		PluralCount: len(userNames),
		TemplateData: map[string]interface{}{
			"admin":  userNames[0],
			"admins": strings.Join(userNames, ", "),
		},
	})
	return listAdmins

}

func (ba *BotAPI) deleteMembers(members []string, channelID string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		deleteMembersFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteMembersFailed",
				Description: "",
				Other:       "Could not remove the following members: {{.users}}\n",
			},
			TemplateData: map[string]interface{}{
				"users": failed,
			},
		})
		text += deleteMembersFailed
	}
	if len(deleted) != 0 {
		deleteMembersSucceed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteMembersSucceed",
				Description: "",
				Other:       "The following members were removed: {{.users}}\n",
			},
			TemplateData: map[string]interface{}{
				"users": deleted,
			},
		})
		text += deleteMembersSucceed
	}

	return text
}

func (ba *BotAPI) deleteAdmins(users []string) string {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

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
		adminRemoved := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "PMRemoved",
				Description: "Display message when user removed as admin from Comedian",
				Other:       "You have been removed as Admin from Comedian",
			},
		})

		err = ba.Bot.SendUserMessage(userID, adminRemoved)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted += u
	}

	if len(failed) != 0 {
		deleteAdminsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteAdminsFailed",
				Description: "",
				Other:       "Could not remove users as admins: {{.admins}}\n",
			},
			TemplateData: map[string]interface{}{
				"admins": failed,
			},
		})
		text += deleteAdminsFailed
	}
	if len(deleted) != 0 {
		deleteAdminsSucceed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteAdminsSucceed",
				Description: "",
				Other:       "Users were removed as admins: {{.admins}}\n",
			},
			TemplateData: map[string]interface{}{
				"admins": deleted,
			},
		})
		text += deleteAdminsSucceed
	}

	return text
}
