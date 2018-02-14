package notifier

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/storage"
	"github.com/maddevsio/comedian/config"
	log "github.com/sirupsen/logrus"
	"time"
)

type Notifier struct {
	Chat chat.Chat
	DB   storage.Storage
}

func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	return &Notifier{Chat: chat, DB: conn}, nil
}

func (n *Notifier) Start() error {
	log.Println("Starting notifier...")
	gocron.Every(10).Seconds().Do(taskWithParams, n.Chat, n.DB)
	channel := gocron.Start()
	for {
		report := <-channel
		log.Println(report)
	}
	return nil
}

func taskWithParams(chat chat.Chat, db storage.Storage) {
	standupTimes, err := db.ListAllStandupTime()
	if err != nil {
		log.Println(err)
	}
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		standupTime := time.Unix(standupTime.Time, 0)

		log.Printf("CHANNEL: %s, TIME: %v\n", channelID, standupTime)
		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			err := chat.SendMessage(channelID,
				fmt.Sprintf("TEST MESSAGE! CURRTIME: %v, STANDUPTIME: %v", currTime, standupTime))
			if err != nil {
				log.Printf("ERROR: %s", err.Error())
			}
		}
	}
}
