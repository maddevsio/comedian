package api

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type (
	// FullSlackForm struct used for parsing full payload from slack
	FullSlackForm struct {
		Command     string `schema:"command"`
		Text        string `schema:"text"`
		ChannelID   string `schema:"channel_id"`
		ChannelName string `schema:"channel_name"`
	}
	// ChannelIDTextForm struct used for parsing text and channel_id param
	ChannelIDTextForm struct {
		Command   string `schema:"command"`
		Text      string `schema:"text"`
		ChannelID string `schema:"channel_id"`
	}
	// ChannelIDForm struct used for parsing channel_id param
	ChannelIDForm struct {
		Command   string `schema:"command"`
		ChannelID string `schema:"channel_id"`
	}
	// ChannelForm struct used for parsing channel_id and channel_name payload
	ChannelForm struct {
		Command     string `schema:"command"`
		ChannelID   string `schema:"channel_id"`
		ChannelName string `schema:"channel_name"`
	}
)

// Validate validates struct
func (c FullSlackForm) Validate() error {
	if c.Text == "" {
		err := errors.New("`text` cannot be empty")
		logrus.Errorf("api/models: FullSlackForm Validate failed: %s", err.Error())
		return err
	}
	if c.ChannelID == "" {
		err := errors.New("`channel_id` cannot be empty")
		logrus.Errorf("api/models: FullSlackForm Validate failed: %s", err.Error())
		return err
	}
	if c.ChannelName == "" {
		err := errors.New("`channel_name` cannot be empty")
		logrus.Errorf("api/models: FullSlackForm Validate failed: %s", err.Error())
		return err
	}
	return nil
}

// Validate validates struct
func (s ChannelIDTextForm) Validate() error {
	if s.Text == "" {
		err := errors.New("`text` cannot be empty")
		logrus.Errorf("api/models: ChannelIDTextForm Validate failed: %s", err.Error())
		return err
	}
	if s.ChannelID == "" {
		err := errors.New("`channel_id` cannot be empty")
		logrus.Errorf("api/models: ChannelIDTextForm Validate failed: %s", err.Error())
		return err
	}

	return nil
}

// Validate validates struct
func (s ChannelIDForm) Validate() error {
	if s.ChannelID == "" {
		err := errors.New("`channel_id` cannot be empty")
		logrus.Errorf("api/models: ChannelIDForm Validate failed: %s", err.Error())
		return err
	}

	return nil
}

// Validate validates struct
func (s ChannelForm) Validate() error {
	if s.ChannelID == "" {
		err := errors.New("`channel_id` cannot be empty")
		logrus.Errorf("api/models: ChannelForm Validate failed: %s", err.Error())
		return err
	}
	if s.ChannelName == "" {
		err := errors.New("`channel_name` cannot be empty")
		logrus.Errorf("api/models: ChannelForm Validate failed: %s", err.Error())
		return err
	}

	return nil
}
