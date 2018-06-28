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
	text, err := StandupReportByProject(db, channelID, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report QWERTY123:\n\nNo data for this period", text)

	_, err = db.CreateStandupTime(model.StandupTime{
		ChannelID: channelID,
		Channel:   channelName,
		Time:      standupTime.Unix(),
	})
	assert.NoError(t, err)

	user1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		FullName:    "",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)
	assert.Equal(t, "Full Standup Report chanName:\n\n\n\nReport from 2018-06-28 to 2018-06-29:\n\n<@user1> ignored standup!\n", text)

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
