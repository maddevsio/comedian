package model

import (
	"errors"
	"time"
)

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
		Comment    string    `db:"comment" json:"comment"`
		MessageTS  string    `db:"message_ts" json:"message_ts"`
	}

	// StandupUser model used for serialization/deserialization stored standupUsers
	StandupUser struct {
		ID          int64     `db:"id" json:"id"`
		Created     time.Time `db:"created" json:"created"`
		Modified    time.Time `db:"modified" json:"modified"`
		SlackUserID string    `db:"slack_user_id" json:"slack_user_id"`
		SlackName   string    `db:"username" json:"username"`
		Channel     string    `db:"channel" json:"channel"`
		ChannelID   string    `db:"channel_id" json:"channelId"`
		Role        string    `db:"role" json:"role"`
	}

	// StandupTime model used for serialization/deserialization stored standupTime
	StandupTime struct {
		ID        int64     `db:"id" json:"id"`
		Created   time.Time `db:"created" json:"created"`
		Channel   string    `db:"channel" json:"channel"`
		ChannelID string    `db:"channel_id" json:"channelId"`
		Time      int64     `db:"standuptime" json:"time"`
	}

	// StandupEditHistory model used for serialization/deserialization stored standup edit history
	StandupEditHistory struct {
		ID          int64     `db:"id" json:"id"`
		Created     time.Time `db:"created" json:"created"`
		StandupID   int64     `db:"standup_id" json:"standupId"`
		StandupText string    `db:"standuptext" json:"standuptext"`
	}
)

// Validate validates Standup struct
func (c Standup) Validate() error {
	return nil
}

// Validate validates StandupUser struct
func (c StandupUser) Validate() error {
	if c.SlackName == "" && c.SlackUserID == "" {
		err := errors.New("User cannot be empty")
		return err
	}
	return nil
}

// Validate validates StandupTime struct
func (c StandupTime) Validate() error {
	if c.Time == 0 {
		err := errors.New("Time cannot be empty")
		return err
	}
	return nil
}

// Validate validates StandupTimeHistory struct
func (c StandupEditHistory) Validate() error {
	if c.StandupText == "" {
		err := errors.New("Text cannot be empty")
		return err
	}
	return nil
}
