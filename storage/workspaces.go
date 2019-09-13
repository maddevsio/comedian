package storage

import (
	"github.com/maddevsio/comedian/model"
)

//CreateWorkspace creates bot properties for the newly created bot
func (m *DB) CreateWorkspace(bs model.Workspace) (model.Workspace, error) {
	err := bs.Validate()
	if err != nil {
		return bs, err
	}

	res, err := m.db.Exec(
		`INSERT INTO workspaces (
			created_at,
			notifier_interval, 
			max_reminders, 
			reminder_offset,
			workspace_id, 
			workspace_name, 
			bot_access_token, 
			bot_user_id, 
			projects_reports_enabled, 
			reporting_channel, 
			reporting_time, 
			language
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bs.CreatedAt,
		bs.NotifierInterval,
		bs.MaxReminders,
		bs.ReminderOffset,
		bs.WorkspaceID,
		bs.WorkspaceName,
		bs.BotAccessToken,
		bs.BotUserID,
		bs.ProjectsReportsEnabled,
		bs.ReportingChannel,
		bs.ReportingTime,
		bs.Language,
	)
	if err != nil {
		return bs, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return bs, err
	}

	bs.ID = id

	return bs, nil
}

//UpdateWorkspace updates bot
func (m *DB) UpdateWorkspace(settings model.Workspace) (model.Workspace, error) {
	err := settings.Validate()
	if err != nil {
		return settings, err
	}

	_, err = m.db.Exec(
		`UPDATE workspaces set 
			notifier_interval=?, 
			max_reminders=?, 
			workspace_id=?, 
			reminder_offset=?,
			workspace_name=?, 
			bot_access_token=?, 
			bot_user_id=?, 
			projects_reports_enabled=?, 
			reporting_channel=?, 
			reporting_time=?, 
			language=?
			where id=?`,
		settings.NotifierInterval,
		settings.MaxReminders,
		settings.WorkspaceID,
		settings.ReminderOffset,
		settings.WorkspaceName,
		settings.BotAccessToken,
		settings.BotUserID,
		settings.ProjectsReportsEnabled,
		settings.ReportingChannel,
		settings.ReportingTime,
		settings.Language,
		settings.ID,
	)
	if err != nil {
		return settings, err
	}
	return settings, nil
}

//GetAllWorkspaces returns all workspaces stored in DB
func (m *DB) GetAllWorkspaces() ([]model.Workspace, error) {
	bs := []model.Workspace{}
	err := m.db.Select(&bs, "SELECT * FROM `workspaces`")
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetWorkspaceByWorkspaceID returns a particular bot
func (m *DB) GetWorkspaceByWorkspaceID(workspaceID string) (model.Workspace, error) {
	bs := model.Workspace{}
	err := m.db.Get(&bs, "SELECT * FROM `workspaces` where workspace_id=?", workspaceID)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetWorkspaceByBotAccessToken returns a particular bot
func (m *DB) GetWorkspaceByBotAccessToken(botAccessToken string) (model.Workspace, error) {
	bs := model.Workspace{}
	err := m.db.Get(&bs, "SELECT * FROM `workspaces` where bot_access_token=?", botAccessToken)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetWorkspace returns a particular bot
func (m *DB) GetWorkspace(id int64) (model.Workspace, error) {
	bs := model.Workspace{}
	err := m.db.Get(&bs, "SELECT * FROM `workspaces` where id=?", id)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//DeleteWorkspaceByID deletes bot
func (m *DB) DeleteWorkspaceByID(id int64) error {
	_, err := m.db.Exec("DELETE FROM `workspaces` WHERE id=?", id)
	return err
}

//DeleteWorkspace deletes bot
func (m *DB) DeleteWorkspace(teamID string) error {
	_, err := m.db.Exec("DELETE FROM `workspaces` WHERE workspace_id=?", teamID)
	return err
}
