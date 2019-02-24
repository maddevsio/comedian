package botuser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/utils"
)

func (bot *Bot) addCommand(accessLevel int, channelID, params string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
	accessAtLeastAdmin := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastAdmin",
			Description: "Displays warning that role must be at least admin",
			Other:       "Access Denied! You need to be at least admin in this slack to use this command!",
		},
	})

	accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastPM",
			Description: "Displays warning that role must be at least pm",
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
		return bot.addAdmins(members)
	case "developer", "разработчик", "":
		if accessLevel > 3 {
			return accessAtLeastPM
		}
		return bot.addMembers(members, "developer", channelID)
	case "pm", "пм":
		if accessLevel > 2 {
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
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)
	accessAtLeastAdmin := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastAdmin",
			Description: "Displays warning that role must be at least admin",
			Other:       "Access Denied! You need to be at least admin in this slack to use this command!",
		},
	})

	accessAtLeastPM := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AccessAtLeastPM",
			Description: "Displays warning that role must be at least pm",
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
		return bot.deleteAdmins(members)
	case "developer", "разработчик", "pm", "пм", "":
		if accessLevel > 3 {
			return accessAtLeastPM
		}
		return bot.deleteMembers(members, channelID)
	default:
		return bot.DisplayHelpText("remove")
	}
}

func (bot *Bot) addMembers(users []string, role, channel string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	var failed, exist, added []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.DB.FindChannelMemberByUserID(userID, channel)
		if err != nil {
			chanMember, _ := bot.DB.CreateChannelMember(model.ChannelMember{
				TeamID:        bot.Properties.TeamID,
				UserID:        userID,
				ChannelID:     channel,
				RoleInChannel: role,
			})
			logrus.Infof("ChannelMember created! ID:%v", chanMember.ID)
		}
		if user.UserID == userID && user.ChannelID == channel {
			exist = append(exist, u)
			continue
		}
		added = append(added, u)
	}

	if len(failed) != 0 {
		if role == "pm" {
			addPMsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsFailed",
					Description: "Displays a message when errors occur when assigning users as PM",
					One:         "Could not assign user as PM: {{.PM}}. Wrong username.",
					Other:       "Could not assign users as PMs: {{.PMs}}. Wrong usernames.",
				},
				PluralCount: len(failed),
				TemplateData: map[string]interface{}{
					"PM":  failed[0],
					"PMs": strings.Join(failed, ", "),
				},
			})
			text += addPMsFailed + "\n"
		} else {
			addMembersFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersFailed",
					Description: "Displays a message when errors occur when assigning users",
					One:         "Could not assign member: {{.user}} . Wrong username.",
					Other:       "Could not assign members: {{.users}} . Wrong usernames.",
				},
				PluralCount: len(failed),
				TemplateData: map[string]interface{}{
					"user":  failed[0],
					"users": strings.Join(failed, ", "),
				},
			})
			text += addMembersFailed + "\n"
		}

	}
	if len(exist) != 0 {
		if role == "pm" {
			addPMsExist := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsExist",
					Description: "Displays a message when errors occur when assigning users, which has already PM-role",
					One:         "User {{.PM}} already has role.",
					Other:       "Users {{.PMs}} already have roles.",
				},
				PluralCount: len(exist),
				TemplateData: map[string]interface{}{
					"PM":  exist[0],
					"PMs": strings.Join(exist, ", "),
				},
			})
			text += addPMsExist + "\n"
		} else {
			addMembersExist := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersExist",
					Description: "Displays a message when errors occur when assigning users, which already has role",
					One:         "Member {{.user}} already has role.",
					Other:       "Members {{.users}} already have roles.",
				},
				PluralCount: len(exist),
				TemplateData: map[string]interface{}{
					"user":  exist[0],
					"users": strings.Join(exist, ", "),
				},
			})
			text += addMembersExist + "\n"
		}

	}
	if len(added) != 0 {
		if role == "pm" {
			addPMsAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddPMsAdded",
					Description: "Displays a message when users successfully assigning as PMs",
					One:         "User {{.PM}} is assigned as PM.",
					Other:       "Users {{.PMs}} are assigned as PMs.",
				},
				PluralCount: len(added),
				TemplateData: map[string]interface{}{
					"PM":  added[0],
					"PMs": strings.Join(added, ", "),
				},
			})
			text += addPMsAdded + "\n"
		} else {
			addMembersAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:          "AddMembersAdded",
					Description: "Displays a message when users are successfully assigned",
					One:         "Member {{.user}} is assigned.",
					Other:       "Members {{.users}} are assigned",
				},
				PluralCount: len(added),
				TemplateData: map[string]interface{}{
					"user":  added[0],
					"users": strings.Join(added, ", "),
				},
			})
			text += addMembersAdded + "\n"
		}

	}
	return text
}

