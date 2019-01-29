package main

import (
	"os"

	"github.com/evalphobia/logrus_sentry"
	raven "github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/api"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/notifier"
	"gitlab.com/team-monitoring/comedian/reporting"
	"gitlab.com/team-monitoring/comedian/sprint"
)

func init() {
	raven.SetSampleRate(0.25)
	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: false,
		PrettyPrint:      true,
	})
	log.SetReportCaller(true)

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

	sreporter := sprint.NewSprintReporter(bot)
	go func() { sreporter.Start() }()

	bot.Run()
}
