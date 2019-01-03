package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
)

func TestDisplayHelpText(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	testCase := []struct {
		command  string
		expected string
	}{
		{"add", "To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, defaulting to developer if no role selected! \n"},
		{"show", "To view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_ \n"},
		{"remove", "To remove members use `remove` command. If you provide a role name, you will remove members with this role. _admin, pm, developer, designer_ \n"},
		{"add_deadline", "To set standup deadline use `add_deadline` command. You need to provide it with hours and minutes in the 24hour format like 13:54 \n"},
		{"show_deadline", "To view standup deadline in the channel use `show_deadline` command \n"},
		{"remove_deadline", "To remove standup deadline in the channel use `remove_deadline` command \n"},
		{"add_timetable", "To configure individual standup schedule for members use `add_timetable` command. First tag users then add keyworn *on*, after it include weekdays you want to set individual schedule (mon tue, wed, thu, fri, sat, sun) and then select time with keywork *at* (18:45). Example: `@user1 @user2 on mon tue at 14:02` \n"},
		{"show_timetable", "To view individual standup schedule for members use `show_timetable` command and tag members. Example: `show_timetable @user1 @user2` \n"},
		{"remove_timetable", "To remove individual standup schedule for members use `remove_timetable` command and tag members. Example: `remove_timetable @user1 @user2`  \n"},
		{"report_on_user", "To view standup report on user use `report_on_user` command. Tag user, then insert date from and date to you want to view your report. Example: `report_on_user @user 2017-01-01 2017-01-31` \n"},
		{"report_on_project", "To view standup for project use `report_on_project` command. Tag project, then insert date from and date to you want to view your report. Example: `report_on_project #projectName 2017-01-01 2017-01-31` \n"},
		{"report_on_user_in_project", "To view standup for user in project use `report_on_user_in_project` command. Tag user and then project, then insert date from and date to you want to view your report. Example: `report_on_user_in_project @user #projectName 2017-01-01 2017-01-31`\n"},
		{"", "Below you will see examples of how to use Comedian slash commands: \nTo add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, defaulting to developer if no role selected! \nTo view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_ \nTo remove members use `remove` command. If you provide a role name, you will remove members with this role. _admin, pm, developer, designer_ \nTo set standup deadline use `add_deadline` command. You need to provide it with hours and minutes in the 24hour format like 13:54 \nTo view standup deadline in the channel use `show_deadline` command \nTo remove standup deadline in the channel use `remove_deadline` command \nTo configure individual standup schedule for members use `add_timetable` command. First tag users then add keyworn *on*, after it include weekdays you want to set individual schedule (mon tue, wed, thu, fri, sat, sun) and then select time with keywork *at* (18:45). Example: `@user1 @user2 on mon tue at 14:02` \nTo remove individual standup schedule for members use `remove_timetable` command and tag members. Example: `remove_timetable @user1 @user2`  \nTo view individual standup schedule for members use `show_timetable` command and tag members. Example: `show_timetable @user1 @user2` \nTo view standup report on user use `report_on_user` command. Tag user, then insert date from and date to you want to view your report. Example: `report_on_user @user 2017-01-01 2017-01-31` \nTo view standup for project use `report_on_project` command. Tag project, then insert date from and date to you want to view your report. Example: `report_on_project #projectName 2017-01-01 2017-01-31` \nTo view standup for user in project use `report_on_user_in_project` command. Tag user and then project, then insert date from and date to you want to view your report. Example: `report_on_user_in_project @user #projectName 2017-01-01 2017-01-31`\n"},
	}
	for _, test := range testCase {
		actual := botAPI.DisplayHelpText(test.command)
		assert.Equal(t, test.expected, actual)
	}
}
