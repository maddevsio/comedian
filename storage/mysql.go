package storage

import (
	"fmt"
	"strings"
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
	res, err := m.conn.Exec(
		"INSERT INTO `standup` (created, modified, username, comment, channel, channel_id, username_id, message_ts) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		time.Now().UTC(), time.Now().UTC(), s.Username, s.Comment, s.Channel, s.ChannelID, s.UsernameID, s.MessageTS,
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
		"UPDATE `standup` SET modified=?, username=?, username_id=?, comment=?, channel=?, channel_id=?, message_ts=? WHERE id=?",
		time.Now().UTC(), s.Username, s.UsernameID, s.Comment, s.Channel, s.ChannelID, s.MessageTS, s.ID,
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

// SelectStandupsByChannelID selects standup entry by channel ID from database
func (m *MySQL) SelectStandupsByChannelID(channelID string) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE channel_id=?", channelID)
	return items, err
}

// SelectStandupsByChannelIDForPeriod selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsByChannelIDForPeriod(channelID string, dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE channel_id=? AND created BETWEEN ? AND ?",
		channelID, dateStart, dateEnd)
	return items, err
}

// SelectStandupsByChannelIDForPeriod selects standup entrys by channel ID and time period from database
func (m *MySQL) SelectStandupsFiltered(slack_user_id, channelID string, dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE channel_id=? AND username_id =? AND created BETWEEN ? AND ?",
		channelID, slack_user_id, dateStart, dateEnd)
	fmt.Println("STANDUP:", items)
	return items, err
}

// SelectStandupByUserNameForPeriod selects standup entrys by username and time period from database
func (m *MySQL) SelectStandupByUserNameForPeriod(username string, dateStart,
	dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE username=? AND created BETWEEN ? AND ? ",
		username, dateStart, dateEnd)
	return items, err
}

// ListStandups returns array of standup entries from database
func (m *MySQL) ListStandups() ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup`")
	return items, err
}

// SelectStandupsForPeriod selects standup entrys for time period from database
func (m *MySQL) SelectStandupsForPeriod(dateStart, dateEnd time.Time) ([]model.Standup, error) {
	items := []model.Standup{}
	err := m.conn.Select(&items, "SELECT * FROM `standup` WHERE created BETWEEN ? AND ?",
		dateStart, dateEnd)
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

//FindStandupUserInChannel finds user in channel
func (m *MySQL) FindStandupUserInChannel(username, channelID string) (model.StandupUser, error) {
	var u model.StandupUser
	err := m.conn.Get(&u, "SELECT * FROM `standup_users` WHERE username=? AND channel_id=?", username, channelID)
	return u, err
}

//FindStandupUserInChannelByUserID finds user in channel
func (m *MySQL) FindStandupUserInChannelByUserID(usernameID, channelID string) (model.StandupUser, error) {
	var u model.StandupUser
	err := m.conn.Get(&u, "SELECT * FROM `standup_users` WHERE slack_user_id=? AND channel_id=?", usernameID, channelID)
	return u, err
}

//FindStandupUser finds user in
func (m *MySQL) FindStandupUser(username string) (model.StandupUser, error) {
	var u model.StandupUser
	err := m.conn.Get(&u, "SELECT * FROM `standup_users` WHERE username=?", username)
	return u, err
}

//FindStandupUsers finds users in table
func (m *MySQL) FindStandupUsers(username string) ([]model.StandupUser, error) {
	users := []model.StandupUser{}
	err := m.conn.Select(&users, "SELECT * FROM `standup_users` WHERE username=?", username)
	return users, err
}

// ListAllStandupUsers returns array of standup entries from database
func (m *MySQL) ListAllStandupUsers() ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` where role!='admin'")
	return items, err
}

func (m *MySQL) GetNonReporters(channelID string, dateFrom, dateTo time.Time) ([]model.StandupUser, error) {
	nonReporterUserIDs := []string{}
	err := m.conn.Select(&nonReporterUserIDs, `SELECT username_id FROM standup where channel_id=? and not created BETWEEN ? AND ?`, channelID, dateFrom, dateTo)
	users := []model.StandupUser{}
	err = m.conn.Select(&users, `SELECT * FROM standup_users where channel_id = ? and role!='admin' and slack_user_id IN (?)`, channelID, strings.Join(nonReporterUserIDs, ","))
	return users, err
}

func (m *MySQL) CheckNonReporter(user model.StandupUser, dateFrom, dateTo time.Time) (bool, error) {
	fmt.Println(dateFrom)
	fmt.Println(dateTo)
	fmt.Println(user.ChannelID, user.SlackUserID)
	var id int
	err := m.conn.Get(&id, `SELECT id FROM standup where channel_id=? and username_id=? and created between ? and ?`, user.ChannelID, user.SlackUserID, dateFrom, dateTo)
	if err != nil {
		fmt.Println(err)
		return true, err
	}
	fmt.Println(id)
	if id != 0 {
		return false, nil
	}
	fmt.Println("Anyway the user is a rook!")
	return true, nil
}

// isAdmin checks if user in channel is of a role admin
func (m *MySQL) IsAdmin(slack_user_id, channel_id string) bool {
	var u model.StandupUser
	err := m.conn.Get(&u, `SELECT * FROM standup_users where channel_id = ? and slack_user_id = ? and role = ?`, channel_id, slack_user_id, "admin")
	if err != nil {
		return false
	}
	return true
}

func (m *MySQL) GetNonReporter(userID, channelID string, dateFrom, dateTo time.Time) ([]model.StandupUser, error) {
	us := []model.StandupUser{}
	err := m.conn.Select(&us, `SELECT r.* FROM standup_users r left join standup s
								on s.username_id = r.slack_user_id and r.channel_id = ? 
								where s.created BETWEEN ? AND ?`, channelID, dateFrom, dateTo)
	usernames := []string{}
	for _, user := range us {
		usernames = append(usernames, user.SlackName)
	}
	user := []model.StandupUser{}
	err = m.conn.Select(&user, `SELECT * FROM standup_users where channel_id = ? and username NOT IN (?) and slack_user_id = ? and created between ? and ?`, channelID, strings.Join(usernames, ","), userID, dateFrom, dateTo)

	return user, err
}

// ListStandupUsersByChannelID returns array of standup entries from database
func (m *MySQL) ListStandupUsersByChannelID(channelID string) ([]model.StandupUser, error) {
	items := []model.StandupUser{}
	err := m.conn.Select(&items, "SELECT * FROM `standup_users` WHERE channel_id=? AND role!='admin'", channelID)
	return items, err
}

// DeleteStandupUserByUsername deletes standup_users entry from database
func (m *MySQL) DeleteStandupUserByUsername(username, channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `standup_users` WHERE username=? AND channel_id=?", username, channelID)
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

// ListStandupTime returns standup time entry from database
func (m *MySQL) ListStandupTime(channelID string) (model.StandupTime, error) {
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
