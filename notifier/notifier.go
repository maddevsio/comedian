package notifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
)

var localizer *i18n.Localizer

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
		log.Errorf("ERROR: %s", err.Error())
		return nil, err
	}
	localizer, err = config.GetLocalizer()
	if err != nil {
		return nil, err
	}
	return &Notifier{Chat: chat, DB: conn, CheckInterval: c.NotifierCheckInterval}, nil
}

// Start starts all notifier treads
func (n *Notifier) Start(c config.Config) error {

	reportTimeParsed, err := time.Parse("15:04", c.ReportTime)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
		return err
	}
	gocron.Every(n.CheckInterval).Seconds().Do(managerStandupReport, n.Chat, c, n.DB, reportTimeParsed)
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
		log.Errorf("ERROR: %s", err.Error())
	}
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		standupTime := time.Unix(standupTime.Time, 0)
		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			notifyStandupStart(chat, db, channelID)
			directRemindStandupers(chat, db, channelID)
		}
		nonReporters, err := getNonReporters(chat, db, channelID)
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
		}
		if len(nonReporters) > 0 {
			pauseTime := time.Minute * 1 //repeats after n minutes
			repeatCount := 5             //repeat n times
			for i := 1; i <= repeatCount; i++ {
				notifyTime := standupTime.Add(pauseTime * time.Duration(i))
				if notifyTime.Hour() == currTime.Hour() && notifyTime.Minute() == currTime.Minute() {
					//periodic reminder for non reporters!
					text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyNotAll"})
					if err != nil {
						log.Fatal(err)
					}
					chat.SendMessage(channelID, fmt.Sprintf(text, strings.Join(nonReporters, ", ")))
				}
			}
		}

	}
}

// managerStandupReport reminds manager about missing or completed standups from channels
func managerStandupReport(chat chat.Chat, c config.Config, db storage.Storage, reportTimeParsed time.Time) {
	currTime := time.Now()
	if reportTimeParsed.Hour() == currTime.Hour() && reportTimeParsed.Minute() == currTime.Minute() {
		// selects all users that have to do standup
		standupUsers, err := db.ListAllStandupUsers()
		if err != nil {
			log.Errorf("ERROR: %s", err.Error())
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
		// for each unique channel, create a separate msg report to manager
		for _, channel := range standupChannelsList {
			userStandupRaw, err := db.SelectStandupsByChannelID(channel)
			if err != nil {
				log.Errorf("ERROR: %s", err.Error())
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
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyManagerAllDone"})
				if err != nil {
					log.Fatal(err)
				}
				err = chat.SendMessage(c.DirectManagerChannelID, fmt.Sprintf(text, c.Manager, channel))
				if err != nil {
					log.Errorf("ERROR: %s", err.Error())
				}
			} else {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyManagerNotAll"})
				if err != nil {
					log.Fatal(err)
				}
				err = chat.SendMessage(c.DirectManagerChannelID, fmt.Sprintf(text, c.Manager, channel, strings.Join(nonReporters, ", ")))
				if err != nil {
					log.Errorf("ERROR: %s", err.Error())
				}
			}
		}
	}
}

// notifyStandupStart reminds users about upcoming standups
func notifyStandupStart(chat chat.Chat, db storage.Storage, channelID string) {
	list, err := getNonReporters(chat, db, channelID)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}
	if len(list) > 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyUsersWarning"})
		if err != nil {
			log.Fatal(err)
		}
		err = chat.SendMessage(channelID, fmt.Sprintf(text, strings.Join(list, ", ")))
	} else {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyAllDone"})
		if err != nil {
			log.Fatal(err)
		}
		err = chat.SendMessage(channelID, text)
	}

}

func getNonReporters(chat chat.Chat, db storage.Storage, channelID string) ([]string, error) {
	standupUsersRaw, err := db.ListStandupUsersByChannelID(channelID)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}
	var standupUsersList []string
	for _, standupUser := range standupUsersRaw {
		user := standupUser.SlackName
		standupUsersList = append(standupUsersList, user)
	}
	currentTime := time.Now()
	dateStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := dateStart.Add(time.Hour * 24)
	userStandupRaw, err := db.SelectStandupsByChannelIDForPeriod(channelID, dateStart, dateEnd)

	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
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
	return nonReporters, nil
}

func directRemindStandupers(chat chat.Chat, db storage.Storage, channelID string) error {
	standupUsers, err := db.ListStandupUsersByChannelID(channelID)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}
	currentTime := time.Now()
	dateStart := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := dateStart.Add(time.Hour * 24)

	standups, err := db.SelectStandupsByChannelIDForPeriod(channelID, dateStart, dateEnd)
	if err != nil {
		log.Errorf("ERROR: %s", err.Error())
	}
	var usersWhoCreatedStandup []string
	for _, userStandup := range standups {
		user := userStandup.Username
		usersWhoCreatedStandup = append(usersWhoCreatedStandup, user)
	}
	var nonReporters []model.StandupUser
	for _, user := range standupUsers {
		found := false
		for _, standupCreator := range usersWhoCreatedStandup {
			if user.SlackName == standupCreator {
				found = true
				break
			}
		}
		if !found {
			nonReporters = append(nonReporters, user)
		}
	}
	for _, motherFucker := range nonReporters {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyDirectMessage"})
		if err != nil {
			log.Fatal(err)
		}
		chat.SendUserMessage(motherFucker.SlackUserID, fmt.Sprintf(text, motherFucker.SlackName, motherFucker.ChannelID))
	}
	return nil
}
