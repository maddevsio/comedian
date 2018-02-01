package chat

import (
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/stretchr/testify/assert"
)

func TestCleanMessage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := NewSlack(c)
	assert.NoError(t, err)
	s.myUsername = "comedian"
	assert.NoError(t, err)
	text, ok := s.cleanMessage("<@comedian> hey there")
	assert.Equal(t, "hey there", text)
	assert.True(t, ok)
	text, ok = s.cleanMessage("<@comedian> What's up?")
	assert.Equal(t, "What's up?", text)
	assert.True(t, ok)
}
