package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct used for configuration of app with env variables
type Config struct {
	SlackToken            string `envconfig:"SLACK_TOKEN" required:"true"`
	DatabaseURL           string `envconfig:"DATABASE" required:"false" default:"comedian:comedian@/comedian?parseTime=true"`
	HTTPBindAddr          string `envconfig:"HTTP_BIND_ADDR" required:"false" default:"0.0.0.0:8080"`
	NotifierInterval      int    `envconfig:"NOTIFIER_INTERVAL" required:"false" default:2`
	ManagerSlackUserID    string `envconfig:"MANAGER_SLACK_USER_ID" required:"true"`
	ReportingChannel      string `envconfig:"REPORTING_CHANNEL" required:"false"`
	ReportTime            string `envconfig:"REPORT_TIME" required:"false" default:"13:05"`
	Language              string `envconfig:"LANGUAGE" required:"false" default:"en_US"`
	ReminderRepeatsMax    int    `envconfig:"REMINDER_REPEATS_MAX" required:"false" default:5`
	ReminderTime          int64  `envconfig:"REMINDER_TIME" required:"false" default:5`
	TeamMonitoringEnabled bool   `envconfig:"ENABLE_TEAM_MONITORING" required:"false" default:false`
	CollectorURL          string `envconfig:"COLLECTOR_URL" required:"false"`
	CollectorToken        string `envconfig:"COLLECTOR_TOKEN" required:"false"`
	Translate             Translate
}

// Get method processes env variables and fills Config struct
func Get() (Config, error) {
	var c Config
	err := envconfig.Process("comedian", &c)
	if err != nil {
		return c, err
	}
	t, err := GetTranslation(c.Language)
	if err != nil {
		return c, err
	}
	c.Translate = t
	return c, nil
}
