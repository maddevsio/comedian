package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

// Config struct used for configuration of app with env variables
type Config struct {
	DatabaseURL    string `envconfig:"DATABASE" required:"true" default:"comedian:comedian@/comedian?parseTime=true"`
	HTTPBindAddr   string `envconfig:"HTTP_BIND_ADDR" required:"true" default:"0.0.0.0:8080"`
	SlackToken     string `envconfig:"SLACK_TOKEN" required:"true"`
	CollectorURL   string `envconfig:"COLLECTOR_URL" required:"false"`
	CollectorToken string `envconfig:"COLLECTOR_TOKEN" required:"false"`
	SecretToken    string `envconfig:"SECRET_TOKEN" required:"false"`
	Login          string `envconfig:"LOGIN"`
	Password       string `envconfig:"PASSWORD"`
}

// Get method processes env variables and fills Config struct
func Get() (Config, error) {
	var c Config
	err := envconfig.Process("comedian", &c)
	if err != nil {
		return c, err
	}

	logrus.Info(c)

	return c, nil
}
