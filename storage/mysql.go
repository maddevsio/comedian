package storage

import (
	"fmt"
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
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
	conn.SetConnMaxLifetime(time.Second)
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
// Helper function for testing
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

// CreateChannelMember creates comedian entry in database
func (m *MySQL) CreateChannelMember(s model.ChannelMember) (model.ChannelMember, error) {
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

//FindChannelMemberByUserID finds user in channel
func (m *MySQL) FindChannelMemberByUserID(userID, channelID string) (model.ChannelMember, error) {
	var u model.ChannelMember
	err := m.conn.Get(&u, "SELECT * FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	return u, err
}

//FindMembersByUserID finds user in channel
func (m *MySQL) FindMembersByUserID(userID string) ([]model.ChannelMember, error) {
	var u []model.ChannelMember
	err := m.conn.Select(&u, "SELECT * FROM `channel_members` WHERE user_id=?", userID)
	return u, err
}

//SelectChannelMember finds user in channel
func (m *MySQL) SelectChannelMember(id int64) (model.ChannelMember, error) {
	var u model.ChannelMember
	err := m.conn.Get(&u, "SELECT * FROM `channel_members` WHERE id=?", id)
	return u, err
}

//FindChannelMemberByUserName finds user in channel
func (m *MySQL) FindChannelMemberByUserName(userName, channelID string) (model.ChannelMember, error) {
	var u model.ChannelMember
	err := m.conn.Get(&u, "SELECT * FROM `channel_members` WHERE user_id=(select user_id from users where user_name=?) and channel_id=?", userName, channelID)
	return u, err
}

// ListAllChannelMembers returns array of standup entries from database
func (m *MySQL) ListAllChannelMembers() ([]model.ChannelMember, error) {
	items := []model.ChannelMember{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members`")
	return items, err
}

//GetNonReporters returns a list of non reporters in selected time period
func (m *MySQL) GetNonReporters(channelID string, dateFrom, dateTo time.Time) ([]model.ChannelMember, error) {
	nonReporters := []model.ChannelMember{}
	err := m.conn.Select(&nonReporters, `SELECT * FROM channel_members where channel_id=? AND role_in_channel != 'pm' AND user_id NOT IN (SELECT user_id FROM standups where channel_id=? and created BETWEEN ? AND ?)`, channelID, channelID, dateFrom, dateTo)
	return nonReporters, err
}

//SubmittedStandupToday shows if a user submitted standup today
func (m *MySQL) SubmittedStandupToday(userID, channelID string) bool {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	var standup string
	err := m.conn.Get(&standup, `SELECT comment FROM standups where channel_id=? and user_id=? and created between ? and ?`, channelID, userID, timeFrom, time.Now())
	if err != nil {
		logrus.Infof("User '%v' did not write standup in channel '%v' today yet \n", userID, channelID)
		return false
	}
	return true
}

// IsNonReporter returns true if user did not submit standup in time period, false othervise
func (m *MySQL) IsNonReporter(userID, channelID string, dateFrom, dateTo time.Time) (bool, error) {
	var standup string
	query := fmt.Sprintf("SELECT comment FROM standups where channel_id='%v' and user_id='%v' and created between '%v' and '%v'", channelID, userID, dateFrom, dateTo)
	logrus.Infof("IsNonreporter Query: %s", query)
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
func (m *MySQL) ListChannelMembers(channelID string) ([]model.ChannelMember, error) {
	items := []model.ChannelMember{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members` WHERE channel_id=?", channelID)
	return items, err
}

//ListChannelMembersByRole lists channels members with the same role
func (m *MySQL) ListChannelMembersByRole(channelID, role string) ([]model.ChannelMember, error) {
	items := []model.ChannelMember{}
	err := m.conn.Select(&items, "SELECT * FROM `channel_members` WHERE channel_id=? and role_in_channel=?", channelID, role)
	return items, err
}

// DeleteChannelMember deletes channel_members entry from database
func (m *MySQL) DeleteChannelMember(userID, channelID string) error {
	_, err := m.conn.Exec("DELETE FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	return err
}

// CreateStandupTime creates time entry in database
func (m *MySQL) CreateStandupTime(st int64, channelID string) error {
	_, err := m.conn.Exec("UPDATE `channels` SET channel_standup_time=? WHERE channel_id=?", st, channelID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateChannelStandupTime updates time entry in database
func (m *MySQL) UpdateChannelStandupTime(st int64, channelID string) error {
	_, err := m.conn.Exec("UPDATE `channels` SET channel_standup_time=? WHERE channel_id=?", st, channelID)
	if err != nil {
		return err
	}
	return nil
}

// GetChannelStandupTime returns standup time entry from database
func (m *MySQL) GetChannelStandupTime(channelID string) (int64, error) {
	var time int64
	err := m.conn.Get(&time, "SELECT channel_standup_time FROM `channels` WHERE channel_id=?", channelID)
	return time, err
}

// ListAllStandupTime returns standup time entry for all channels from database
func (m *MySQL) ListAllStandupTime() ([]int64, error) {
	deadlines := []int64{}
	err := m.conn.Select(&deadlines, "SELECT channel_standup_time FROM `channels` where channel_standup_time>0")
	return deadlines, err
}

// DeleteStandupTime deletes channels entry for channel from database
func (m *MySQL) DeleteStandupTime(channelID string) error {
	_, err := m.conn.Exec("UPDATE `channels` SET channel_standup_time=0 WHERE channel_id=?", channelID)
	return err
}

// CreateChannel creates standup entry in database
func (m *MySQL) CreateChannel(c model.Channel) (model.Channel, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `channels` (team_id, channel_name, channel_id, channel_standup_time) VALUES (?, ?, ?, ?)",
		c.TeamID, c.ChannelName, c.ChannelID, 0,
	)
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

//GetAllChannels returns list of unique channels
func (m *MySQL) GetAllChannels() ([]model.Channel, error) {
	channels := []model.Channel{}
	err := m.conn.Select(&channels, "SELECT * FROM `channels`")
	return channels, err
}

//GetUserChannels returns list of user's channels
func (m *MySQL) GetUserChannels(userID string) ([]string, error) {
	channels := []string{}
	err := m.conn.Select(&channels, "SELECT channel_id FROM `channel_members` where user_id=?", userID)
	return channels, err
}

//GetChannelName returns channel name
func (m *MySQL) GetChannelName(channelID string) (string, error) {
	var channelName string
	err := m.conn.Get(&channelName, "SELECT channel_name FROM `channels` where channel_id=?", channelID)
	if err != nil {
		return "", err
	}
	return channelName, err
}

//GetChannelID returns channel name
func (m *MySQL) GetChannelID(channelName string) (string, error) {
	var channelID string
	err := m.conn.Get(&channelID, "SELECT channel_id FROM `channels` where channel_name=?", channelName)
	if err != nil {
		return "", err
	}
	return channelID, nil
}

// SelectChannel selects Channel entry from database
func (m *MySQL) SelectChannel(channelID string) (model.Channel, error) {
	var c model.Channel
	err := m.conn.Get(&c, "SELECT * FROM `channels` WHERE channel_id=?", channelID)
	if err != nil {
		return c, err
	}
	return c, err
}

// GetChannels selects Channel entry from database
func (m *MySQL) GetChannels() ([]model.Channel, error) {
	var c []model.Channel
	err := m.conn.Select(&c, "SELECT * FROM `channels`")
	if err != nil {
		return c, err
	}
	return c, err
}

// GetTeamChannels selects Channel entry from database with specific team_id
func (m *MySQL) GetTeamChannels(teamID string) ([]model.Channel, error) {
	var c []model.Channel
	err := m.conn.Select(&c, "SELECT * FROM `channels` where team_id=?", teamID)
	if err != nil {
		return c, err
	}
	return c, err
}

// DeleteChannel deletes Channel entry from database
func (m *MySQL) DeleteChannel(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `channels` WHERE id=?", id)
	return err
}

// CreateUser creates standup entry in database
func (m *MySQL) CreateUser(c model.User) (model.User, error) {
	res, err := m.conn.Exec(
		"INSERT INTO `users` (team_id, user_name, user_id, role, real_name) VALUES (?, ?, ?, ?, ?)",
		c.TeamID, c.UserName, c.UserID, c.Role, c.RealName,
	)
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

// UpdateUser updates User entry in database
func (m *MySQL) UpdateUser(c model.User) (model.User, error) {
	_, err := m.conn.Exec(
		"UPDATE `users` SET role=?, real_name=?, team_id=? WHERE id=?",
		c.Role, c.RealName, c.TeamID, c.ID,
	)
	if err != nil {
		return c, err
	}
	var i model.User
	err = m.conn.Get(&i, "SELECT * FROM `users` WHERE id=?", c.ID)
	return i, err
}

// SelectUser selects User entry from database
func (m *MySQL) SelectUser(userID string) (model.User, error) {
	var c model.User
	err := m.conn.Get(&c, "SELECT * FROM `users` WHERE user_id=?", userID)
	if err != nil {
		return c, err
	}
	return c, err
}

// ListUsers selects Users from database
func (m *MySQL) ListUsers() ([]model.User, error) {
	var u []model.User
	err := m.conn.Select(&u, "SELECT * FROM `users`")
	if err != nil {
		return u, err
	}
	return u, err
}

// SelectUserByUserName selects User entry from database
func (m *MySQL) SelectUserByUserName(userName string) (model.User, error) {
	var c model.User
	err := m.conn.Get(&c, "SELECT * FROM `users` WHERE user_name=?", userName)
	if err != nil {
		return c, err
	}
	return c, err
}

// DeleteUser deletes User entry from database
func (m *MySQL) DeleteUser(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `users` WHERE id=?", id)
	return err
}

// ListAdmins selects User entry from database
func (m *MySQL) ListAdmins() ([]model.User, error) {
	var c []model.User
	err := m.conn.Select(&c, "SELECT * FROM `users` WHERE role='admin'")
	if err != nil {
		return c, err
	}
	return c, err
}

// UserIsPMForProject returns true if user is a project's PM.
func (m *MySQL) UserIsPMForProject(userID, channelID string) bool {
	var role string
	err := m.conn.Get(&role, "SELECT role_in_channel FROM `channel_members` WHERE user_id=? AND channel_id=?", userID, channelID)
	if err != nil {
		return false
	}
	logrus.Infof("Role in channel %v", role)
	if role == "pm" {
		return true
	}
	return false
}

//CreateControlPannel creates bot properties for the newly created bot
func (m *MySQL) CreateControlPannel(token, teamID, teamName string) (model.ControlPannel, error) {
	var cp model.ControlPannel
	_, err := m.conn.Exec(
		"INSERT INTO `controll_pannel` (notifier_interval, manager_slack_user_id, reporting_channel, report_time, language, reminder_repeats_max, reminder_time, collector_enabled,sprint_report_status,sprint_report_time,sprint_report_channel,sprint_weekdays,individual_reporting_status,bot_access_token, team_id, team_name, password, task_done_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		30, "", "", "10:00", "en_US", 3, int64(10), false, false, "9:00", "", "", false, token, teamID, teamName, teamName, "")
	if err != nil {
		return model.ControlPannel{}, err
	}
	cp, err = m.GetControlPannel(teamName)
	if err != nil {
		return cp, err
	}

	return cp, nil
}

//GetControlPannels returns all controll pannels
func (m *MySQL) GetControlPannels() ([]model.ControlPannel, error) {
	var cp []model.ControlPannel
	err := m.conn.Select(&cp, "SELECT * FROM `controll_pannel`")
	if err != nil {
		return cp, err
	}
	return cp, nil
}

//GetControlPannel returns a particular controll pannel
func (m *MySQL) GetControlPannel(teamName string) (model.ControlPannel, error) {
	var cp model.ControlPannel
	err := m.conn.Get(&cp, "SELECT * FROM `controll_pannel` where team_name=?", teamName)
	if err != nil {
		return cp, err
	}
	return cp, nil
}

//UpdateControlPannel updates controll pannel
func (m *MySQL) UpdateControlPannel(cp model.ControlPannel) (model.ControlPannel, error) {
	_, err := m.conn.Exec(
		"UPDATE `controll_pannel` set notifier_interval=?, manager_slack_user_id=?, reporting_channel=?, report_time=?, language=?, reminder_repeats_max=?, reminder_time=?, collector_enabled=?, sprint_report_status=?, sprint_report_time=?, sprint_report_channel=?,sprint_weekdays=?,individual_reporting_status=?, password=?, task_done_status=? where id=?",
		cp.NotifierInterval, cp.ManagerSlackUserID, cp.ReportingChannel, cp.ReportTime, cp.Language, cp.ReminderRepeatsMax, cp.ReminderTime, cp.CollectorEnabled, cp.SprintReportStatus, cp.SprintReportTime, cp.SprintReportChannel, cp.SprintWeekdays, cp.IndividualReportingStatus, cp.Password, cp.TaskDoneStatus, cp.ID,
	)
	if err != nil {
		return cp, err
	}
	var ControlPannel model.ControlPannel
	err = m.conn.Get(&cp, "SELECT * FROM `controll_pannel`")
	if err != nil {
		return cp, err
	}
	return ControlPannel, err
}

//DeleteControlPannel deletes controll pannel
func (m *MySQL) DeleteControlPannel(teamID string) error {
	_, err := m.conn.Exec("DELETE FROM `controll_pannel` WHERE team_id=?", teamID)
	return err
}
