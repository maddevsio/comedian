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

var nowFunc func() time.Time

var managerStandupChannelID = "PUT REAL CHANNEL ID HERE"

type Notifier struct {
	Chat          chat.Chat
	DB            storage.Storage
	CheckInterval uint64
}

func init() {
	nowFunc = func() time.Time {
		return time.Now()
	}
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

	//todo: refactor and delete
	manager := "@managerName"
	directManagerChannelID := "D8DTA18UA"
	reportTime := "09:35"
	reportTimeParsed, err := time.Parse("15:04", reportTime)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("REPORT TIME PARSED", reportTimeParsed)
	//todo: end

	gocron.Every(n.CheckInterval).Seconds().Do(standupReminderForChannel, n.Chat, n.DB)
	gocron.Every(n.CheckInterval).Seconds().Do(managerStandupReport, n.Chat, n.DB, manager,
		directManagerChannelID, reportTimeParsed)
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

		currTime := now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			notifyStandupStart(chat, db, channelID)
		}
		pauseTime := time.Minute * 30
		repeatCount := 3
		for i := 1; i <= repeatCount; i++ {
			notifyTime := standupTime.Add(pauseTime * time.Duration(i))
			if notifyTime.Hour() == currTime.Hour() && notifyTime.Minute() == currTime.Minute() {
				//periodic remind
				err = notifyNonReporters(chat, db, channelID)
				if err != nil {
					log.Error(err)
				}
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

	err = chat.SendMessage(channelID,
		fmt.Sprintf("Hey! We are still waiting standup from you: %s", strings.Join(list, ", ")))
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}

}

func managerStandupReport(chat chat.Chat, db storage.Storage, manager, directManagerChannelID string, reportTimeParsed time.Time) {
	log.Printf("CHANNEL: %s, TIME: %v\n", managerStandupChannelID, reportTimeParsed)
	currTime := now()
	if reportTimeParsed.Hour() == currTime.Hour() && reportTimeParsed.Minute() == currTime.Minute() {
		standupUsersRaw, err := db.ListStandupUsers(managerStandupChannelID)
		if err != nil {
			log.Error(err)
		}
		var standupUsersList []string
		for _, standupUser := range standupUsersRaw {
			user := standupUser.SlackName
			standupUsersList = append(standupUsersList, user)
		}
		userStandupRaw, err := db.SelectStandupByChannelID(managerStandupChannelID)
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
					managerStandupChannelID))
			if err != nil {
				log.Errorf("ERROR: %s", err.Error())
			}
		} else {
			err = chat.SendMessage(directManagerChannelID,
				fmt.Sprintf("%v, in channel '%s' not all standupers wrote standup today, "+
					"this users ignored standup today: %v.",
					manager, managerStandupChannelID, strings.Join(nonReporters, ", ")))
			if err != nil {
				log.Errorf("ERROR: %s", err.Error())
			}
		}
	}
}

func notifyNonReporters(chat chat.Chat, db storage.Storage, channelID string) error {
	standupUsersRaw, err := db.ListStandupUsers(channelID)
	if err != nil {
		return err
	}
	var standupUsersList []string
	for _, standupUser := range standupUsersRaw {
		user := standupUser.SlackName
		standupUsersList = append(standupUsersList, user)
	}
	currentTime := now()
	dateStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := dateStart.Add(time.Hour * 24)
	userStandupRaw, err := db.SelectStandupByChannelIDForPeriod(channelID, dateStart, dateEnd)

	if err != nil {
		return err
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
		return chat.SendMessage(channelID, "Hey, in this channel all standupers have written standup today")
	} else {
		return chat.SendMessage(channelID,
			fmt.Sprintf("In this channel not all standupers wrote standup today, "+
				"shame on you: %v.", strings.Join(nonReporters, ", ")))
	}
}

func now() time.Time {
	return nowFunc().UTC()
}
