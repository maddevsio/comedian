package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/notifier"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	api, err := api.NewRESTAPI(c)
	if err != nil {
		log.Fatal(err)
	}

	go func() { log.Fatal(api.Start()) }()

	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}

	notifier := notifier.NewNotifier(c, slack)
	go func() { log.Fatal(notifier.Start()) }()

	slack.Run()
}
