package main

import (
	"os"

	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/api"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
)

func init() {
	raven.SetSampleRate(0.25)

	hook, err := logrus_sentry.NewSentryHook(os.Getenv("SENTRY_DSN"), []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})

	if err == nil {
		log.AddHook(hook)
	}
}

func main() {
	config, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}

	comedian, err := comedianbot.New(config)
	if err != nil {
		log.Fatal(err)
	}
	go comedian.StartBots()
	api, err := api.NewComedianAPI(comedian)
	if err != nil {
		log.Fatal(err)
	}

	api.Start()
}
