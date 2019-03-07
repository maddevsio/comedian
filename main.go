package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/api"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
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
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("active.en.toml")
	if err != nil {
		log.Fatalf("Load active.en.toml message file failed: %v", err)
	}
	_, err = bundle.LoadMessageFile("active.ru.toml")
	if err != nil {
		log.Fatalf("Load active.ru.toml message file failed: %v", err)
	}

	config, err := config.Get()
	if err != nil {
		log.Fatalf("Get config failed: %v", err)
	}

	db, err := storage.NewDB(config)
	if err != nil {
		log.Fatalf("NewDB failed: %v", err)
	}

	comedian := comedianbot.New(bundle, db)

	go comedian.StartBots()

	api, err := api.NewComedianAPI(config, db, comedian)
	if err != nil {
		log.Fatalf("NewComedianAPI failed: %v", err)
	}

	api.Start()
}
