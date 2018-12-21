package api

import "strings"

//HelpText struct contains info for help text
type HelpText struct {
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

func DisplayHelpText(command string) string {

	ht := generateHelpText()

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

func generateHelpText() HelpText {
	ht := HelpText{
		AddMembers:            "To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, defaulting to developer if no role selected! \n",
		ShowMembers:           "To view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_ \n",
		RemoveMembers:         "To remove members use `remove` command. If you provide a role name, you will remove members with this role. _admin, pm, developer, designer_ \n",
		AddDeadline:           "To set standup deadline use `add_deadline` command. You need to provide it with hours and minutes in the 24hour format like 13:54 \n",
		ShowDeadline:          "To view standup deadline in the channel use `show_deadline` command \n",
		RemoveDeadline:        "To remove standup deadline in the channel use `remove_deadline` command \n",
		AddTimetable:          "To configure individual standup schedule for members use `add_timetable` command. First tag users then add keyworn *on(по)*, after it include weekdays you want to set individual schedule (mon tue, wed, thu, fri, sat, sun) or their Rus verstions (пн, вт, ср, чт, пт, сб, вс) and then select time with keywork at(в) (18:45). Example: `@user1 @user2 on mon tue at 14:02`  \n",
		ShowTimetable:         "To view individual standup schedule for members use `show_timetable` command and tag members. Example: `show_timetable @user1 @user2`  \n",
		RemoveTimetable:       "To remove individual standup schedule for members use `remove_timetable` command and tag members. Example: `remove_timetable @user1 @user2`  \n",
		ReportOnUser:          "To view standup report on user use `report_on_user` command. Tag user, then insert date from and date to you want to view your report. Example: `report_on_user @user 2017-01-01 2017-01-31` \n",
		ReportOnProject:       "To view standup for project use `report_on_project` command. Tag project, then insert date from and date to you want to view your report. Example: `report_on_project #projectName 2017-01-01 2017-01-31` \n",
		ReportOnUserInProject: "To view standup for user in project use `report_on_user_in_project` command. Tag user and then project, then insert date from and date to you want to view your report. Example: `report_on_user_in_project @user #projectName 2017-01-01 2017-01-31`\n",
	}
	return ht
}

func (ht HelpText) showAllHelp() string {
	textHead := "Below you will see examples of how to use Comedian slash commands: \n"
	textBody := ht.AddMembers + ht.ShowMembers + ht.RemoveMembers + ht.AddDeadline + ht.ShowDeadline + ht.RemoveDeadline + ht.AddTimetable + ht.RemoveTimetable + ht.ShowTimetable + ht.ReportOnUser + ht.ReportOnProject + ht.ReportOnUserInProject

	return textHead + textBody
}
