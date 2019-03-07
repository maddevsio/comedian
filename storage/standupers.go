package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/model"
)

// CreateStanduper creates comedian entry in database
func (m *MySQL) CreateStanduper(s model.Standuper) (model.Standuper, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.conn.Exec(
		"INSERT INTO `channel_members` (team_id, user_id, channel_id, submitted_standup_today, role_in_channel, created) VALUES (?,?,?,?,?,?)",
		s.TeamID, s.UserID, s.ChannelID, true, s.RoleInChannel, time.Now())
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
func (m *MySQL) UpdateStanduper(st model.Standuper) (model.Standuper, error) {
	_, err := m.conn.Exec(
		"UPDATE `channel_members` SET submitted_standup_today=?, role_in_channel=? WHERE id=?",
		st.SubmittedStandupToday, st.RoleInChannel, st.ID,
	)
	if err != nil {
		return st, err
	}
	var i model.Standuper
	err = m.conn.Get(&i, "SELECT * FROM `channel_members` WHERE id=?", st.ID)
	return i, err
}

//FindStansuperByUserID finds user in channel
func (m *MySQL) FindStansuperByUserID(userID, channelID string) (model.Standuper, error) {
	var u model.Standuper
	err := m.conn.Get(&u, "SELECT * FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	return u, err
}

// ListStandupers returns array of standup entries from database
func (m *MySQL) ListStandupers() ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members`")
	return items, err
}

//GetStanduper returns a standuper
func (m *MySQL) GetStanduper(id int64) (model.Standuper, error) {
	standuper := model.Standuper{}
	err := m.conn.Get(&standuper, "SELECT * FROM channel_members where id=?", id)
	return standuper, err
}

// ListChannelStandupers returns array of standup entries from database
func (m *MySQL) ListChannelStandupers(channelID string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members` WHERE channel_id=?", channelID)
	return items, err
}

// DeleteStanduper deletes channel_members entry from database
func (m *MySQL) DeleteStanduper(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `channel_members` WHERE id=?", id)
	return err
}
