package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateStandup creates standup entry in database
func (m *DB) CreateStandup(s model.Standup) (model.Standup, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}

	res, err := m.db.Exec(
		`INSERT INTO standups (
			created_at,
			workspace_id, 
			channel_id, 
			user_id, 
			comment, 
			message_ts
		) VALUES (?, ?, ?, ?, ?, ?)`,
		s.CreatedAt,
		s.WorkspaceID,
		s.ChannelID,
		s.UserID,
		s.Comment,
		s.MessageTS,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	s.ID = id

	return s, nil
}

// UpdateStandup updates standup entry in database
func (m *DB) UpdateStandup(s model.Standup) (model.Standup, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}

	_, err = m.db.Exec(
		"UPDATE `standups` SET comment=?, message_ts=? WHERE id=?",
		s.Comment, s.MessageTS, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.db.Get(&i, "SELECT * FROM `standups` WHERE id=?", s.ID)
	return i, err
}

// ListStandups returns array of standup entries from database
func (m *DB) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.db.Select(&items, "SELECT * FROM `standups` order by id desc")
	return items, err
}

// ListTeamStandups returns array of standup entries from database
func (m *DB) ListTeamStandups(teamID string) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.db.Select(&items, "SELECT * FROM `standups` where workspace_id=? order by id desc", teamID)
	return items, err
}

//GetStandup returns standup by its ID
func (m *DB) GetStandup(id int64) (model.Standup, error) {
	var s model.Standup
	err := m.db.Get(&s, "SELECT * FROM `standups` WHERE id=?", id)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
func (m *DB) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.db.Get(&s, "SELECT * FROM `standups` WHERE message_ts=?", messageTS)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectLatestStandupByUser selects standup entry from database filtered by user
func (m *DB) SelectLatestStandupByUser(userID, channelID string) (model.Standup, error) {
	var s model.Standup
	err := m.db.Get(&s,
		`select * from standups 
		where user_id=? and channel_id=? 
		order by id desc limit 1`,
		userID, channelID,
	)
	if err != nil {
		return s, err
	}
	return s, nil
}

// GetStandupForPeriod selects standup entry from database filtered by user
func (m *DB) GetStandupForPeriod(userID, channelID string, timeFrom, timeTo int64) (*model.Standup, error) {
	s := &model.Standup{}
	err := m.db.Get(s,
		`select * from standups 
		where user_id=? and channel_id=? 
		and created_at BETWEEN ? AND ? 
		limit 1`,
		userID,
		channelID,
		timeFrom,
		timeTo,
	)
	if err != nil {
		return s, err
	}
	return s, nil
}

// DeleteStandup deletes standup entry from database
func (m *DB) DeleteStandup(id int64) error {
	_, err := m.db.Exec("DELETE FROM `standups` WHERE id=?", id)
	return err
}
