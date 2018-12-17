package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config struct used for configuration of app with env variables
type Config struct {
	DatabaseURL    string
	HTTPBindAddr   string
	SlackToken     string
	CollectorURL   string
	CollectorToken string

	NotifierInterval   int
	ManagerSlackUserID string
	ReportingChannel   string
	ReportTime         string
	Language           string
	ReminderRepeatsMax int
	ReminderTime       int64
	CollectorEnabled   bool
	SecretToken        string

	Translate Translate
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
