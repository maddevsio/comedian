package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
)

// MySQL provides api for work with mysql database
type MySQL struct {
	conn *sqlx.DB
}

// NewMySQL creates a new instance of database API
func NewMySQL(c config.Config) (*MySQL, error) {
	conn, err := sqlx.Open("mysql", c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	m := &MySQL{}
	m.conn = conn
	return m, nil
}

// CreateStandup creates standup entry in database
func (m *MySQL) CreateStandup(s model.Standup) (model.Standup, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standup` (created, modified, username, comment, channel, channel_id, username_id, full_name, message_ts) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), s.Username, s.Comment, s.Channel, s.ChannelID, s.UsernameID, s.FullName, s.MessageTS,
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
		"UPDATE `standup` SET modified=?, username=?, username_id=?, comment=?, channel=?, channel_id=?, full_name=?, message_ts=? WHERE id=?",
		time.Now().UTC(), s.Username, s.UsernameID, s.Comment, s.Channel, s.ChannelID, s.FullName, s.MessageTS, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.conn.Get(&i, "SELECT * FROM `standup` WHERE id=?", s.ID)
	return i, err
}

// SelectStandup selects standup entry from database
func (m *MySQL) SelectStandup(id int64) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standup` WHERE id=?", id)
	return s, err
}

// SelectStandupByMessageTS selects standup entry from database
func (m *MySQL) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standup` WHERE message_ts=?", messageTS)
	return s, err
}

// DeleteStandup deletes standup entry from database
func (m *MySQL) DeleteStandup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standup` WHERE id=?", id)
	return err
}

// ListStandups returns array of standup entries from database
func (m *MySQL) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup`")
	return items, err
}

// CreateStandupUser creates comedian entry in database
func (m *MySQL) CreateStandupUser(c model.StandupUser) (model.StandupUser, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standup_users` (created, modified, username, full_name, channel_id, channel) VALUES (?, ?, ?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), c.SlackName, c.FullName, c.ChannelID, c.Channel)
	if err != nil {
		return c, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return c, err
	}
	c.ID = id
	return c, nil
}

// DeleteStandupUser deletes standup_users entry from database
func (m *MySQL) DeleteStandupUserByUsername(username, channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_users` WHERE username=? AND channel_id=?", username, channelID)
	return err
}

// ListStandupUsers returns array of standup entries from database
func (m *MySQL) ListStandupUsers(channelID string) ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` WHERE channel_id=?", channelID)
	return items, err
}

// CreateStandupTime creates time entry in database
func (m *MySQL) CreateStandupTime(c model.StandupTime) (model.StandupTime, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standup_time` (created, channel_id, channel, standuptime) VALUES (?, ?, ?, ?)",
		time.Now().UTC(), c.ChannelID, c.Channel, c.Time)
	if err != nil {
		return c, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return c, err
	}
	c.ID = id
	return c, nil
}

// DeleteStandupTime deletes standup_time entry for channel from database
func (m *MySQL) DeleteStandupTime(channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_time` WHERE channel_id=?", channelID)
	return err
}

// ListStandupTime returns standup time entry from database
func (m *MySQL) ListStandupTime(channelID string) (model.StandupTime, error) {
	var time model.StandupTime
	err := m.conn.Get(&time, "SELECT * FROM `standup_time` WHERE channel_id=?", channelID)
	return time, err
}

//AddToStandupHistory creates backup standup entry in standup_edit_history database
func (m *MySQL) AddToStandupHistory(s model.StandupEditHistory) (model.StandupEditHistory, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `standup_edit_history` (created, standup_id, standup_text) VALUES (?, ?, ?)",
		time.Now().UTC(), s.StandupID, s.StandupText)
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
