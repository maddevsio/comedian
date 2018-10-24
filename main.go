package main

import (
	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/notifier"
	"github.com/maddevsio/comedian/reporting"
	log "github.com/sirupsen/logrus"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}

	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}

	api, err := api.NewRESTAPI(slack)
	if err != nil {
		log.Fatal(err)
	}

	go func() { log.Fatal(api.Start()) }()

	r := reporting.NewReporter(slack)
	go func() { r.Start() }()

	notifier, err := notifier.NewNotifier(slack)
	if err != nil {
		log.Fatal(err)
	}

	go func() { log.Fatal(notifier.Start()) }()

	slack.Run()
}
