package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
)

func TestCreateBotSettings(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	bot, err := mysql.CreateBotSettings("", "", "")
	assert.Error(t, err)
	assert.Equal(t, int64(0), bot.ID)

	bot, err = mysql.CreateBotSettings("token", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	assert.NoError(t, mysql.DeleteBotSettingsByID(bot.ID))
}

func TestBotSettings(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	_, err = mysql.GetAllBotSettings()
	assert.NoError(t, err)

	bot, err := mysql.CreateBotSettings("token", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)

	bot, err = mysql.GetBotSettings(bot.ID)
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = mysql.GetBotSettings(int64(0))
	assert.Error(t, err)

	bot, err = mysql.GetBotSettingsByTeamName("foo")
	assert.NoError(t, err)
	assert.Equal(t, "teamID", bot.TeamID)

	bot, err = mysql.GetBotSettingsByTeamName("bar")
	assert.Error(t, err)

	assert.NoError(t, mysql.DeleteBotSettingsByID(bot.ID))
}

func TestUpdateAndDeleteBotSettings(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	bot, err := mysql.CreateBotSettings("token", "teamID", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.TeamName)
	assert.Equal(t, "en_US", bot.Language)

	bot.Language = "ru_RU"

	bot, err = mysql.UpdateBotSettings(bot)
	assert.NoError(t, err)
	assert.Equal(t, "ru_RU", bot.Language)

	assert.NoError(t, mysql.DeleteBotSettings(bot.TeamID))
}
