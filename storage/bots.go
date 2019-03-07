package storage

import (

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/model"
)

//CreateBotSettings creates bot properties for the newly created bot
func (m *MySQL) CreateBotSettings(token, teamID, teamName string) (model.BotSettings, error) {
	bs := model.BotSettings{
		NotifierInterval:   30,
		Language:           "en_US",
		ReminderRepeatsMax: 3,
		ReminderTime:       int64(10),
		AccessToken:        token,
		TeamID:             teamID,
		TeamName:           teamName,
		Password:           teamName,
	}

	err := bs.Validate()
	if err != nil {
		return bs, err
	}

	_, err = m.conn.Exec(
		"INSERT INTO `bot_settings` (notifier_interval, language, reminder_repeats_max, reminder_time, bot_access_token, team_id, team_name, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		bs.NotifierInterval, bs.Language, bs.ReminderRepeatsMax, bs.ReminderTime, bs.AccessToken, bs.TeamID, bs.TeamName, bs.Password)
	if err != nil {
		return bs, err
	}

	bs, err = m.GetBotSettingsByTeamName(teamName)
	if err != nil {
		return bs, err
	}

	return bs, nil
}

//GetAllBotSettings returns all bots
func (m *MySQL) GetAllBotSettings() ([]model.BotSettings, error) {
	var bs []model.BotSettings
	err := m.conn.Select(&bs, "SELECT * FROM `bot_settings`")
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettingsByTeamName returns a particular bot
// When dashboard is ready - DELETE!
func (m *MySQL) GetBotSettingsByTeamName(teamName string) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.conn.Get(&bs, "SELECT * FROM `bot_settings` where team_name=?", teamName)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettings returns a particular bot
func (m *MySQL) GetBotSettings(id int64) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.conn.Get(&bs, "SELECT * FROM `bot_settings` where id=?", id)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//UpdateBotSettings updates bot
func (m *MySQL) UpdateBotSettings(settings model.BotSettings) (model.BotSettings, error) {
	_, err := m.conn.Exec(
		"UPDATE `bot_settings` set notifier_interval=?, language=?, reminder_repeats_max=?, reminder_time=?, password=? where id=?",
		settings.NotifierInterval, settings.Language, settings.ReminderRepeatsMax, settings.ReminderTime, settings.Password, settings.ID,
	)
	if err != nil {
		return settings, err
	}
	var bs model.BotSettings
	err = m.conn.Get(&bs, "SELECT * FROM `bot_settings` where id=?", settings.ID)
	if err != nil {
		return settings, err
	}
	return bs, err
}

//DeleteBotSettingsByID deletes bot
func (m *MySQL) DeleteBotSettingsByID(id int64) error {
	_, err := m.conn.Exec("DELETE FROM `bot_settings` WHERE id=?", id)
	return err
}

//DeleteBotSettings deletes bot
func (m *MySQL) DeleteBotSettings(teamID string) error {
	_, err := m.conn.Exec("DELETE FROM `bot_settings` WHERE team_id=?", teamID)
	return err
}
