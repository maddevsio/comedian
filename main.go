package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/comedianbot"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
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
		log.Fatal("Load active.en.toml message file failed: ", err)
	}
	_, err = bundle.LoadMessageFile("active.ru.toml")
	if err != nil {
		log.Fatal("Load active.ru.toml message file failed: ", err)
	}

	config, err := config.Get()
	if err != nil {
		log.Fatal("Get config failed: ", err)
	}

	db, err := storage.New(config)
	if err != nil {
		log.Fatal("New storage failed: ", err)
	}

	comedian := comedianbot.New(bundle, db)

	go comedian.StartBots()

	api := api.New(config, db, comedian)

	err = api.Start()
	if err != nil {
		log.Fatal("API start failed: ", err)
	}
}
