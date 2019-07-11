package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestCreateBotSettings(t *testing.T) {
	db := setupDB()

	bot, err := db.CreateBotSettings("", "", "", "")
	assert.Error(t, err)
	assert.Equal(t, int64(0), bot.ID)

	bot, err = db.CreateBotSettings("token", "userID", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	assert.NoError(t, db.DeleteBotSettingsByID(bot.ID))
}

func TestBotSettings(t *testing.T) {
	db := setupDB()

	_, err := db.GetAllBotSettings()
	assert.NoError(t, err)

	bot, err := db.CreateBotSettings("token", "", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	bot, err = db.GetBotSettings(bot.ID)
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = db.GetBotSettings(int64(0))
	assert.Error(t, err)

	bot, err = db.GetBotSettingsByTeamName("foo")
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = db.GetBotSettingsByTeamName("bar")
	assert.Error(t, err)

	assert.NoError(t, db.DeleteBotSettingsByID(bot.ID))
}

func TestUpdateAndDeleteBotSettings(t *testing.T) {
	db := setupDB()

	bot, err := db.CreateBotSettings("token", "", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)
	assert.Equal(t, "en_US", bot.Language)

	bot.Language = "ru_RU"

	bot, err = db.UpdateBotSettings(bot)
	assert.NoError(t, err)
	assert.Equal(t, "ru_RU", bot.Language)

	assert.NoError(t, db.DeleteBotSettings(bot.TeamID))
}
