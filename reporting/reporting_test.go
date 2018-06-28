package reporting

import (
	"testing"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/stretchr/testify/assert"
)

func TestStandupReportByProject(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.NewMySQL(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	standupTime, err := time.Parse("15:04", "13:00")

	dateFrom, err := time.Parse("2006-01-02", "2018-06-27")
	assert.NoError(t, err)
	dateTo, err := time.Parse("2006-01-02", "2018-06-28")
	assert.NoError(t, err)

	//First test when no data
	text, err := StandupReportByProject(db, channelID, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report QWERTY123:\n\nNo data for this period", text)

	//Create standup time
	_, err = db.CreateStandupTime(model.StandupTime{
		ChannelID: channelID,
		Channel:   channelName,
		Time:      standupTime.Unix(),
	})
	assert.NoError(t, err)

	//create user who did not write standup
	user1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test for no standup submitted
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report chanName:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\n<@user1> ignored standup!\n", text)

	//create standup for user
	standup1, err := db.CreateStandup(model.Standup{
		Channel:    channelName,
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	assert.NoError(t, err)

	//test if user submitted standup success
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report chanName:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\nStandup from <@user1>:\nmy standup\n", text)

	//create another user
	user2, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report chanName:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\nStandup from <@user1>:\nmy standup\n\n<@user2> ignored standup!\n", text)

	//create standup for user2
	standup2, err := db.CreateStandup(model.Standup{
		Channel:    channelName,
		ChannelID:  channelID,
		Comment:    "user2 standup",
		UsernameID: "userID2",
		Username:   "user2",
		MessageTS:  "1234",
	})
	assert.NoError(t, err)

	//test if both users had written standups
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report chanName:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\nStandup from <@user1>:\nmy standup\n\nStandup from <@user2>:\nuser2 standup\n", text)

	assert.NoError(t, db.DeleteStandup(standup1.ID))
	assert.NoError(t, db.DeleteStandup(standup2.ID))
	assert.NoError(t, db.DeleteStandupUserByUsername(user1.SlackName, user1.ChannelID))
	assert.NoError(t, db.DeleteStandupUserByUsername(user2.SlackName, user2.ChannelID))
}

func TestStandupReportByUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.NewMySQL(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"
	user, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)
	dateFrom, err := time.Parse("2006-01-02", "2018-06-24")
	assert.NoError(t, err)
	dateTo, err := time.Parse("2006-01-02", "2018-06-28")
	assert.NoError(t, err)
	text, err := StandupReportByUser(db, user, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report for user <@user1>:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\n<@user1>: ignored standup\n", text)

	standup1, err := db.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	text, err = StandupReportByUser(db, user, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report for user <@user1>:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\nOn project: <#QWERTY123>\nmy standup\n", text)
	assert.NoError(t, db.DeleteStandup(standup1.ID))

	assert.NoError(t, db.DeleteStandupUserByUsername(user.SlackName, user.ChannelID))
}

func TestStandupReportByProjectAndUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.NewMySQL(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"
	user1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)
	dateFrom, err := time.Parse("2006-01-02", "2018-06-24")
	assert.NoError(t, err)
	dateTo, err := time.Parse("2006-01-02", "2018-06-28")
	assert.NoError(t, err)
	text, err := StandupReportByProjectAndUser(db, channelName, user1, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Standup Report Project: chanName, User: <@user1>\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\n<@user1> ignored standup!\n", text)

	user2, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)
	assert.NoError(t, db.DeleteStandupUserByUsername(user1.SlackName, user1.ChannelID))
	assert.NoError(t, db.DeleteStandupUserByUsername(user2.SlackName, user2.ChannelID))
}
