package model

import (
	"time"
)

type (
	// Standup model used for serialization/deserialization stored standups
	Standup struct {
		ID        int64     `db:"id" json:"id"`
		Created   time.Time `db:"created" json:"created"`
		Modified  time.Time `db:"modified" json:"modified"`
		ChannelID string    `db:"channel_id" json:"channelId"`
		UserID    string    `db:"user_id" json:"userId"`
		Comment   string    `db:"comment" json:"comment"`
		MessageTS string    `db:"message_ts" json:"message_ts"`
	}

	// User model used for serialization/deserialization stored Users
	User struct {
		ID       int64  `db:"id" json:"id"`
		UserName string `db:"user_name" json:"user_name"`
		UserID   string `db:"user_id" json:"user_id"`
		Role     string `db:"role" json:"role"`
	}

	// Channel model used for serialization/deserialization stored Channels
	Channel struct {
		ID          int64  `db:"id" json:"id"`
		ChannelName string `db:"channel_name" json:"channel_name"`
		ChannelID   string `db:"channel_id" json:"channel_id"`
		StandupTime int64  `db:"channel_standup_time" json:"time"`
	}

	// ChannelMember model used for serialization/deserialization stored ChannelMembers
	ChannelMember struct {
		ID          int64  `db:"id" json:"id"`
		UserID      string `db:"user_id" json:"user_id"`
		ChannelID   string `db:"channel_id" json:"channel_id"`
		StandupTime int64  `db:"standup_time" json:"time"`
	}

	// StandupEditHistory model used for serialization/deserialization stored standup edit history
	StandupEditHistory struct {
		ID          int64     `db:"id" json:"id"`
		Created     time.Time `db:"created" json:"created"`
		StandupID   int64     `db:"standup_id" json:"standupId"`
		StandupText string    `db:"standuptext" json:"standuptext"`
	}
)
