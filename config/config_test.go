package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Setenv("COMEDIAN_DATABASE", "DB")
	os.Setenv("COMEDIAN_HTTP_BIND_ADDR", "0.0.0.0:8080")
	os.Setenv("COMEDIAN_NOTIFIER_CHECK_INTERVAL", "15")
	os.Setenv("COMEDIAN_MANAGER", "manager")
	os.Setenv("COMEDIAN_DIRECT_MANAGER_CHANNEL_ID", "someID")
	os.Setenv("COMEDIAN_REPORT_TIME", "17:00")
	os.Setenv("COMEDIAN_DEBUG", "true")
	os.Setenv("COMEDIAN_SLACK_TOKEN", "token")

	conf, err := Get()
	assert.NoError(t, err)
	assert.Equal(t, conf.SlackToken, "token")
	assert.Equal(t, conf.DatabaseURL, "DB")
	assert.Equal(t, conf.HTTPBindAddr, "0.0.0.0:8080")
	assert.Equal(t, conf.NotifierCheckInterval, uint64(15))
	assert.Equal(t, conf.Manager, "manager")
	assert.Equal(t, conf.DirectManagerChannelID, "someID")
	assert.Equal(t, conf.ReportTime, "17:00")
	assert.Equal(t, conf.Debug, true)
}
