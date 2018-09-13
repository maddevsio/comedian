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
	err := s.Validate()
	if err != nil {
		return s, err
	}
	var channelName string
	err = m.conn.Get(&channelName, "SELECT channel FROM `standup_users` where channel_id =?", s.ChannelID)
	if err != nil {
		return s, err
	}

	res, err := m.conn.Exec(
		"INSERT INTO `standup` (created, modified, comment, channel, channel_id, username_id, message_ts) VALUES (?, ?, ?, ?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), s.Comment, channelName, s.ChannelID, s.UsernameID, s.MessageTS,
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
	err := s.Validate()
	if err != nil {
		return s, err
	}
	_, err = m.conn.Exec(
		"UPDATE `standup` SET modified=?, username_id=?, comment=?, channel_id=?, message_ts=? WHERE id=?",
		time.Now().UTC(), s.UsernameID, s.Comment, s.ChannelID, s.MessageTS, s.ID,
	)
	if err != nil {
		return s, err
	}
	var i model.Standup
	err = m.conn.Get(&i, "SELECT * FROM `standup` WHERE id=?", s.ID)

	return i, err
}

// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
func (m *MySQL) SelectStandupByMessageTS(messageTS string) (model.Standup, error) {
	var s model.Standup
	err := m.conn.Get(&s, "SELECT * FROM `standup` WHERE message_ts=?", messageTS)

	return s, err
}

// SelectStandupsByChannelIDForPeriod selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsByChannelIDForPeriod(channelID string, dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE channel_id=? AND created BETWEEN ? AND ?",
		channelID, dateStart, dateEnd)
	return items, err
}

// SelectStandupsFiltered selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsFiltered(slackUserID, channelID string, dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE channel_id=? AND username_id =? AND created BETWEEN ? AND ?",
		channelID, slackUserID, dateStart, dateEnd)
	return items, err
}

// DeleteStandup deletes standup entry from database
func (m *MySQL) DeleteStandup(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `standup` WHERE id=?", id)
	return err
}

// CreateStandupUser creates comedian entry in database
func (m *MySQL) CreateStandupUser(s model.StandupUser) (model.StandupUser, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.conn.Exec(
		"INSERT INTO `standup_users` (created, modified,slack_user_id, username, channel_id, channel, role) VALUES (?, ?, ?, ?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), s.SlackUserID, s.SlackName, s.ChannelID, s.Channel, s.Role)
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

//FindStandupUserInChannelByUserID finds user in channel
func (m *MySQL) FindStandupUserInChannelByUserID(usernameID, channelID string) (model.StandupUser, error) {
	var u model.StandupUser
	err := m.conn.Get(&u, "SELECT * FROM `standup_users` WHERE slack_user_id=? AND channel_id=?", usernameID, channelID)
	return u, err
}

//FindStandupUserByUserName finds user in channel
func (m *MySQL) FindStandupUserByUserName(username string) (model.StandupUser, error) {
	var u model.StandupUser
	err := m.conn.Get(&u, "SELECT * FROM `standup_users` WHERE username=? limit 1", username)
	return u, err
}

// ListAllStandupUsers returns array of standup entries from database
func (m *MySQL) ListAllStandupUsers() ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` where role!='admin'")
	return items, err
}

//GetNonReporters returns a list of non reporters in selected time period
func (m *MySQL) GetNonReporters(channelID string, dateFrom, dateTo time.Time) ([]model.StandupUser, error) {
	nonReporters := []model.StandupUser{}
	err := m.conn.Select(&nonReporters, `SELECT * FROM standup_users where channel_id=? and role!='admin' AND slack_user_id NOT IN (SELECT username_id FROM standup where channel_id=? and created BETWEEN ? AND ?)`, channelID, channelID, dateFrom, dateTo)
	return nonReporters, err
}

// IsNonReporter returns true if user did not submit standup in time period, false othervise
func (m *MySQL) IsNonReporter(slackUserID, channelID string, dateFrom, dateTo time.Time) (bool, error) {
	var standup string
	err := m.conn.Get(&standup, `SELECT comment FROM standup where channel_id=? and username_id=? and created between ? and ?`, channelID, slackUserID, dateFrom, dateTo)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return false, err
	}
	if standup != "" {
		return false, nil
	}
	return true, err
}

// HasExistedAlready returns true if user existed already and therefore could submit standup
func (m *MySQL) HasExistedAlready(slackUserID, channelID string, dateFrom time.Time) (bool, error) {
	var id int
	err := m.conn.Get(&id, `SELECT id FROM standup_users where channel_id=? and slack_user_id=? and created <=?`, channelID, slackUserID, dateFrom)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return false, err
	}
	if id != 0 {
		return true, nil
	}
	return false, nil
}

// CheckIfUserExist returns true if user existed already and therefore could submit standup
func (m *MySQL) CheckIfUserExist(slackUserID string) (bool, error) {
	var id []int
	err := m.conn.Select(&id, `SELECT id FROM standup_users where slack_user_id=?`, slackUserID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return false, err
	}
	if len(id) > 0 {
		return true, nil
	}
	return false, nil
}

// IsAdmin checks if user in channel is of a role admin
func (m *MySQL) IsAdmin(slackUserID, channelID string) bool {
	var u model.StandupUser
	err := m.conn.Get(&u, `SELECT * FROM standup_users where channel_id = ? and slack_user_id = ? and role = ?`, channelID, slackUserID, "admin")
	if err != nil {
		return false
	}
	return true
}

// ListStandupUsersByChannelID returns array of standup entries from database
func (m *MySQL) ListStandupUsersByChannelID(channelID string) ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` WHERE channel_id=? AND role!='admin'", channelID)
	return items, err
}

// ListAdminsByChannelID returns array of standup entries from database
func (m *MySQL) ListAdminsByChannelID(channelID string) ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` WHERE channel_id=? AND role='admin'", channelID)
	return items, err
}

// DeleteStandupUser deletes standup_users entry from database
func (m *MySQL) DeleteStandupUser(username, channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_users` WHERE username=? AND channel_id=? AND role!='admin'", username, channelID)
	return err
}

