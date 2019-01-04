package api

import (
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//HelpText struct contains info for help text
type HelpText struct {
	TextHead              string
	AddMembers            string
	ShowMembers           string
	RemoveMembers         string
	AddDeadline           string
	RemoveDeadline        string
	ShowDeadline          string
	AddTimetable          string
	RemoveTimetable       string
	ShowTimetable         string
	ReportOnUser          string
	ReportOnProject       string
	ReportOnUserInProject string
}

//DisplayHelpText displays help text
func (ba *BotAPI) DisplayHelpText(command string) string {

	ht := ba.generateHelpText()

	switch strings.ToLower(command) {
	case "add":
		return ht.AddMembers
	case "show":
		return ht.ShowMembers
	case "remove":
		return ht.RemoveMembers
	case "add_deadline":
		return ht.AddDeadline
	case "show_deadline":
		return ht.ShowDeadline
	case "remove_deadline":
		return ht.RemoveDeadline
	case "add_timetable":
		return ht.AddTimetable
	case "show_timetable":
		return ht.ShowTimetable
	case "remove_timetable":
		return ht.RemoveTimetable
	case "report_on_user":
		return ht.ReportOnUser
	case "report_on_project":
		return ht.ReportOnProject
	case "report_on_user_in_project":
		return ht.ReportOnUserInProject
	default:
		return ht.showAllHelp()
	}
}

func (ba *BotAPI) generateHelpText() HelpText {
	localizer := i18n.NewLocalizer(ba.Bot.Bundle, ba.Bot.CP.Language)

	textHead := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "TextHead",
			Description: "Displays help text head",
			Other:       "Below you will see examples of how to use Comedian slash commands: ",
		},
	})
	addMembers := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AddMembers",
			Description: "Displays usage of add command",
			Other:       "To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, default is a developer role, if the role is not selected! ",
		},
	})
	showMembers := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ShowMembers",
			Description: "Displays usage of show command",
			Other:       "To view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_ ",
		},
	})
	removeMembers := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RemoveMembers",
			Description: "Displays usage of remove command",
			Other:       "To remove members use `remove` command. If you provide a role name, you will remove members with this role. _admin, pm, developer, designer_ ",
		},
	})
	addDeadline := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AddDeadline",
			Description: "Displays usage of add_deadline command",
			Other:       "To set standup deadline use `add_deadline` command. You need to provide it with hours and minutes in the 24hour format like 13:54 ",
		},
	})
	showDeadline := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ShowDeadline",
			Description: "Displays usage of show_deadline command",
			Other:       "To view standup deadline in the channel use `show_deadline` command ",
		},
	})
	removeDeadline := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RemoveDeadline",
			Description: "Displays usage of show_deadline command",
			Other:       "To remove standup deadline in the channel use `remove_deadline` command ",
		},
	})
	addTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "AddTimetable",
			Description: "Displays usage of add_timetable command",
			Other:       "To configure individual standup schedule for members use `add_timetable` command. First tag users then add keyworn *on*, after it include weekdays you want to set individual schedule (mon tue, wed, thu, fri, sat, sun) and then select time with keywork *at* (18:45). Example: `@user1 @user2 on mon tue at 14:02` ",
		},
	})
	showTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ShowTimetable",
			Description: "Displays usage of add_timetable command",
			Other:       "To view individual standup schedule for members use `show_timetable` command and tag members. Example: `show_timetable @user1 @user2` ",
		},
	})
	removeTimetable := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "RemoveTimetable",
			Description: "Displays usage of add_timetable command",
			Other:       "To remove individual standup schedule for members use `remove_timetable` command and tag members. Example: `remove_timetable @user1 @user2`  ",
		},
	})
	reportOnUser := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnUser",
			Description: "Displays usage of add_timetable command",
			Other:       "To view standup report on user use `report_on_user` command. Tag user, then insert date from and date to you want to view your report. Example: `report_on_user @user 2017-01-01 2017-01-31` ",
		},
	})
	reportOnProject := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnProject",
			Description: "Displays usage of add_timetable command",
			Other:       "To view standup for project use `report_on_project` command. Tag project, then insert date from and date to you want to view your report. Example: `report_on_project #projectName 2017-01-01 2017-01-31` ",
		},
	})
	reportOnUserInProject := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "ReportOnUserInProject",
			Description: "Displays usage of add_timetable command",
			Other:       "To view standup for user in project use `report_on_user_in_project` command. Tag user and then project, then insert date from and date to you want to view your report. Example: `report_on_user_in_project @user #projectName 2017-01-01 2017-01-31`",
		},
	})

	ht := HelpText{
		TextHead:              textHead + "\n",
		AddMembers:            addMembers + "\n",
		ShowMembers:           showMembers + "\n",
		RemoveMembers:         removeMembers + "\n",
		AddDeadline:           addDeadline + "\n",
		ShowDeadline:          showDeadline + "\n",
		RemoveDeadline:        removeDeadline + "\n",
		AddTimetable:          addTimetable + "\n",
		ShowTimetable:         showTimetable + "\n",
		RemoveTimetable:       removeTimetable + "\n",
		ReportOnUser:          reportOnUser + "\n",
		ReportOnProject:       reportOnProject + "\n",
		ReportOnUserInProject: reportOnUserInProject + "\n",
	}
	return ht
}

func (ht HelpText) showAllHelp() string {
	textBody := ht.AddMembers + ht.ShowMembers + ht.RemoveMembers + ht.AddDeadline + ht.ShowDeadline + ht.RemoveDeadline + ht.AddTimetable + ht.RemoveTimetable + ht.ShowTimetable + ht.ReportOnUser + ht.ReportOnProject + ht.ReportOnUserInProject

	return ht.TextHead + textBody
}
