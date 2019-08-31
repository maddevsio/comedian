package storage

import (
	"testing"

	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateWorkspace(t *testing.T) {
	bot, err := db.CreateWorkspace(model.Workspace{})
	assert.Error(t, err)
	assert.Equal(t, int64(0), bot.ID)

	bs := model.Workspace{
		NotifierInterval:       30,
		Language:               "en_US",
		MaxReminders:           3,
		ReminderOffset:         int64(10),
		BotAccessToken:         "token",
		BotUserID:              "userID",
		WorkspaceID:            "WorkspaceID",
		WorkspaceName:          "foo",
		ReportingChannel:       "",
		ReportingTime:          "9:00",
		ProjectsReportsEnabled: false,
	}

	bot, err = db.CreateWorkspace(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.WorkspaceName)

	assert.NoError(t, db.DeleteWorkspaceByID(bot.ID))
}

func TestWorkspace(t *testing.T) {

	_, err := db.GetAllWorkspaces()
	assert.NoError(t, err)

	bs := model.Workspace{
		NotifierInterval:       30,
		Language:               "en_US",
		MaxReminders:           3,
		ReminderOffset:         int64(10),
		BotAccessToken:         "token",
		BotUserID:              "userID",
		WorkspaceID:            "WorkspaceID",
		WorkspaceName:          "foo",
		ReportingChannel:       "",
		ReportingTime:          "9:00",
		ProjectsReportsEnabled: false,
	}

	bot, err := db.CreateWorkspace(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.WorkspaceName)

	bot, err = db.GetWorkspace(bot.ID)
	assert.NoError(t, err)
	assert.Equal(t, "WorkspaceID", bot.WorkspaceID)

	bot, err = db.GetWorkspace(int64(0))
	assert.Error(t, err)

	bot, err = db.GetWorkspaceByWorkspaceID("WorkspaceID")
	assert.NoError(t, err)
	assert.Equal(t, "WorkspaceID", bot.WorkspaceID)

	bot, err = db.GetWorkspaceByWorkspaceID("teamWrongID")
	assert.Error(t, err)

	assert.NoError(t, db.DeleteWorkspaceByID(bot.ID))
}

func TestUpdateAndDeleteWorkspace(t *testing.T) {
	bs := model.Workspace{
		NotifierInterval:       30,
		Language:               "en_US",
		MaxReminders:           3,
		ReminderOffset:         int64(10),
		BotAccessToken:         "token",
		BotUserID:              "userID",
		WorkspaceID:            "WorkspaceID",
		WorkspaceName:          "foo",
		ReportingChannel:       "",
		ReportingTime:          "9:00",
		ProjectsReportsEnabled: false,
	}

	bot, err := db.CreateWorkspace(bs)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bot.WorkspaceName)
	assert.Equal(t, "en_US", bot.Language)

	bot.Language = "ru_RU"

	bot, err = db.UpdateWorkspace(bot)
	assert.NoError(t, err)
	assert.Equal(t, "ru_RU", bot.Language)

	assert.NoError(t, db.DeleteWorkspace(bot.WorkspaceID))
}
