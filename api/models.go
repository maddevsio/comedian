package api

import "errors"

type (
	// AddStandup struct used for parsing form of add command from slack
	AddStandup struct {
		Command     string `schema:"command"`
		Text        string `schema:"text"`
		ChannelID   string `schema:"channel_id"`
		ChannelName string `schema:"channel_name"`
	}
	// Standup struct user for parsing form of list standups
	Standup struct {
		Command   string `schema:"command"`
		Text      string `schema:"text"`
		ChannelID string `schema:"channel_id"`
	}
	ChannelForm struct {
		Command   string `schema:"command"`
		ChannelID string `schema:"channel_id"`
	}
	ChannelNameForm struct {
		Command     string `schema:"command"`
		ChannelID   string `schema:"channel_id"`
		ChannelName string `schema:"channel_name"`
	}
)

// IsValid validates struct
func (c AddStandup) IsValid() error {
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

// IsValid validates struct
func (s Standup) IsValid() error {
	if s.Text == "" {
		return errors.New("`text` cannot be empty")
	}
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}

	return nil
}

// IsValid validates struct
func (s ChannelForm) IsValid() error {
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}

	return nil
}

// IsValid validates struct
func (s ChannelNameForm) IsValid() error {
	if s.ChannelID == "" {
		return errors.New("`channel_id` cannot be empty")
	}
	if s.ChannelName == "" {
		return errors.New("`channel_name` cannot be empty")
	}

	return nil
}
