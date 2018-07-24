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
	"github.com/sirupsen/logrus"
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
		logrus.Errorf("notifier: NewMySQL failed: %v\n", err)
		return nil, err
	}
	localizer, err = config.GetLocalizer()
	if err != nil {
		logrus.Errorf("notifier: GetLocalizer failed: %v\n", err)
		return nil, err
	}
	notifier := &Notifier{Chat: chat, DB: conn, CheckInterval: c.NotifierCheckInterval}
	logrus.Infof("notifier: Created Notifier: %v", notifier)
	return notifier, nil
}

// Start starts all notifier treads
func (n *Notifier) Start(c config.Config) error {

	reportTimeParsed, err := time.Parse("15:04", c.ReportTime)
	if err != nil {
		logrus.Errorf("notifier: time.Parse failed: %v\n", err)
		return err
	}
	logrus.Infof("notifier: Report time parsed: %v", reportTimeParsed)
	gocron.Every(n.CheckInterval).Seconds().Do(managerStandupReport, n.Chat, c, n.DB, reportTimeParsed)
	gocron.Every(n.CheckInterval).Seconds().Do(standupReminderForChannel, n.Chat, n.DB)
	channel := gocron.Start()
	for {
		report := <-channel
		logrus.Println(report)
	}
}

// standupReminderForChannel reminds users of channels about upcoming or missing standups
func standupReminderForChannel(chat chat.Chat, db storage.Storage) {
	standupTimes, err := db.ListAllStandupTime()
	if err != nil {
		logrus.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
	}
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		//channel := standupTime.Channel
		standupTime := time.Unix(standupTime.Time, 0)
		//logrus.Infof("notifier: Standup time for Channel: #%v is %v:%v\n", channel, standupTime.Hour(), standupTime.Minute())
		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			logrus.Infof("notifier: Standup time in Channel: %v, Current time: %v, IT's TIME TO DO A STANDUP", channelID, currTime)
			notifyStandupStart(chat, db, channelID)
			directRemindStandupers(chat, db, channelID)

			nonReporters, err := getNonReporters(db, channelID)
			if err != nil {
				logrus.Errorf("notifier: getNonReporters failed: %v\n", err)
			}
			logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)
			if len(nonReporters) != 0 {
				nonReportersSlackIDs := []string{}
				for _, nonReporter := range nonReporters {
					nonReportersSlackIDs = append(nonReportersSlackIDs, nonReporter.SlackUserID)
				}

				pauseTime := time.Minute * 1 //repeats after n minutes
				repeatCount := 5             //repeat n times
				logrus.Infof("notifier: Notifier pause time: %v", pauseTime)
				for i := 1; i <= repeatCount; i++ {
					logrus.Infof("notifier: Notifier repeated: %v times", i)
					notifyTime := standupTime.Add(pauseTime * time.Duration(i))
					if notifyTime.Hour() == currTime.Hour() && notifyTime.Minute() == currTime.Minute() {
						//periodic reminder for non reporters!
						text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyNotAll"})
						if err != nil {
							logrus.Errorf("notifier: Localize failed: %v\n", err)
						}
						chat.SendMessage(channelID, fmt.Sprintf(text, strings.Join(nonReportersSlackIDs, ", ")))
					}
				}
			} else {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyAllDone"})
				if err != nil {
					logrus.Errorf("notifier: Localize failed: %v\n", err)
				}
				chat.SendMessage(channelID, text)
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
			logrus.Errorf("notifier: ListAllStandupUsers failed: %v\n", err)
		}
		logrus.Infof("notifier: Notifier standup users: %v", standupUsers)

		// list unique usernames
		var allStandupUsernames []string
		for _, standupUser := range standupUsers {
			user := standupUser.SlackName
			inList := false
			for i := 0; i < len(allStandupUsernames); i++ {
				if allStandupUsernames[i] == user {
					inList = true
				}
			}
			if inList == false {
				allStandupUsernames = append(allStandupUsernames, user)
			}
		}
		logrus.Infof("notifier: Notifier all standup usernames: %v", allStandupUsernames)

		// list unique channels
		standupChannelsList, err := db.GetAllChannels()
		if err != nil {
			logrus.Errorf("notifier: GetAllChannels: %v\n", err)
		}

		logrus.Infof("notifier: Notifier standup chanels list: %v", standupChannelsList)

		// for each unique channel, create a separate msg report to manager
		for _, channel := range standupChannelsList {
			logrus.Infof("notifier: Notifier manager report for channel: %v", channel)

			nonReporters, err := getNonReporters(db, channel)
			if err != nil {
				logrus.Errorf("notifier: getNonReporters: %v\n", err)
			}
			logrus.Infof("notifier: Notifier manager report nonReporters: %v", nonReporters)
			nonReportersSlackIDs := []string{}
			for _, nonReporter := range nonReporters {
				nonReportersSlackIDs = append(nonReportersSlackIDs, "<@"+nonReporter.SlackUserID+">")
			}
			if len(nonReportersSlackIDs) == 0 {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyManagerAllDone"})
				if err != nil {
					logrus.Errorf("notifier: Localize failed: %v\n", err)
				}

				err = chat.SendUserMessage(c.ManagerSlackUserID, fmt.Sprintf(text, c.ManagerSlackUserID, channel))
				if err != nil {
					logrus.Errorf("notifier: chat.SendUserMessage failed: %v\n", err)
				}
			} else {
				text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyManagerNotAll"})
				if err != nil {
					logrus.Errorf("notifier: Localize failed: %v\n", err)
				}
				err = chat.SendUserMessage(c.ManagerSlackUserID, fmt.Sprintf(text, c.ManagerSlackUserID, channel, strings.Join(nonReportersSlackIDs, ", ")))
				if err != nil {
					logrus.Errorf("notifier: chat.SendUserMessage failed: %v\n", err)
				}
			}
		}
	}
}

