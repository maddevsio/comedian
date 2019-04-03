package storage

import (

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/crypto"
	"gitlab.com/team-monitoring/comedian/model"
)

//CreateBotSettings creates bot properties for the newly created bot
func (m *DB) CreateBotSettings(token, password, userID, teamID, teamName string) (model.BotSettings, error) {
	bs := model.BotSettings{
		NotifierInterval:   30,
		Language:           "en_US",
		ReminderRepeatsMax: 3,
		ReminderTime:       int64(10),
		AccessToken:        token,
		UserID:             userID,
		TeamID:             teamID,
		TeamName:           teamName,
		Password:           password,
	}

	err := bs.Validate()
	if err != nil {
		return bs, err
	}

	_, err = m.DB.Exec(
		"INSERT INTO `bot_settings` (notifier_interval, language, reminder_repeats_max, reminder_time, bot_access_token, user_id, team_id, team_name, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bs.NotifierInterval, bs.Language, bs.ReminderRepeatsMax, bs.ReminderTime, bs.AccessToken, bs.UserID, bs.TeamID, bs.TeamName, bs.Password)
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
func (m *DB) GetAllBotSettings() ([]model.BotSettings, error) {
	var bs []model.BotSettings
	err := m.DB.Select(&bs, "SELECT * FROM `bot_settings`")
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettingsByTeamName returns a particular bot
// When dashboard is ready - DELETE!
func (m *DB) GetBotSettingsByTeamName(teamName string) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.DB.Get(&bs, "SELECT * FROM `bot_settings` where team_name=?", teamName)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettings returns a particular bot
func (m *DB) GetBotSettings(id int64) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.DB.Get(&bs, "SELECT * FROM `bot_settings` where id=?", id)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//UpdateBotSettings updates bot
func (m *DB) UpdateBotSettings(settings model.BotSettings) (model.BotSettings, error) {
	err := settings.Validate()
	if err != nil {
		return settings, err
	}

	password, err := crypto.Generate(settings.Password)
	if err != nil {
		return settings, err
	}
	_, err = m.DB.Exec(
		"UPDATE `bot_settings` set notifier_interval=?, language=?, reminder_repeats_max=?, reminder_time=?, password=? where id=?",
		settings.NotifierInterval, settings.Language, settings.ReminderRepeatsMax, settings.ReminderTime, password, settings.ID,
	)
	if err != nil {
		return settings, err
	}
	var bs model.BotSettings
	err = m.DB.Get(&bs, "SELECT * FROM `bot_settings` where id=?", settings.ID)
	if err != nil {
		return settings, err
	}
	return bs, err
}

//DeleteBotSettingsByID deletes bot
func (m *DB) DeleteBotSettingsByID(id int64) error {
	_, err := m.DB.Exec("DELETE FROM `bot_settings` WHERE id=?", id)
	return err
}

//DeleteBotSettings deletes bot
func (m *DB) DeleteBotSettings(teamID string) error {
	_, err := m.DB.Exec("DELETE FROM `bot_settings` WHERE team_id=?", teamID)
	return err
}
