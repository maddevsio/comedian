package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestAddTimeTable(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel",
		ChannelID:   "testChannelid1",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	//adds channel member with timetable
	chanMemb, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "testChanMemb",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates timetable
	timeTable, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chanMemb.ID,
		Created:         time.Now(),
		//needs to update timetable
		Modified:  time.Now(),
		Monday:    0,
		Tuesday:   12345,
		Wednesday: 12345,
		Thursday:  0,
		Friday:    12345,
		Saturday:  0,
		Sunday:    0,
	})
	assert.NoError(t, err)
	//updates timetable
	_, err = bot.DB.UpdateTimeTable(timeTable)
	assert.NoError(t, err)
	//calculate hours and minutes from timestamp
	//there may be different values on different computers
	timeSt := time.Unix(12345, 0)
	hour := timeSt.Hour()
	minute := timeSt.Minute()
	Strhour := strconv.Itoa(hour)
	Strminute := strconv.Itoa(minute)

	testCase := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, channel.ChannelID, "", "Access Denied! You need to be at least PM in this project to use this command!"},
		//wrong parameters
		{3, channel.ChannelID, "user on mon", "To configure individual standup schedule for members use `add_timetable` command. First tag users then add keyworn *on*, after it include weekdays you want to set individual schedule (mon tue, wed, thu, fri, sat, sun) and then select time with keywork *at* (18:45). Example: `@user1 @user2 on mon tue at 14:02` \n"},
		//wrong user data
		{3, channel.ChannelID, "user on mon at 10:30", "Seems like you misspelled username. Please, check and try command again!\n"},
		//channel member with timetable
		{3, channel.ChannelID, fmt.Sprintf("<@%v|username on mon at 15:00", chanMemb.UserID), "Timetable for <@testChanMemb> updated: | Monday 15:00 | Tuesday 08:25 | Wednesday 08:25 | Friday 08:25 |\n"},
		//user isn't member in channel
		//channel member will be created
		{3, channel.ChannelID, "<@newUser|NewUserName> random on mon at 10:00", "Timetable for <@newUser> created: | Monday 10:00 |\n Seems like you misspelled username. Please, check and try command again!\n"},
	}
	for _, test := range testCase {
		actual := botAPI.addTimeTable(test.accessLevel, test.channelID, test.params)
		//replace "08:25" on calculated hour and minute from timestamp
		var expected string
		if hour < 9 {
			expected = strings.Replace(test.expected, "08:25", "0"+Strhour+":"+Strminute, -1)
		} else {
			expected = strings.Replace(test.expected, "08:25", Strhour+":"+Strminute, -1)
		}
		assert.Equal(t, expected, actual)
	}
	//delete timetable
	err = bot.DB.DeleteTimeTable(timeTable.ID)
	assert.NoError(t, err)
	//deletes channel members
	err = bot.DB.DeleteChannelMember(chanMemb.UserID, chanMemb.ChannelID)
	assert.NoError(t, err)
	//delete created channel member and his timetable
	cm, err := bot.DB.FindChannelMemberByUserID("newUser", channel.ChannelID)
	assert.NoError(t, err)
	timeT, err := bot.DB.SelectTimeTable(cm.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteTimeTable(timeT.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember("newUser", channel.ChannelID)
	assert.NoError(t, err)
	//deletes channel
	err = bot.DB.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestShowTimeTable(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel with members
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: 12345,
	})
	assert.NoError(t, err)
	//adds channel members
	chanMemb1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates timetable for chanMemb1
	timeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chanMemb1.ID,
		Created:         time.Now(),
		//needs to update timetable
		Modified:  time.Now(),
		Monday:    12345,
		Tuesday:   12345,
		Wednesday: 0,
		Thursday:  0,
		Friday:    12345,
		Saturday:  12345,
		Sunday:    0,
	})
	assert.NoError(t, err)
	//calculate hours and minutes from timestamp
	//there may be different values on different computers
	timeSt := time.Unix(12345, 0)
	hour := timeSt.Hour()
	minute := timeSt.Minute()
	Strhour := strconv.Itoa(hour)
	Strminute := strconv.Itoa(minute)
	//updates timetable
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)
	//creates channel member without timetable
	chanMemb2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)

	testCases := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{2, channel.ChannelID, fmt.Sprintf("<@%v|username1>", chanMemb1.UserID), "Timetable for <@username1> is: | Monday 08:25 | Tuesday 08:25 | Friday 08:25 | Saturday 08:25 |\n"},
		{4, channel.ChannelID, fmt.Sprintf("<@%v|username2>", chanMemb2.UserID), "<@username2> does not have a timetable!\n"},
		{2, channel.ChannelID, "<@randomid|randomName>", "Seems like <@randomName> is not even assigned as standuper in this channel!\n"},
		{4, channel.ChannelID, "wrongParameters", "Seems like you misspelled username. Please, check and try command again!\n"},
		{2, channel.ChannelID, fmt.Sprintf("<@%v|username1> <@randomid|randomName>", chanMemb1.UserID), "Timetable for <@username1> is: | Monday 08:25 | Tuesday 08:25 | Friday 08:25 | Saturday 08:25 |\nSeems like <@randomName> is not even assigned as standuper in this channel!\n"},
	}
	for _, test := range testCases {
		actual := botAPI.showTimeTable(test.accessLevel, test.channelID, test.params)
		//replace "08:25" on calculated hour and minute from timestamp
		var expected string
		if hour < 9 {
			expected = strings.Replace(test.expected, "08:25", "0"+Strhour+":"+Strminute, -1)
		} else {
			expected = strings.Replace(test.expected, "08:25", Strhour+":"+Strminute, -1)
		}
		assert.Equal(t, expected, actual)
	}
	//deletes timetables
	err = bot.DB.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	//deletes channel members
	err = bot.DB.DeleteChannelMember(chanMemb1.UserID, channel.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(chanMemb2.UserID, channel.ChannelID)
	assert.NoError(t, err)
	//deletes channel
	err = bot.DB.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestRemoveTimeTable(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	//creates channel with members
	channel, err := bot.DB.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: int64(100),
	})
	assert.NoError(t, err)
	//adds member without timetable
	chanMemb1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//adds member with timetable
	chanMemb2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//create timetable
	timeTable1, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chanMemb2.ID,
		Created:         time.Now(),
		//needs to update timetable
		Modified:  time.Now(),
		Monday:    12345,
		Tuesday:   12345,
		Wednesday: 0,
		Thursday:  0,
		Friday:    12345,
		Saturday:  12345,
		Sunday:    0,
	})
	assert.NoError(t, err)
	//updates timetable
	_, err = bot.DB.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCases := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, "", "", "Access Denied! You need to be at least PM in this project to use this command!"},
		{3, channel.ChannelID, "wrongparams", "Seems like you misspelled username. Please, check and try command again!\n"},
		//user without timetable
		{3, channel.ChannelID, fmt.Sprintf("<@%v|username1>", chanMemb1.UserID), "<@username1> does not have a timetable!\n"},
		//user with timetable
		{3, channel.ChannelID, fmt.Sprintf("<@%v|username2>", chanMemb2.UserID), "Timetable removed for <@username2>\n"},
		//user isn't member of this channel
		{3, channel.ChannelID, "<@randomUser|RandomUsername>", "Seems like <@RandomUsername> is not even assigned as standuper in this channel!\n"},
		//several parameters
		{3, channel.ChannelID, "<@uid1|username> <@randomUser|RandomUsername>", "<@username> does not have a timetable!\nSeems like <@RandomUsername> is not even assigned as standuper in this channel!\n"},
	}
	for _, test := range testCases {
		actual := botAPI.removeTimeTable(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	assert.NoError(t, bot.DB.DeleteChannel(channel.ID))
	//deletes channel members
	err = bot.DB.DeleteChannelMember(chanMemb1.UserID, chanMemb1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(chanMemb2.UserID, chanMemb2.ChannelID)
	assert.NoError(t, err)
}

func TestReturnTimeTable(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	chm1, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     "chan1",
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	timetable1, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chm1.ID,
		Monday:          12345,
		Tuesday:         12345,
		Wednesday:       12345,
		Thursday:        12345,
		Friday:          12345,
		Saturday:        12345,
		Sunday:          12345,
	})
	assert.NoError(t, err)
	_, err = bot.DB.UpdateTimeTable(timetable1)
	assert.NoError(t, err)
	//second timetable
	chm2, err := bot.DB.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     "chan1",
		RoleInChannel: "",
	})
	assert.NoError(t, err)
	timetable2, err := bot.DB.CreateTimeTable(model.TimeTable{
		ChannelMemberID: chm2.ID,
		Monday:          12345,
		Tuesday:         0,
		Wednesday:       12345,
		Thursday:        0,
		Friday:          12345,
		Saturday:        12345,
		Sunday:          0,
	})
	assert.NoError(t, err)
	_, err = bot.DB.UpdateTimeTable(timetable2)
	assert.NoError(t, err)

	//calculate hours and minutes from timestamp
	//there may be different values on different computers
	timeSt := time.Unix(12345, 0)
	hour := timeSt.Hour()
	minute := timeSt.Minute()
	Strhour := strconv.Itoa(hour)
	Strminute := strconv.Itoa(minute)

	testCase := []struct {
		timetable model.TimeTable
		expected  string
	}{
		{timetable1, "| Monday 08:25 | Tuesday 08:25 | Wednesday 08:25 | Thursday 08:25 | Friday 08:25 | Saturday 08:25 | Sunday 08:25 |"},
		{timetable2, "| Monday 08:25 | Wednesday 08:25 | Friday 08:25 | Saturday 08:25 |"},
	}
	for _, test := range testCase {
		actual := botAPI.returnTimeTable(test.timetable)
		//replace "08:25" on calculated hour and minute from timestamp
		var expected string
		if hour < 9 {
			expected = strings.Replace(test.expected, "08:25", "0"+Strhour+":"+Strminute, -1)
		} else {
			expected = strings.Replace(test.expected, "08:25", Strhour+":"+Strminute, -1)
		}
		assert.Equal(t, expected, actual)
	}

	err = bot.DB.DeleteChannelMember(chm1.UserID, chm1.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteTimeTable(timetable1.ID)
	assert.NoError(t, err)
	err = bot.DB.DeleteChannelMember(chm2.UserID, chm2.ChannelID)
	assert.NoError(t, err)
	err = bot.DB.DeleteTimeTable(timetable2.ID)
	assert.NoError(t, err)
}