// notifyStandupStart reminds users about upcoming standups
func notifyStandupStart(chat chat.Chat, db storage.Storage, channelID string) {
	listNonReporters, err := getNonReporters(db, channelID)
	if len(listNonReporters) == 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyAllDone"})
		if err != nil {
			logrus.Errorf("notifier: Localize failed: %v\n", err)
		}
		err = chat.SendMessage(channelID, text)
		return
	}
	logrus.Infof("notifier: ListNonReporters results: %v", listNonReporters)
	if err != nil {
		logrus.Errorf("notifier: getNonReporters failed: %v\n", err)
	}
	slackUserID := []string{}
	for _, sui := range listNonReporters {
		slackUserID = append(slackUserID, "<@"+sui.SlackUserID+">")
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyUsersWarning"})
	if err != nil {
		logrus.Errorf("notifier: Localize failed: %v\n", err)
	}
	err = chat.SendMessage(channelID, fmt.Sprintf(text, strings.Join(slackUserID, ", ")))

}

func getNonReporters(db storage.Storage, channelID string) ([]model.StandupUser, error) {
	currentTime := time.Now()
	timeFrom := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)

	nonReporters, err := db.ListNonReportersByTimeAndChannelID(channelID, timeFrom, currentTime)
	if err != nil {
		logrus.Errorf("notifier: ListNonReportersByTimeAndChannelID failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}

func directRemindStandupers(chat chat.Chat, db storage.Storage, channelID string) error {
	nonReporters, err := getNonReporters(db, channelID)
	if err != nil {
		logrus.Errorf("notifier: getNonReporters: %v\n", err)
	}
	logrus.Infof("notifier: Notifier GetNonReporters nonReporters: %v", nonReporters)
	for _, nonReporter := range nonReporters {
		logrus.Infof("notifier: Notifier Send Message to non reporter: %v", nonReporter)
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyDirectMessage"})
		if err != nil {
			logrus.Errorf("notifier: Localize failed: %v\n", err)
		}
		chat.SendUserMessage(nonReporter.SlackUserID, fmt.Sprintf(text, nonReporter.SlackName, nonReporter.ChannelID))
	}
	return nil
}
