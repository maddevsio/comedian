package main

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/api"
	"gitlab.com/team-monitoring/comedian/chat"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/notifier"
	"gitlab.com/team-monitoring/comedian/reporting"
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