func TestSplitTimeTableCommand(t *testing.T) {
	d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	botAPI, err := NewBotAPI(bot)
	assert.NoError(t, err)

	testCases := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy on friday at 01:00", "@anatoliy", "friday", int64(1514833200), ""},
		{"@anatoliy n friday ft 01:00", "", "", int64(0), "Sorry, could not understand where are the standupers and where is the rest of the command. Please, check the text for mistakes and try again"},
		{"@anatoliy on Friday at 01:00", "@anatoliy", "friday", int64(1514833200), ""},
		{"<@UB9AE7CL9|fedorenko.tolik> on monday at 01:00", "<@UB9AE7CL9|fedorenko.tolik>", "monday", int64(1514833200), ""},
		{"@anatoliy @erik @alex on friday tuesday monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1514833200), ""},
		{"@anatoliy @erik @alex on friday, tuesday, monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1514833200), ""},
	}
	for _, tt := range testCases {
		users, weekdays, _, err := botAPI.SplitTimeTableCommand(tt.command, " on ", " at ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		//assert.Equal(t, tt.time, deadline)
		if err != nil {
			assert.Equal(t, errors.New(tt.err), err)
		}
	}
	bot.CP.Language = "ru_RUS"
	testCasesRus := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy по пятницам в 02:04", "@anatoliy", "пятницам", int64(1514837040), ""},
		{"@anatoliy @erik @alex по понедельникам пятницам вторникам в 23:04", "@anatoliy @erik @alex", "понедельникам пятницам вторникам", int64(1514912640), ""},
	}
	for _, tt := range testCasesRus {
		users, weekdays, _, err := botAPI.SplitTimeTableCommand(tt.command, " по ", " в ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		//assert.Equal(t, tt.time, deadline)
		if err != nil {
			assert.Equal(t, errors.New(tt.err), err)
		}
	}
}
