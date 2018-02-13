package notifier

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	log "github.com/sirupsen/logrus"
	"time"
)

type Notifier struct {
	Chat chat.Chat
}

func NewNotifier(c config.Config, chat chat.Chat) *Notifier {
	return &Notifier{Chat: chat}
}

func (n *Notifier) Start() error {
	log.Println("Starting notifier...")
	gocron.Every(10).Seconds().Do(taskWithParams, n.Chat)
	channel := gocron.Start()
	for {
		report := <-channel
		log.Println(report)
	}
	return nil
}

func taskWithParams(chat chat.Chat) {
	currTime := time.Now()
	err := chat.SendMessage("D8DTA18UA", fmt.Sprintf("TEST MESSAGE! TIME: "))
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}

	log.Printf("%s -- %+v\n", currTime, chat)
}