func (bot *Bot) addAdmins(users []string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	var failed, exist, added []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.DB.SelectUser(userID)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		if user.Role == "admin" {
			exist = append(exist, u)
			continue
		}
		user.Role = "admin"
		bot.DB.UpdateUser(user)
		adminAssigned := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "PMAssigned",
				Description: "Displays message when user added as admin for Comedian",
				Other:       "You have been added as Admin for Comedian",
			},
		})

		err = bot.SendUserMessage(userID, adminAssigned)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		added = append(added, u)
	}

	if len(failed) != 0 {
		addAdminsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsFailed",
				Description: "Displays message when user added as admin for Comedian",
				One:         "Could not assign user as admin: {{.admin}}. User does not exist.",
				Other:       "Could not assign users as admins: {{.admins}}. User does not exist.",
			},
			PluralCount: len(failed),
			TemplateData: map[string]interface{}{
				"admin":  failed[0],
				"admins": strings.Join(failed, ", "),
			},
		})
		text += addAdminsFailed + "\n"
	}
	if len(exist) != 0 {
		addAdminsExist := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsExist",
				Description: "Displays message when users were already assigned as admins",
				One:         "User {{.admin}} was already assigned as admin.",
				Other:       "Users {{.admins}} were already assigned as admins.",
			},
			PluralCount: len(exist),
			TemplateData: map[string]interface{}{
				"admin":  exist[0],
				"admins": strings.Join(exist, ", "),
			},
		})
		text += addAdminsExist + "\n"
	}
	if len(added) != 0 {
		addAdminsAdded := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "AddAdminsAdded",
				Description: "Displays message when users successfully assigned as admins",
				One:         "User {{.admin}} is assigned as admin.",
				Other:       "Users {{.admins}} are assigned as admins.",
			},
			PluralCount: len(added),
			TemplateData: map[string]interface{}{
				"admin":  added[0],
				"admins": strings.Join(added, ", "),
			},
		})
		text += addAdminsAdded + "\n"
	}

	return text
}

func (bot *Bot) listMembers(channelID, role string) string {
	logrus.Info(bot.Properties)
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	members, err := bot.DB.ListChannelMembersByRole(channelID, role)
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
					Description: "Displays message about there are no PMs in channel",
					Other:       "No PMs in this channel! To add one, please, use `/comedian add` slash command",
				},
			})
			return listNoPMs
		}
		listPMs := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "ListPMs",
				Description: "Displays list of pms",
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
				Other:       "No standupers in this channel! To add one, please, use `/comedian add` slash command",
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

func (bot *Bot) listAdmins() string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	admins, err := bot.DB.ListAdmins()
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
				Other:       "No admins in this workspace! To add one, please, use `/comedian add` slash command",
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

func (bot *Bot) deleteMembers(members []string, channelID string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	var failed, deleted []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range members {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.DB.FindChannelMemberByUserID(userID, channelID)
		if err != nil {
			logrus.Errorf("rest: FindChannelMemberByUserID failed: %v\n", err)
			failed = append(failed, u)
			continue
		}
		bot.DB.DeleteChannelMember(user.UserID, channelID)
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		deleteMembersFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteMembersFailed",
				Description: "Displays a message when user deletion errors occur",
				One:         "Could not remove the member: {{.user}} . User is not standuper and not tracked.",
				Other:       "Could not remove the following members: {{.users}}. Users are standupers and not tracked.",
			},
			PluralCount: len(failed),
			TemplateData: map[string]interface{}{
				"user":  failed[0],
				"users": strings.Join(failed, ", "),
			},
		})
		text += deleteMembersFailed + "\n"
	}
	if len(deleted) != 0 {
		deleteMembersSucceed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteMembersSucceed",
				Description: "Displays a message when users have been successfully deleted",
				One:         "The member {{.user}} removed.",
				Other:       "The following members were removed: {{.users}}.",
			},
			PluralCount: len(deleted),
			TemplateData: map[string]interface{}{
				"user":  deleted[0],
				"users": strings.Join(deleted, ", "),
			},
		})
		text += deleteMembersSucceed + "\n"
	}

	return text
}

func (bot *Bot) deleteAdmins(users []string) string {
	localizer := i18n.NewLocalizer(bot.Bundle, bot.Properties.Language)

	var failed, deleted []string
	var text string

	rg, _ := regexp.Compile("<@([a-z0-9]+)|([a-z0-9]+)>")

	for _, u := range users {
		if !rg.MatchString(u) {
			failed = append(failed, u)
			continue
		}
		userID, _ := utils.SplitUser(u)
		user, err := bot.DB.SelectUser(userID)
		if err != nil {
			failed = append(failed, u)
			continue
		}
		if user.Role != "admin" {
			failed = append(failed, u)
			continue
		}
		user.Role = ""
		bot.DB.UpdateUser(user)
		adminRemoved := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "PMRemoved",
				Description: "Displays message when user removed as admin from Comedian",
				Other:       "You have been removed as Admin from Comedian",
			},
		})

		err = bot.SendUserMessage(userID, adminRemoved)
		if err != nil {
			logrus.Errorf("rest: SendUserMessage failed: %v\n", err)
		}
		deleted = append(deleted, u)
	}

	if len(failed) != 0 {
		deleteAdminsFailed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteAdminsFailed",
				Description: "Diplays message when admin deletion errors occur",
				One:         "Could not remove user as admin: {{.admin}}. User does not exist.",
				Other:       "Could not remove users as admins: {{.admins}}. Users do not exist.",
			},
			PluralCount: len(failed),
			TemplateData: map[string]interface{}{
				"admin":  failed[0],
				"admins": strings.Join(failed, ", "),
			},
		})
		text += deleteAdminsFailed + "\n"
	}
	if len(deleted) != 0 {
		deleteAdminsSucceed := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          "DeleteAdminsSucceed",
				Description: "Diplays message when admins have been successfully deleted",
				One:         "User {{.admin}} removed as admin.",
				Other:       "Users were removed as admins: {{.admins}}",
			},
			PluralCount: len(deleted),
			TemplateData: map[string]interface{}{
				"admin":  deleted[0],
				"admins": strings.Join(deleted, ", "),
			},
		})
		text += deleteAdminsSucceed + "\n"
	}

	return text
}
