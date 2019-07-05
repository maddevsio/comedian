package storage

import (

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"github.com/maddevsio/comedian/crypto"
	"github.com/maddevsio/comedian/model"
)

//CreateBotSettings creates bot properties for the newly created bot
func (m *DB) CreateBotSettings(token, password, userID, teamID, teamName string) (model.BotSettings, error) {
	bs := model.BotSettings{
		NotifierInterval:    30,
		Language:            "en_US",
		ReminderRepeatsMax:  3,
		ReminderTime:        int64(10),
		AccessToken:         token,
		UserID:              userID,
		TeamID:              teamID,
		TeamName:            teamName,
		Password:            password,
		ReportingChannel:    "",
		ReportingTime:       "9:00",
		IndividualReportsOn: false,
	}

	err := bs.Validate()
	if err != nil {
		return bs, err
	}

	_, err = m.db.Exec(
		"INSERT INTO `bot_settings` (notifier_interval, language, reminder_repeats_max, reminder_time, bot_access_token, user_id, team_id, team_name, password, reporting_channel, reporting_time, individual_reports_on) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bs.NotifierInterval, bs.Language, bs.ReminderRepeatsMax, bs.ReminderTime, bs.AccessToken, bs.UserID, bs.TeamID, bs.TeamName, bs.Password, bs.ReportingChannel, bs.ReportingTime, bs.IndividualReportsOn)
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
	err := m.db.Select(&bs, "SELECT * FROM `bot_settings`")
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettingsByTeamName returns a particular bot
func (m *DB) GetBotSettingsByTeamName(teamName string) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.db.Get(&bs, "SELECT * FROM `bot_settings` where team_name=?", teamName)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettingsByTeamID returns a particular bot
func (m *DB) GetBotSettingsByTeamID(teamID string) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.db.Get(&bs, "SELECT * FROM `bot_settings` where team_id=?", teamID)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettings returns a particular bot
func (m *DB) GetBotSettings(id int64) (model.BotSettings, error) {
	var bs model.BotSettings
	err := m.db.Get(&bs, "SELECT * FROM `bot_settings` where id=?", id)
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

	_, err = m.db.Exec(
		"UPDATE `bot_settings` set bot_access_token=?, password=?, user_id=?, notifier_interval=?, language=?, reminder_repeats_max=?, reminder_time=?, reporting_channel=?, reporting_time=?, individual_reports_on=? where id=?",
		settings.AccessToken, settings.Password, settings.UserID, settings.NotifierInterval, settings.Language, settings.ReminderRepeatsMax, settings.ReminderTime, settings.ReportingChannel, settings.ReportingTime, settings.IndividualReportsOn, settings.ID,
	)
	if err != nil {
		return settings, err
	}
	var bs model.BotSettings
	err = m.db.Get(&bs, "SELECT * FROM `bot_settings` where id=?", settings.ID)
	if err != nil {
		return settings, err
	}
	return bs, err
}

//UpdateBotPassword updates bot pass
func (m *DB) UpdateBotPassword(settings model.BotSettings) (model.BotSettings, error) {
	err := settings.Validate()
	if err != nil {
		return settings, err
	}

	password, err := crypto.Generate(settings.Password)
	if err != nil {
		return settings, err
	}
	_, err = m.db.Exec("UPDATE `bot_settings` set password=? where id=?", password, settings.ID)
	if err != nil {
		return settings, err
	}
	var bs model.BotSettings
	err = m.db.Get(&bs, "SELECT * FROM `bot_settings` where id=?", settings.ID)
	if err != nil {
		return settings, err
	}
	return bs, err
}

//DeleteBotSettingsByID deletes bot
func (m *DB) DeleteBotSettingsByID(id int64) error {
	_, err := m.db.Exec("DELETE FROM `bot_settings` WHERE id=?", id)
	return err
}

//DeleteBotSettings deletes bot
func (m *DB) DeleteBotSettings(teamID string) error {
	_, err := m.db.Exec("DELETE FROM `bot_settings` WHERE team_id=?", teamID)
	return err
}
