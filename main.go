package main

import (
	"log"

	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
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
	slack.Run()
}
