package api

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestAddTimeTable(t *testing.T) {
	r := SetUp()

	//creates channel
	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel",
		ChannelID:   "testChannelid1",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	//adds channel member with timetable
	chanMemb, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "testChanMemb",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates timetable
	timeTable, err := r.db.CreateTimeTable(model.TimeTable{
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
	_, err = r.db.UpdateTimeTable(timeTable)
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
		{3, channel.ChannelID, "user on mon", "Sorry, could not understand where are the weekdays and where is the time. Please, check the text for mistakes and try again"},
		//wrong user data
		{3, channel.ChannelID, "user on mon at 10:30", "Seems like you misspelled username. Please, check and try command again!"},
		//channel member with timetable
		{3, channel.ChannelID, fmt.Sprintf("<@%v|username on mon at 15:00", chanMemb.UserID), "Timetable for <@testChanMemb> updated: | Monday 15:00 | Tuesday 08:25 | Wednesday 08:25 | Friday 08:25 | \n"},
		//user isn't member in channel
		//channel member will be created
		{3, channel.ChannelID, "<@newUser|NewUserName> random on mon at 10:00", "Timetable for <@newUser> created: | Monday 10:00 | \nSeems like you misspelled username. Please, check and try command again!"},
	}
	for _, test := range testCase {
		actual := r.addTimeTable(test.accessLevel, test.channelID, test.params)
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
	err = r.db.DeleteTimeTable(timeTable.ID)
	assert.NoError(t, err)
	//deletes channel members
	err = r.db.DeleteChannelMember(chanMemb.UserID, chanMemb.ChannelID)
	assert.NoError(t, err)
	//delete created channel member and his timetable
	cm, err := r.db.FindChannelMemberByUserID("newUser", channel.ChannelID)
	assert.NoError(t, err)
	timeT, err := r.db.SelectTimeTable(cm.ID)
	assert.NoError(t, err)
	err = r.db.DeleteTimeTable(timeT.ID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember("newUser", channel.ChannelID)
	assert.NoError(t, err)
	//deletes channel
	err = r.db.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestShowTimeTable(t *testing.T) {
	r := SetUp()
	//creates channel with members
	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: 12345,
	})
	assert.NoError(t, err)
	//adds channel members
	chanMemb1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//creates timetable for chanMemb1
	timeTable1, err := r.db.CreateTimeTable(model.TimeTable{
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
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)
	//creates channel member without timetable
	chanMemb2, err := r.db.CreateChannelMember(model.ChannelMember{
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
		{4, channel.ChannelID, "wrongParameters", "Seems like you misspelled username. Please, check and try command again!"},
		{2, channel.ChannelID, fmt.Sprintf("<@%v|username1> <@randomid|randomName>", chanMemb1.UserID), "Timetable for <@username1> is: | Monday 08:25 | Tuesday 08:25 | Friday 08:25 | Saturday 08:25 |\nSeems like <@randomName> is not even assigned as standuper in this channel!\n"},
	}
	for _, test := range testCases {
		actual := r.showTimeTable(test.accessLevel, test.channelID, test.params)
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
	err = r.db.DeleteTimeTable(timeTable1.ID)
	assert.NoError(t, err)
	//deletes channel members
	err = r.db.DeleteChannelMember(chanMemb1.UserID, channel.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(chanMemb2.UserID, channel.ChannelID)
	assert.NoError(t, err)
	//deletes channel
	err = r.db.DeleteChannel(channel.ID)
	assert.NoError(t, err)
}

func TestRemoveTimeTable(t *testing.T) {
	r := SetUp()
	//creates channel with members
	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "testChannel1",
		ChannelID:   "testChannelId1",
		StandupTime: int64(100),
	})
	assert.NoError(t, err)
	//adds member without timetable
	chanMemb1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid1",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//adds member with timetable
	chanMemb2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:        "uid2",
		ChannelID:     channel.ChannelID,
		RoleInChannel: "",
		Created:       time.Now(),
	})
	assert.NoError(t, err)
	//create timetable
	timeTable1, err := r.db.CreateTimeTable(model.TimeTable{
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
	_, err = r.db.UpdateTimeTable(timeTable1)
	assert.NoError(t, err)

	testCases := []struct {
		accessLevel int
		channelID   string
		params      string
		expected    string
	}{
		{4, "", "", "Access Denied! You need to be at least PM in this project to use this command!"},
		{3, channel.ChannelID, "wrongparams", "Seems like you misspelled username. Please, check and try command again!"},
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
		actual := r.removeTimeTable(test.accessLevel, test.channelID, test.params)
		assert.Equal(t, test.expected, actual)
	}
	//deletes channels
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
	//deletes channel members
	err = r.db.DeleteChannelMember(chanMemb1.UserID, chanMemb1.ChannelID)
	assert.NoError(t, err)
	err = r.db.DeleteChannelMember(chanMemb2.UserID, chanMemb2.ChannelID)
	assert.NoError(t, err)
}
