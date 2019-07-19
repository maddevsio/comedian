package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	conf, err := Get()
	assert.NoError(t, err)

	os.Setenv("DATABASE", "DB")
	os.Setenv("HTTP_BIND_ADDR", "0.0.0.0:8080")
	os.Setenv("SLACK_CLIENT_ID", "ID")
	os.Setenv("SLACK_CLIENT_SECRET", "SECRET")

	conf, err = Get()
	assert.NoError(t, err)
	assert.Equal(t, conf.DatabaseURL, "DB")
	assert.Equal(t, conf.HTTPBindAddr, "0.0.0.0:8080")
	assert.Equal(t, conf.SlackClientID, "ID")
	assert.Equal(t, conf.SlackClientSecret, "SECRET")

}
