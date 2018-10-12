package api

import (
	"errors"
)

// FullSlackForm struct used for parsing full payload from slack
type FullSlackForm struct {
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ChannelID   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
}

// Validate validates struct
func (c FullSlackForm) Validate() error {
	if c.ChannelID == "" || c.ChannelName == "" {
		return errors.New("I cannot understand which channel I am in. Please, double check if I am invited (or reinvite me one more time) to the channel and try again. Thank you!")
	}
	return nil
}
