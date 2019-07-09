package main

import (
	"github.com/BurntSushi/toml"
	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/comedianbot"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

func main() {
	bundle := i18n.NewBundle(language.English)
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
