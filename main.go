package main

import (
	"log"

	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
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

	go log.Fatal(api.Start())

	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}
	slack.Run()
}
