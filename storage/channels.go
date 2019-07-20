package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateChannel creates standup entry in database
func (m *DB) CreateChannel(ch model.Channel) (model.Channel, error) {
	err := ch.Validate()
	if err != nil {
		return ch, err
	}
	res, err := m.db.Exec(
		"INSERT INTO `channels` (team_id, channel_name, channel_id, channel_standup_time) VALUES (?, ?, ?, ?)",
		ch.TeamID, ch.ChannelName, ch.ChannelID, ch.StandupTime,
	)
	if err != nil {
		return ch, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ch, err
	}
	ch.ID = id

	return ch, nil
}

// UpdateChannel updates Channel entry in database
func (m *DB) UpdateChannel(ch model.Channel) (model.Channel, error) {
	err := ch.Validate()
	if err != nil {
		return ch, err
	}
	_, err = m.db.Exec(
		"UPDATE `channels` SET channel_standup_time=?  WHERE id=?",
		ch.StandupTime, ch.ID,
	)
	if err != nil {
		return ch, err
	}
	return ch, nil
}

//ListChannels returns list of channels
func (m *DB) ListChannels() ([]model.Channel, error) {
	channels := []model.Channel{}
	err := m.db.Select(&channels, "SELECT * FROM `channels`")
	return channels, err
}

//ListTeamChannels returns list of channels
func (m *DB) ListTeamChannels(teamID string) ([]model.Channel, error) {
	channels := []model.Channel{}
	err := m.db.Select(&channels, "SELECT * FROM `channels` where team_id=?", teamID)
	return channels, err
}

// SelectChannel selects Channel entry from database
func (m *DB) SelectChannel(channelID string) (model.Channel, error) {
	var c model.Channel
	err := m.db.Get(&c, "SELECT * FROM `channels` WHERE channel_id=?", channelID)
	if err != nil {
		return c, err
	}
	return c, err
}

// GetChannel selects Channel entry from database with specific id
func (m *DB) GetChannel(id int64) (model.Channel, error) {
	var c model.Channel
	err := m.db.Get(&c, "SELECT * FROM `channels` where id=?", id)
	if err != nil {
		return c, err
	}
	return c, err
}

// DeleteChannel deletes Channel entry from database
func (m *DB) DeleteChannel(id int64) error {
	_, err := m.db.Exec("DELETE FROM `channels` WHERE id=?", id)
	return err
}
