package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct used for configuration of app with env variables
type Config struct {
	DatabaseURL            string `envconfig:"DATABASE" required:"false" default:"comedian:comedian@/comedian?parseTime=true"`
	CollectorURL           string `envconfig:"COLLECTOR_URL" required:"false" default:""`
	CollectorToken         string `envconfig:"COLLECTOR_TOKEN" required:"false" default:""`
	HTTPBindAddr           string `envconfig:"HTTP_BIND_ADDR" required:"false" default:"0.0.0.0:8080"`
	SlackClientID          string `envconfig:"SLACK_CLIENT_ID" required:"false"`
	SlackClientSecret      string `envconfig:"SLACK_CLIENT_SECRET" required:"false"`
	SlackVerificationToken string `envconfig:"SLACK_VERIFICATION_TOKEN" required:"false"`
	UIurl                  string `envconfig:"UI_URL" required:"false"`
	OwnerSlackTeamID       string `envconfig:"OWNER_SLACK_TEAM_ID" required:"false"`
}

// Get method processes env variables and fills Config struct
func Get() (*Config, error) {
	c := &Config{}
	err := envconfig.Process("comedian", c)
	return c, err
}
