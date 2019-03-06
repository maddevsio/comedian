package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/model"
)

// CreateStandup creates standup entry in database
func (m *MySQL) CreateStandup(s model.Standup) (model.Standup, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.conn.Exec(
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
func (m *MySQL) UpdateStandup(s model.Standup) (model.Standup, error) {
	_, err := m.conn.Exec(
		"UPDATE `standups` SET team_id=?, modified=?, comment=?, message_ts=? WHERE id=?",
		s.TeamID, time.Now().UTC(), s.Comment, s.MessageTS, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.conn.Get(&i, "SELECT * FROM `standups` WHERE id=?", s.ID)
	return i, err
}

func (m *MySQL) GetStandup(id int64) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standups` WHERE id=?", id)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
func (m *MySQL) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standups` WHERE message_ts=?", messageTS)
	if err != nil {
		return s, err
	}
	return s, nil
}

// SelectStandupsByChannelIDForPeriod selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsByChannelIDForPeriod(channelID string, dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standups` WHERE channel_id=? AND created BETWEEN ? AND ?",
		channelID, dateStart, dateEnd)
	return items, err
}

// SelectStandupsFiltered selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsFiltered(userID, channelID string, dateStart, dateEnd time.Time) (model.Standup, error) {
	items := model.Standup{}
	err := m.conn.Get(&items, "SELECT * FROM `standups` WHERE channel_id=? AND user_id =? AND created BETWEEN ? AND ? limit 1",
		channelID, userID, dateStart, dateEnd)
	return items, err
}

// ListStandups returns array of standup entries from database
func (m *MySQL) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standups`")
	return items, err
}

// DeleteStandup deletes standup entry from database
func (m *MySQL) DeleteStandup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standups` WHERE id=?", id)
	return err
}
