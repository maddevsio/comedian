package config

import "github.com/kelseyhightower/envconfig"

type (
	// Config struct used for configuration of app with env variables
	Config struct {
		SlackToken            string `envconfig:"SLACK_TOKEN" required:"true"`
		DatabaseURL           string `envconfig:"DATABASE" required:"true"`
		HTTPBindAddr          string `envconfig:"HTTP_BIND_ADDR" required:"true"`
		NotifierCheckInterval uint64 `envconfig:"NOTIFIER_CHECK_INTERVAL" required:"true"`
		Debug                 bool
	}
)

// Get method processes env variables and fills Config struct
func Get() (Config, error) {
	var c Config
	if err := envconfig.Process("comedian", &c); err != nil {
		return c, err
	}
	return c, nil
}
