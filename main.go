package main

import (
	"log"
	"net/http"

	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
)

func main() {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := api.GetHandler(c)
	if err != nil {
		log.Fatal(err)
	}
	go http.ListenAndServe(":8080", handler)

	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}
	slack.Run()
}
