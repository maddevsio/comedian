package storage

import (
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateBotSettings(t *testing.T) {
	bot, err := db.CreateBotSettings(&model.BotSettings{})
	assert.Error(t, err)
	assert.Equal(t, int64(0), bot.ID)

	bs := &model.BotSettings{
		NotifierInterval:    30,
		Language:            "en_US",
		ReminderRepeatsMax:  3,
		ReminderTime:        int64(10),
		AccessToken:         "token",
		UserID:              "userID",
		TeamID:              "teamID",
		TeamName:            "foo",
		ReportingChannel:    "",
		ReportingTime:       "9:00",
		IndividualReportsOn: false,
	}

	bot, err = db.CreateBotSettings(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	assert.NoError(t, db.DeleteBotSettingsByID(bot.ID))
}

func TestBotSettings(t *testing.T) {

	_, err := db.GetAllBotSettings()
	assert.NoError(t, err)

	bs := &model.BotSettings{
		NotifierInterval:    30,
		Language:            "en_US",
		ReminderRepeatsMax:  3,
		ReminderTime:        int64(10),
		AccessToken:         "token",
		UserID:              "userID",
		TeamID:              "teamID",
		TeamName:            "foo",
		ReportingChannel:    "",
		ReportingTime:       "9:00",
		IndividualReportsOn: false,
	}

	bot, err := db.CreateBotSettings(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	bot, err = db.GetBotSettings(bot.ID)
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = db.GetBotSettings(int64(0))
	assert.Error(t, err)

	bot, err = db.GetBotSettingsByTeamID("teamID")
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = db.GetBotSettingsByTeamID("teamWrongID")
	assert.Error(t, err)

	assert.NoError(t, db.DeleteBotSettingsByID(bot.ID))
}

func TestUpdateAndDeleteBotSettings(t *testing.T) {
	bs := &model.BotSettings{
		NotifierInterval:    30,
		Language:            "en_US",
		ReminderRepeatsMax:  3,
		ReminderTime:        int64(10),
		AccessToken:         "token",
		UserID:              "userID",
		TeamID:              "teamID",
		TeamName:            "foo",
		ReportingChannel:    "",
		ReportingTime:       "9:00",
		IndividualReportsOn: false,
	}

	bot, err := db.CreateBotSettings(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)
	assert.Equal(t, "en_US", bot.Language)

	bot.Language = "ru_RU"

	bot, err = db.UpdateBotSettings(bot)
	assert.NoError(t, err)
	assert.Equal(t, "ru_RU", bot.Language)

	assert.NoError(t, db.DeleteBotSettings(bot.TeamID))
}
