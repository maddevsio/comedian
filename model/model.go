package model

import "time"

type (
	// Standup model used for serialization/deserialization stored standups
	Standup struct {
		ID         int64     `db:"id" json:"id"`
		Created    time.Time `db:"created" json:"created"`
		Channel    string    `db:"channel" json:"channel"`
		ChannelID  string    `db:"channel_id" json:"channelId"`
		Modified   time.Time `db:"modified" json:"modified"`
		UsernameID string    `db:"username_id" json:"userNameId"`
		Username   string    `db:"username" json:"userName"`
		FullName   string    `db:"full_name" json:"fullName"`
		Comment    string    `db:"comment" json:"comment"`
		MessageTS  string    `db:"message_ts" json:"message_ts"`
	}

	// StandupUser model used for serialization/deserialization stored standupUsers
	StandupUser struct {
		ID        int64     `db:"id" json:"id"`
		Created   time.Time `db:"created" json:"created"`
		Modified  time.Time `db:"modified" json:"modified"`
		SlackName string    `db:"username" json:"username"`
		FullName  string    `db:"full_name" json:"fullName"`
		Channel   string    `db:"channel" json:"channel"`
		ChannelID string    `db:"channel_id" json:"channelId"`
	}

	// StandupTime model used for serialization/deserialization stored standupTime
	StandupTime struct {
		ID        int64     `db:"id" json:"id"`
		Created   time.Time `db:"created" json:"created"`
		Channel   string    `db:"channel" json:"channel"`
		ChannelID string    `db:"channel_id" json:"channelId"`
		Time      int64     `db:"time" json:"time"`
	}
)