// DeleteAdmin deletes standup_users entry from database
func (m *MySQL) DeleteAdmin(username, channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_users` WHERE username=? AND channel_id=? AND role='admin'", username, channelID)
	return err
}

// CreateStandupTime creates time entry in database
func (m *MySQL) CreateStandupTime(s model.StandupTime) (model.StandupTime, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	res, err := m.conn.Exec(
		"INSERT INTO `standup_time` (created, channel_id, channel, standuptime) VALUES (?, ?, ?, ?)",
		time.Now().UTC(), s.ChannelID, s.Channel, s.Time)
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

// UpdateStandupTime updates time entry in database
func (m *MySQL) UpdateStandupTime(s model.StandupTime) (model.StandupTime, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
	_, err = m.conn.Exec("UPDATE `standup_time` SET standuptime=? WHERE id=?", s.Time, s.ID)
	if err != nil {
		return s, err
	}
	var i model.StandupTime
	err = m.conn.Get(&i, "SELECT * FROM `standup_time` WHERE id=?", s.ID)
	if err != nil {
		return s, err
	}
	return s, nil
}

// GetChannelStandupTime returns standup time entry from database
func (m *MySQL) GetChannelStandupTime(channelID string) (model.StandupTime, error) {
	var time model.StandupTime
	err := m.conn.Get(&time, "SELECT * FROM `standup_time` WHERE channel_id=?", channelID)
	return time, err
}

// ListAllStandupTime returns standup time entry for all channels from database
func (m *MySQL) ListAllStandupTime() ([]model.StandupTime, error) {
	reminders := []model.StandupTime{}
	err := m.conn.Select(&reminders, "SELECT * FROM `standup_time`")
	return reminders, err
}

// DeleteStandupTime deletes standup_time entry for channel from database
func (m *MySQL) DeleteStandupTime(channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_time` WHERE channel_id=?", channelID)
	return err
}

// AddToStandupHistory creates backup standup entry in standup_edit_history database
func (m *MySQL) AddToStandupHistory(s model.StandupEditHistory) (model.StandupEditHistory, error) {
	err := s.Validate()
	if err != nil {
		return s, err
	}
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

//GetAllChannels returns list of unique channels
func (m *MySQL) GetAllChannels() ([]string, error) {
	channels := []string{}
	err := m.conn.Select(&channels, "SELECT DISTINCT channel_id FROM `standup_users`")
	return channels, err
}

//GetUserChannels returns list of user's channels
func (m *MySQL) GetUserChannels(slackUserID string) ([]string, error) {
	channels := []string{}
	err := m.conn.Select(&channels, "SELECT channel_id FROM `standup_users` where slack_user_id=?", slackUserID)
	return channels, err
}

//GetChannelName returns channel name
func (m *MySQL) GetChannelName(channelID string) (string, error) {
	var channelName string
	err := m.conn.Get(&channelName, "SELECT channel FROM `standup_users` where channel_id =?", channelID)
	if err != nil {
		return "", err
	}

	return channelName, err
}

//GetChannelID returns channel name
func (m *MySQL) GetChannelID(channelName string) (string, error) {
	var channelID string
	err := m.conn.Get(&channelID, "SELECT channel_id FROM `standup_users` where channel=? limit 1", channelName)
	if err != nil {
		return "", err
	}

	return channelID, nil
}

// ListStandups returns array of standup entries from database
// Helper function for testing
func (m *MySQL) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup`")
	return items, err
}
