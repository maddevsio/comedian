package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/model"
)

// CreateStandup creates standup entry in database
func (m *DB) CreateStandup(s model.Standup) (model.Standup, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.DB.Exec(
		"INSERT INTO `standups` (team_id, created, modified, comment, channel_id, user_id, message_ts) VALUES (?,?, ?, ?, ?, ?, ?)",
		s.TeamID, time.Now().UTC(), time.Now().UTC(), s.Comment, s.ChannelID, s.UserID, s.MessageTS,
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
	_, err = m.DB.Exec(
		"UPDATE `standups` SET modified=?, comment=?, message_ts=? WHERE id=?",
		time.Now().UTC(), s.Comment, s.MessageTS, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.DB.Get(&i, "SELECT * FROM `standups` WHERE id=?", s.ID)
	return i, err
}

//GetStandup returns standup by its ID
func (m *DB) GetStandup(id int64) (model.Standup, error) {
	var s model.Standup
	err := m.DB.Get(&s, "SELECT * FROM `standups` WHERE id=?", id)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
func (m *DB) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.DB.Get(&s, "SELECT * FROM `standups` WHERE message_ts=?", messageTS)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectLatestStandupByUser selects standup entry from database filtered by user
func (m *DB) SelectLatestStandupByUser(userID, channelID string) (model.Standup, error) {
	var s model.Standup
	err := m.DB.Get(&s, "select * from standups where user_id=? and channel_id=? order by id desc limit 1", userID, channelID)
	if err != nil {
		return s, err
	}
	return s, nil
}

// GetStandupForPeriod selects standup entry from database filtered by user
func (m *DB) GetStandupForPeriod(userID, channelID string, timeFrom, timeTo time.Time) (*model.Standup, error) {
	s := &model.Standup{}
	err := m.DB.Get(s, "select * from standups where user_id=? and channel_id=? and created BETWEEN ? AND ? limit 1", userID, channelID, timeFrom, timeTo)
	if err != nil {
		return s, err
	}
	return s, nil
}

// ListStandups returns array of standup entries from database
func (m *DB) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.DB.Select(&items, "SELECT * FROM `standups`")
	return items, err
}

// DeleteStandup deletes standup entry from database
func (m *DB) DeleteStandup(id int64) error {
	_, err := m.DB.Exec("DELETE FROM `standups` WHERE id=?", id)
	return err
}
