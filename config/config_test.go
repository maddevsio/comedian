package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Setenv("SLACK_TOKEN", "Hey_token")
	conf, err := Get()
	assert.NoError(t, err)
	assert.Equal(t, conf.SlackToken, "Hey_token")
	os.Setenv("COMEDIAN_DEBUG", "string")
	conf, err = Get()
	assert.Error(t, err)
}
