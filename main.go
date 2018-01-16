package main

import (
	"log"
	"net/http"

	"github.com/maddevsio/comedian/api"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
)

func main() {
	handler := api.GetHandler()
	go http.ListenAndServe(":8080", handler)

	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}
	slack.Run()
}
