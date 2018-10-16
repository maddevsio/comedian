package main

import (
	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/notifier"
	"github.com/maddevsio/comedian/teammonitoring"
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

	notifier, err := notifier.NewNotifier(slack)
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(notifier.Start()) }()

	//team monitoring servise is optional
	if c.TeamMonitoringEnabled {
		tm, err := teammonitoring.NewTeamMonitoring(slack)
		if err != nil {
			log.Fatal(err)
		}
		tm.Start()
	}

	slack.Run()
}
