package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	conf, err := Get()
	assert.Error(t, err)

	os.Setenv("COMEDIAN_DATABASE", "DB")
	os.Setenv("COMEDIAN_HTTP_BIND_ADDR", "0.0.0.0:8080")
	os.Setenv("COMEDIAN_SLACK_TOKEN", "token")
	os.Setenv("COMEDIAN_COLLECTOR_TOKEN", "cotoken")
	os.Setenv("COMEDIAN_COLLECTOR_URL", "www.collector.some")

	conf, err = Get()
	assert.NoError(t, err)
	assert.Equal(t, conf.SlackToken, "token")
	assert.Equal(t, conf.DatabaseURL, "DB")
	assert.Equal(t, conf.HTTPBindAddr, "0.0.0.0:8080")
	assert.Equal(t, conf.CollectorToken, "cotoken")
	assert.Equal(t, conf.CollectorURL, "www.collector.some")
}
