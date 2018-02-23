package notifier

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Notifier struct {
	Chat          chat.Chat
	DB            storage.Storage
	CheckInterval uint64
}

func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	return &Notifier{Chat: chat, DB: conn, CheckInterval: c.NotifierCheckInterval}, nil
}

func (n *Notifier) Start() error {
	log.Println("Starting notifier...")
	gocron.Every(n.CheckInterval).Seconds().Do(standupReminderForChannel, n.Chat, n.DB)
	gocron.Every(n.CheckInterval).Seconds().Do(managerStandupReport, n.Chat, n.DB)
	channel := gocron.Start()
	for {
		report := <-channel
		log.Println(report)
	}
	return nil
}

func standupReminderForChannel(chat chat.Chat, db storage.Storage) {
	standupTimes, err := db.ListAllStandupTime()
	if err != nil {
		log.Error(err)
	}
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		standupTime := time.Unix(standupTime.Time, 0)

		log.Printf("CHANNEL: %s, TIME: %v\n", channelID, standupTime)

		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			notifyStandupStart(chat, db, channelID)
		}
		pauseTime := time.Minute * 30
		repeatCount := 3
		for i := 1; i <= repeatCount; i++ {
			notifyTime := standupTime.Add(pauseTime * time.Duration(i))
			if notifyTime.Hour() == currTime.Hour() && notifyTime.Minute() == currTime.Minute() {
				//periodic remind
				notifyNonReporters(chat, db, channelID)
			}
		}
	}
}

func notifyStandupStart(chat chat.Chat, db storage.Storage, channelID string) {
	standupUsers, err := db.ListStandupUsers(channelID)
	if err != nil {
		log.Error(err)
	}
	var list []string
	for _, standupUser := range standupUsers {
		user := standupUser.SlackName
		list = append(list, user)
	}
	fmt.Printf("USERS : %s", list)

	err = chat.SendMessage(channelID,
		fmt.Sprintf("Hey! We are still waiting standup from you: %s", strings.Join(list, ", ")))
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}

}

func managerStandupReport(chat chat.Chat, db storage.Storage) {
	manager := "@managerName"
	directManagerChannelID := "D8DTA18UA"
	directManagerChannelName := "Space Y"
	reportTime := "09:35"
	reportTimeParsed, err := time.Parse("15:04", reportTime)
	if err != nil {
		log.Error(err)
	}
	log.Printf("CHANNEL: %s, TIME: %v\n", directManagerChannelID, reportTime)
	currTime := time.Now()
	if reportTimeParsed.Hour() == currTime.Hour() && reportTimeParsed.Minute() == currTime.Minute() {
		standupUsersRaw, err := db.ListStandupUsers(directManagerChannelID)
		if err != nil {
			log.Error(err)
		}
		var standupUsersList []string
		for _, standupUser := range standupUsersRaw {
			user := standupUser.SlackName
			standupUsersList = append(standupUsersList, user)
		}
		userStandupRaw, err := db.SelectStandupByChannelID(directManagerChannelID)
		if err != nil {
			log.Error(err)
		}
		var usersWhoCreatedStandup []string
		for _, userStandup := range userStandupRaw {
			user := userStandup.Username
			usersWhoCreatedStandup = append(usersWhoCreatedStandup, user)
		}
		var nonReporters []string
		for _, user := range standupUsersList {
			found := false
			for _, standupCreator := range usersWhoCreatedStandup {
				if user == standupCreator {
					found = true
					break
				}
			}
			if !found {
				nonReporters = append(nonReporters, user)
			}
		}
		nonReportersCheck := len(nonReporters)
		if nonReportersCheck == 0 {
			err = chat.SendMessage(directManagerChannelID,
				fmt.Sprintf("%v, in channel %s all standupers have written standup today", manager,
					directManagerChannelName))
			if err != nil {
				log.Errorf("ERROR: %s", err.Error())
			}
		} else {
			err = chat.SendMessage(directManagerChannelID,
				fmt.Sprintf("%v, in channel '%s' not all standupers wrote standup today, "+
					"this users ignored standup today: %v.",
					manager, directManagerChannelName, strings.Join(nonReporters, ", ")))
			if err != nil {
				log.Errorf("ERROR: %s", err.Error())
			}
		}
	}
}

func notifyNonReporters(chat chat.Chat, db storage.Storage, channelID string) {
	standupUsersRaw, err := db.ListStandupUsers(channelID)
	if err != nil {
		log.Error(err)
	}
	var standupUsersList []string
	for _, standupUser := range standupUsersRaw {
		user := standupUser.SlackName
		standupUsersList = append(standupUsersList, user)
	}
	userStandupRaw, err := db.SelectStandupByChannelID(channelID)
	if err != nil {
		log.Error(err)
	}
	var usersWhoCreatedStandup []string
	for _, userStandup := range userStandupRaw {
		user := userStandup.Username
		usersWhoCreatedStandup = append(usersWhoCreatedStandup, user)
	}
	var nonReporters []string
	for _, user := range standupUsersList {
		found := false
		for _, standupCreator := range usersWhoCreatedStandup {
			if user == standupCreator {
				found = true
				break
			}
		}
		if !found {
			nonReporters = append(nonReporters, user)
		}
	}
	nonReportersCheck := len(nonReporters)
	if nonReportersCheck == 0 {
		err = chat.SendMessage(channelID, "Hey, in this channel all standupers have written standup today")
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
		}
	} else {
		err = chat.SendMessage(channelID,
			fmt.Sprintf("In this channel not all standupers wrote standup today, "+
				"shame on you: %v.", strings.Join(nonReporters, ", ")))
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
		}
	}
}
