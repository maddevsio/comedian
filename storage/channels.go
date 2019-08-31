package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateProject creates standup entry in database
func (m *DB) CreateProject(ch model.Project) (model.Project, error) {
	err := ch.Validate()
	if err != nil {
		return ch, err
	}

	res, err := m.db.Exec(
		`INSERT INTO projects (
			created_at,
			workspace_id, 
			channel_name, 
			channel_id, 
			deadline,
			tz,
			onbording_message,
			submission_days
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ch.CreatedAt,
		ch.WorkspaceID,
		ch.ChannelName,
		ch.ChannelID,
		ch.Deadline,
		ch.TZ,
		ch.OnbordingMessage,
		ch.SubmissionDays,
	)
	if err != nil {
		return ch, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ch, err
	}
	ch.ID = id

	return ch, nil
}

// UpdateProject updates Project entry in database
func (m *DB) UpdateProject(ch model.Project) (model.Project, error) {
	err := ch.Validate()
	if err != nil {
		return ch, err
	}
	_, err = m.db.Exec(
		`UPDATE projects SET 
		deadline=?,
		tz=?,
		onbording_message=?,
		submission_days=? 
		WHERE id=?`,
		ch.Deadline,
		ch.TZ,
		ch.OnbordingMessage,
		ch.SubmissionDays,
		ch.ID,
	)
	if err != nil {
		return ch, err
	}
	return ch, nil
}

//ListProjects returns list of projects
func (m *DB) ListProjects() ([]model.Project, error) {
	projects := []model.Project{}
	err := m.db.Select(&projects, "SELECT * FROM `projects`")
	return projects, err
}

//ListWorkspaceProjects returns list of projects
func (m *DB) ListWorkspaceProjects(ws string) ([]model.Project, error) {
	projects := []model.Project{}
	err := m.db.Select(&projects, "SELECT * FROM `projects` where workspace_id=?", ws)
	return projects, err
}

// SelectProject selects Project entry from database
func (m *DB) SelectProject(channelID string) (model.Project, error) {
	var c model.Project
	err := m.db.Get(&c, "SELECT * FROM `projects` WHERE channel_id=?", channelID)
	if err != nil {
		return c, err
	}
	return c, err
}

// GetProject selects Project entry from database with specific id
func (m *DB) GetProject(id int64) (model.Project, error) {
	var c model.Project
	err := m.db.Get(&c, "SELECT * FROM `projects` where id=?", id)
	if err != nil {
		return c, err
	}
	return c, err
}

// DeleteProject deletes Project entry from database
func (m *DB) DeleteProject(id int64) error {
	_, err := m.db.Exec("DELETE FROM `projects` WHERE id=?", id)
	return err
}
