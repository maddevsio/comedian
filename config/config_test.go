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
	os.Setenv("COMEDIAN_NOTIFIER_CHECK_INTERVAL", "15")
	os.Setenv("COMEDIAN_MANAGER_SLACK_USER_ID", "FAKEUSERID")
	os.Setenv("COMEDIAN_REPORT_TIME", "17:00")
	os.Setenv("COMEDIAN_DEBUG", "true")
	os.Setenv("COMEDIAN_SLACK_TOKEN", "token")
	os.Setenv("COMEDIAN_LANGUAGE", "ru_RU")
	os.Setenv("COMEDIAN_COLLECTOR_TOKEN", "cotoken")
	os.Setenv("COMEDIAN_COLLECTOR_URL", "www.collector.some")
	os.Setenv("COMEDIAN_MANAGER_SLACK_CHAN_GENERAL", "XXXYYYZZZ")

	conf, err = Get()
	assert.NoError(t, err)
	assert.Equal(t, conf.SlackToken, "token")
	assert.Equal(t, conf.DatabaseURL, "DB")
	assert.Equal(t, conf.HTTPBindAddr, "0.0.0.0:8080")
	assert.Equal(t, conf.NotifierCheckInterval, uint64(15))
	assert.Equal(t, conf.ManagerSlackUserID, "FAKEUSERID")
	assert.Equal(t, conf.ReportTime, "17:00")
	assert.Equal(t, conf.Language, "ru_RU")
	assert.Equal(t, conf.CollectorToken, "cotoken")
	assert.Equal(t, conf.CollectorURL, "www.collector.some")
	assert.Equal(t, conf.ChanGeneral, "XXXYYYZZZ")
	assert.Equal(t, conf.Debug, true)
}

func TestGetLocalizer(t *testing.T) {
	localizer, err := GetLocalizer()
	assert.NoError(t, err)
	assert.NotNil(t, localizer)
}
