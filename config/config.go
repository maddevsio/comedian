package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct used for configuration of app with env variables
type Config struct {
	DatabaseURL       string `envconfig:"DATABASE" required:"false" default:"comedian:comedian@/comedian?parseTime=true"`
	HTTPBindAddr      string `envconfig:"HTTP_BIND_ADDR" required:"false" default:"0.0.0.0:8080"`
	TestEnvironment   bool   `envconfig:"TESTENVIRONMENT" required:"false" default:"false"`
	SlackClientID     string `envconfig:"SLACK_CLIENT_ID" required:"true"`
	SlackClientSecret string `envconfig:"SLACK_CLIENT_SECRET" required:"true"`
}

// Get method processes env variables and fills Config struct
func Get() (Config, error) {
	var c Config
	err := envconfig.Process("comedian", &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
