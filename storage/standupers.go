package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateStanduper creates comedian entry in database
func (m *DB) CreateStanduper(s model.Standuper) (model.Standuper, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.db.Exec(
		`INSERT INTO standupers (
			created_at,
			workspace_id, 
			user_id, 
			channel_id, 
			role, 
			real_name, 
			channel_name
		) VALUES (?,?,?,?,?,?,?)`,
		s.CreatedAt,
		s.WorkspaceID,
		s.UserID,
		s.ChannelID,
		s.Role,
		s.RealName,
		s.ChannelName,
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

// UpdateStanduper updates Standuper entry in database
func (m *DB) UpdateStanduper(st model.Standuper) (model.Standuper, error) {
	err := st.Validate()
	if err != nil {
		return st, err
	}
	_, err = m.db.Exec(
		"UPDATE `standupers` SET role=? WHERE id=?",
		st.Role, st.ID,
	)
	if err != nil {
		return st, err
	}
	var i model.Standuper
	err = m.db.Get(&i, "SELECT * FROM `standupers` WHERE id=?", st.ID)
	return i, err
}

//FindStansuperByUserID finds user in channel
func (m *DB) FindStansuperByUserID(userID, channelID string) (model.Standuper, error) {
	var u model.Standuper
	err := m.db.Get(&u, "SELECT * FROM `standupers` WHERE user_id=? AND channel_id=?", userID, channelID)
	return u, err
}

//FindStansupersByUserID finds user in channel
func (m *DB) FindStansupersByUserID(userID string) ([]model.Standuper, error) {
	var u []model.Standuper
	err := m.db.Select(&u, "SELECT * FROM `standupers` WHERE user_id=?", userID)
	return u, err
}

// ListStandupers returns array of standup entries from database
func (m *DB) ListStandupers() ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.db.Select(&items, "SELECT * FROM `standupers`")
	return items, err
}

// ListWorkspaceStandupers returns array of standup entries from database
func (m *DB) ListWorkspaceStandupers(workspaceID string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.db.Select(&items, "SELECT * FROM `standupers` where workspace_id=?", workspaceID)
	return items, err
}

//GetStanduper returns a standuper
func (m *DB) GetStanduper(id int64) (model.Standuper, error) {
	standuper := model.Standuper{}
	err := m.db.Get(&standuper, "SELECT * FROM standupers where id=?", id)
	return standuper, err
}

// ListProjectStandupers returns array of standup entries from database
func (m *DB) ListProjectStandupers(channelID string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.db.Select(&items, "SELECT * FROM `standupers` WHERE channel_id=?", channelID)
	return items, err
}

// ListStandupersByWorkspaceID returns array of standupers which belongs to one team
func (m *DB) ListStandupersByWorkspaceID(wsID string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.db.Select(&items, "SELECT * FROM `standupers` WHERE workspace_id=?", wsID)
	return items, err
}

// DeleteStanduper deletes standupers entry from database
func (m *DB) DeleteStanduper(id int64) error {
	_, err := m.db.Exec("DELETE FROM `standupers` WHERE id=?", id)
	return err
}
