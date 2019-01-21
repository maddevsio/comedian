package main

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/api"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/notifier"
	"gitlab.com/team-monitoring/comedian/reporting"
	"gitlab.com/team-monitoring/comedian/sprint"
)

func main() {
	config, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := bot.NewBot(config)
	if err != nil {
		log.Fatal(err)
	}

	api, err := api.NewBotAPI(bot)
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(api.Start()) }()

	reporter, err := reporting.NewReporter(bot)
	if err != nil {
		log.Fatal(err)
	}
	go func() { reporter.Start() }()

	notifier, err := notifier.NewNotifier(bot)
	if err != nil {
		log.Fatal(err)
	}
	go func() { log.Fatal(notifier.Start()) }()

	sreporter := sprint.NewReporterSprint(bot)
	go func() { sreporter.Start() }()

	bot.Run()
}
