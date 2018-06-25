package notifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
)

// Notifier struct is used to notify users about upcoming or skipped standups
type Notifier struct {
	Chat          chat.Chat
	DB            storage.Storage
	CheckInterval uint64
}

// NewNotifier creates a new notifier
func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	return &Notifier{Chat: chat, DB: conn, CheckInterval: c.NotifierCheckInterval}, nil
}

// Start starts all notifier treads
func (n *Notifier) Start(c config.Config) error {
	log.Println("Starting notifier...")

	reportTimeParsed, err := time.Parse("15:04", c.ReportTime)
	if err != nil {
		log.Error(err)
	}
	fmt.Println("REPORT TIME PARSED", reportTimeParsed)
	config, err := config.Get()
	if err != nil {
		log.Error(err)
	}
	gocron.Every(n.CheckInterval).Seconds().Do(managerStandupReport, n.Chat, config, n.DB, reportTimeParsed)
	gocron.Every(n.CheckInterval).Seconds().Do(standupReminderForChannel, n.Chat, n.DB)
	channel := gocron.Start()
	for {
		report := <-channel
		log.Println(report)
	}
}

// standupReminderForChannel reminds users of channels about upcoming or missing standups
func standupReminderForChannel(chat chat.Chat, db storage.Storage) {
	standupTimes, err := db.ListAllStandupTime()
	if err != nil {
		log.Error(err)
	}
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		standupTime := time.Unix(standupTime.Time, 0)

		log.Printf("STANDUP TIME FOR CHANNEL: %s, IS: %v\n", channelID, standupTime)

		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			notifyStandupStart(chat, db, channelID)
		}
		pauseTime := time.Minute * 5 //repeats after n minutes
		repeatCount := 3             //repeat n times
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

// managerStandupReport reminds manager about missing or completed standups from channels
func managerStandupReport(chat chat.Chat, c config.Config, db storage.Storage, reportTimeParsed time.Time) {

	log.Printf("MANAGER CHANNEL: %s, REPORTING TIME: %v\n", c.DirectManagerChannelID, reportTimeParsed)
	currTime := time.Now()
	log.Printf("REPORT TIME H: %v, CURRENT TIME H: %v, REPORT TIME M: %v, CURRENT TIME M %v\n", reportTimeParsed.Hour(), currTime.Hour(), reportTimeParsed.Minute(), currTime.Minute())
	if reportTimeParsed.Hour() == currTime.Hour() && reportTimeParsed.Minute() == currTime.Minute() {
		// selects all users that have to do standup
		standupUsers, err := db.ListAllStandupUsers()
		if err != nil {
			log.Error(err)
		}

		// list usernames
		var allStandupUsernames []string
		for _, standupUser := range standupUsers {
			user := standupUser.SlackName
			allStandupUsernames = append(allStandupUsernames, user)
		}

		// list unique channels
		var standupChannelsList []string
		for _, standupUser := range standupUsers {
			channel := standupUser.ChannelID
			inList := false
			for i := 0; i < len(standupChannelsList); i++ {
				if standupChannelsList[i] == channel {
					inList = true
				}
			}
			if inList == false {
				standupChannelsList = append(standupChannelsList, channel)
			}
		}
		log.Printf("STANDUP CHANNELS: %v\n", strings.Join(standupChannelsList, ", "))

		// for each unique channel, create a separate msg report to manager
		for _, channel := range standupChannelsList {
			userStandupRaw, err := db.SelectStandupByChannelID(channel)
			if err != nil {
				log.Error(err)
			}
			var usersWhoCreatedStandup []string
			for _, userStandup := range userStandupRaw {
				user := userStandup.Username
				usersWhoCreatedStandup = append(usersWhoCreatedStandup, user)
			}
			var nonReporters []string
			for _, user := range allStandupUsernames {
				found := false
				for _, standupCreator := range usersWhoCreatedStandup {
					if user == standupCreator {
						found = true
						break
					}
				}
				if !found {
					nonReporters = append(nonReporters, "<@"+user+">")
				}
			}
			nonReportersCheck := len(nonReporters)
			if nonReportersCheck == 0 {
				err = chat.SendMessage(c.DirectManagerChannelID,
					fmt.Sprintf("<@%v>, in channel <#%s> all standupers have written standup today", c.Manager,
						channel))
				if err != nil {
					log.Errorf("ERROR: %s", err.Error())
				}
			} else {
				err = chat.SendMessage(c.DirectManagerChannelID,
					fmt.Sprintf("<@%v>, in channel <#%s> not all standupers wrote standup today, "+
						"this users ignored standup today: %v.",
						c.Manager, channel, strings.Join(nonReporters, ", ")))
				if err != nil {
					log.Errorf("ERROR: %s", err.Error())
				}
			}
		}
	}
}

// notifyStandupStart reminds users about upcoming standups
func notifyStandupStart(chat chat.Chat, db storage.Storage, channelID string) {
	standupUsers, err := db.ListStandupUsersByChannelID(channelID)
	if err != nil {
		log.Error(err)
	}
	var list []string
	for _, standupUser := range standupUsers {
		user := standupUser.SlackName
		list = append(list, "<@"+user+">")
	}

	err = chat.SendMessage(channelID,
		fmt.Sprintf("Hey! We are still waiting standup from you: %s", strings.Join(list, ", ")))
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}

}

// notifyNonReporters reminds users who missed deadlines about upcoming standups
func notifyNonReporters(chat chat.Chat, db storage.Storage, channelID string) error {
	standupUsersRaw, err := db.ListStandupUsersByChannelID(channelID)
	if err != nil {
		return err
	}
	var standupUsersList []string
	for _, standupUser := range standupUsersRaw {
		user := standupUser.SlackName
		standupUsersList = append(standupUsersList, user)
	}
	currentTime := time.Now()
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
			nonReporters = append(nonReporters, "<@"+user+">")
		}
	}
	nonReportersCheck := len(nonReporters)
	if nonReportersCheck == 0 {
		return chat.SendMessage(channelID, "Hey, in this channel all standupers have written standup today")
	}
	return chat.SendMessage(channelID,
		fmt.Sprintf("In this channel not all standupers wrote standup today, shame on you: %v.", strings.Join(nonReporters, ", ")))
}
