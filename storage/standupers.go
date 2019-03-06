package storage

import (
	"fmt"
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
		"INSERT INTO `channel_members` (team_id, user_id, channel_id, standup_time, role_in_channel, created) VALUES (?,?, ?,?, ?, ?)",
		s.TeamID, s.UserID, s.ChannelID, 0, s.RoleInChannel, time.Now())
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
		"UPDATE `channel_members` SET standup_time=?, role_in_channel=? WHERE id=?",
		st.StandupTime, st.RoleInChannel, st.ID,
	)
	if err != nil {
		return st, err
	}
	var i model.Standuper
	err = m.conn.Get(&i, "SELECT * FROM `channel_members` WHERE id=?", st.ID)
	return i, err
}

//FindChannelMemberByUserID finds user in channel
func (m *MySQL) FindChannelMemberByUserID(userID, channelID string) (model.Standuper, error) {
	var u model.Standuper
	err := m.conn.Get(&u, "SELECT * FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	return u, err
}

//FindMembersByUserID finds user in channel
func (m *MySQL) FindMembersByUserID(userID string) ([]model.Standuper, error) {
	var u []model.Standuper
	err := m.conn.Select(&u, "SELECT * FROM `channel_members` WHERE user_id=?", userID)
	return u, err
}

// ListStandupers returns array of standup entries from database
func (m *MySQL) ListStandupers() ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members`")
	return items, err
}

//GetNonReporters returns a list of non reporters in selected time period
func (m *MySQL) GetNonReporters(channelID string, dateFrom, dateTo time.Time) ([]model.Standuper, error) {
	nonReporters := []model.Standuper{}
	err := m.conn.Select(&nonReporters, `SELECT * FROM channel_members where channel_id=? AND role_in_channel != 'pm' AND user_id NOT IN (SELECT user_id FROM standups where channel_id=? and created BETWEEN ? AND ?)`, channelID, channelID, dateFrom, dateTo)
	return nonReporters, err
}

//GetStanduper returns a standuper
func (m *MySQL) GetStanduper(id int64) (model.Standuper, error) {
	standuper := model.Standuper{}
	err := m.conn.Get(&standuper, `SELECT * FROM channel_members where id=?)`, id)
	return standuper, err
}

//SubmittedStandupToday shows if a user submitted standup today
func (m *MySQL) SubmittedStandupToday(userID, channelID string) bool {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	var standup string
	err := m.conn.Get(&standup, `SELECT comment FROM standups where channel_id=? and user_id=? and created between ? and ?`, channelID, userID, timeFrom, time.Now())
	if err != nil {
		return false
	}
	return true
}

// IsNonReporter returns true if user did not submit standup in time period, false othervise
func (m *MySQL) IsNonReporter(userID, channelID string, dateFrom, dateTo time.Time) (bool, error) {
	var standup string
	query := fmt.Sprintf("SELECT comment FROM standups where channel_id='%v' and user_id='%v' and created between '%v' and '%v'", channelID, userID, dateFrom, dateTo)

	err := m.conn.Get(&standup, query)
	if err != nil {
		return false, err
	}
	if standup == "" {
		return true, nil
	}
	return false, nil
}

// ListChannelMembers returns array of standup entries from database
func (m *MySQL) ListChannelMembers(channelID string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members` WHERE channel_id=?", channelID)
	return items, err
}

//ListChannelMembersByRole lists channels members with the same role
func (m *MySQL) ListChannelMembersByRole(channelID, role string) ([]model.Standuper, error) {
	items := []model.Standuper{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members` WHERE channel_id=? and role_in_channel=?", channelID, role)
	return items, err
}

// DeleteStanduper deletes channel_members entry from database
func (m *MySQL) DeleteStanduper(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `channel_members` id=?", id)
	return err
}

// UserIsPMForProject returns true if user is a project's PM.
func (m *MySQL) UserIsPMForProject(userID, channelID string) bool {
	var role string
	err := m.conn.Get(&role, "SELECT role_in_channel FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	if err != nil {
		return false
	}
	if role == "pm" {
		return true
	}
	return false
}
