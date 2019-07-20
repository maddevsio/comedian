package main

import (
	"github.com/BurntSushi/toml"
	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

func main() {

	config, err := config.Get()
	if err != nil {
		log.Fatal("Failed to get config : ", err)
	}

	db, err := storage.New(config)
	if err != nil {
		log.Fatal("Failed to connect to db: ", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("active.en.toml")
	bundle.MustLoadMessageFile("active.ru.toml")

	api := api.New(config, db, bundle)

	if err = api.Start(); err != nil {
		log.Fatal("Failed to start API: ", err)
	}
}
