package reporting

import (
	"fmt"
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

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	//First test when no data
	text, err := StandupReportByProject(db, channelID, dateFrom, dateTo)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, "Full Standup Report QWERTY123:\n\nNo data for this period", text)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Полный отчёт по стэндапам QWERTY123:\n\nНет данных за данный период", text)
	}
	//create user who did not write standup
	user1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test for no standup submitted
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)

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

	//create another user
	user2, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	text, err = StandupReportByProject(db, channelName, dateFrom, dateTo)
	assert.NoError(t, err)

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

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	dateNextText := fmt.Sprintf("%d-%02d-%02d", dateNext.Year(), dateNext.Month(), dateNext.Day())
	dateToText := fmt.Sprintf("%d-%02d-%02d", dateTo.Year(), dateTo.Month(), dateTo.Day())
	//dateFromText := fmt.Sprintf("%d-%02d-%02d", dateFrom.Year(), dateFrom.Month(), dateFrom.Day())

	user, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	text, err := StandupReportByUser(db, user, dateTo, dateFrom)
	assert.Error(t, err)
	text, err = StandupReportByUser(db, user, dateNext, dateTo)
	assert.Error(t, err)
	text, err = StandupReportByUser(db, user, dateFrom, dateNext)
	assert.Error(t, err)

	text, err = StandupReportByUser(db, user, dateFrom, dateTo)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("Full Standup Report for user <@user1>:\n\n\n\nReport from %v to %v:\n\n<@user1>: ignored standup\n", dateToText, dateNextText), text)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("Полный отчет по пользователю <@user1>:\n\n\n\nОтчет с %v по %v:\n\n<@user1>: стэндап успешно просран!\n", dateToText, dateNextText), text)
	}
	standup1, err := db.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	text, err = StandupReportByUser(db, user, dateFrom, dateTo)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("Full Standup Report for user <@user1>:\n\n\n\nReport from %v to %v:\n\nOn project: <#QWERTY123>\nmy standup\n", dateToText, dateNextText), text)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("Полный отчет по пользователю <@user1>:\n\n\n\nОтчет с %v по %v:\n\nНа проект: <#QWERTY123>\nmy standup\n", dateToText, dateNextText), text)
	}
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

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	dateNextText := fmt.Sprintf("%d-%02d-%02d", dateNext.Year(), dateNext.Month(), dateNext.Day())
	dateToText := fmt.Sprintf("%d-%02d-%02d", dateTo.Year(), dateTo.Month(), dateTo.Day())
	//dateFromText := fmt.Sprintf("%d-%02d-%02d", dateFrom.Year(), dateFrom.Month(), dateFrom.Day())

	user1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})

	text, err := StandupReportByProjectAndUser(db, channelID, user1, dateFrom, dateTo)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("Standup Report Project: QWERTY123, User: <@user1>\n\n\n\nReport from %v to %v:\n\n<@user1>: ignored standup!\n", dateToText, dateNextText), text)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("Стэндап отчет по проекту: QWERTY123, пользователь: <@user1>\n\n\n\nОтчет с %v по %v:\n\n<@user1>: стэндап успешно просран!\n", dateToText, dateNextText), text)
	}
	standup1, err := db.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	assert.NoError(t, err)

	text, err = StandupReportByProjectAndUser(db, channelID, user1, dateFrom, dateTo)
	assert.NoError(t, err)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("Standup Report Project: QWERTY123, User: <@user1>\n\n\n\nReport from %v to %v:\n\nStandup from <@user1>:\nmy standup\n", dateToText, dateNextText), text)
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("Стэндап отчет по проекту: QWERTY123, пользователь: <@user1>\n\n\n\nОтчет с %v по %v:\n\nСтэндап от <@user1>:\nmy standup\n", dateToText, dateNextText), text)
	}
	assert.NoError(t, db.DeleteStandup(standup1.ID))
	assert.NoError(t, db.DeleteStandupUserByUsername(user1.SlackName, user1.ChannelID))
}
