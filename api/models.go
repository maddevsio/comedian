package api

import "errors"

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
		return errors.New("`text` cannot be empty")
	}
	if c.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}
	if c.ChannelName == "" {
		return errors.New("`channel_name` cannot be empty")
	}
	return nil
}

// Validate validates struct
func (s ChannelIDTextForm) Validate() error {
	if s.Text == "" {
		return errors.New("`text` cannot be empty")
	}
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}

	return nil
}

// Validate validates struct
func (s ChannelIDForm) Validate() error {
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}

	return nil
}

// Validate validates struct
func (s ChannelForm) Validate() error {
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}
	if s.ChannelName == "" {
		return errors.New("`channel_name` cannot be empty")
	}

	return nil
}
