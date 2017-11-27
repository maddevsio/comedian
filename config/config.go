package config

import "github.com/kelseyhightower/envconfig"

type (
	// Config struct used for configuration of app with env variables
	Config struct {
		SlackToken string `envconfig:"SLACK_TOKEN"`
		Debug      bool
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
