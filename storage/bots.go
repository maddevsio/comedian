package storage

import (
	"github.com/maddevsio/comedian/model"
)

//CreateBotSettings creates bot properties for the newly created bot
func (m *DB) CreateBotSettings(bs *model.BotSettings) (*model.BotSettings, error) {
	err := bs.Validate()
	if err != nil {
		return bs, err
	}

	res, err := m.db.Exec(
		"INSERT INTO `bot_settings` (notifier_interval, language, reminder_repeats_max, reminder_time, bot_access_token, user_id, team_id, team_name, reporting_channel, reporting_time, individual_reports_on) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bs.NotifierInterval, bs.Language, bs.ReminderRepeatsMax, bs.ReminderTime, bs.AccessToken, bs.UserID, bs.TeamID, bs.TeamName, bs.ReportingChannel, bs.ReportingTime, bs.IndividualReportsOn)
	if err != nil {
		return bs, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return bs, err
	}

	bs.ID = id

	return bs, nil
}

//GetAllBotSettings returns all bots
func (m *DB) GetAllBotSettings() ([]model.BotSettings, error) {
	bs := []model.BotSettings{}
	err := m.db.Select(&bs, "SELECT * FROM `bot_settings`")
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettingsByTeamID returns a particular bot
func (m *DB) GetBotSettingsByTeamID(teamID string) (*model.BotSettings, error) {
	bs := &model.BotSettings{}
	err := m.db.Get(bs, "SELECT * FROM `bot_settings` where team_id=?", teamID)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//GetBotSettings returns a particular bot
func (m *DB) GetBotSettings(id int64) (*model.BotSettings, error) {
	bs := &model.BotSettings{}
	err := m.db.Get(bs, "SELECT * FROM `bot_settings` where id=?", id)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

//UpdateBotSettings updates bot
func (m *DB) UpdateBotSettings(settings *model.BotSettings) (*model.BotSettings, error) {
	err := settings.Validate()
	if err != nil {
		return settings, err
	}

	_, err = m.db.Exec(
		"UPDATE `bot_settings` set bot_access_token=?, user_id=?, notifier_interval=?, language=?, reminder_repeats_max=?, reminder_time=?, reporting_channel=?, reporting_time=?, individual_reports_on=? where id=?",
		settings.AccessToken, settings.UserID, settings.NotifierInterval, settings.Language, settings.ReminderRepeatsMax, settings.ReminderTime, settings.ReportingChannel, settings.ReportingTime, settings.IndividualReportsOn, settings.ID,
	)
	if err != nil {
		return settings, err
	}
	bs := &model.BotSettings{}
	err = m.db.Get(bs, "SELECT * FROM `bot_settings` where id=?", settings.ID)
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
