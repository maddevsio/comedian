package notifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/maddevsio/comedian/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/sirupsen/logrus"
)

// Notifier struct is used to notify users about upcoming or skipped standups
type Notifier struct {
	Chat          chat.Chat
	DB            storage.Storage
	Config        config.Config
	ReportTime    time.Time
	CheckInterval uint64
}

var localizer *i18n.Localizer

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
	r, err := time.Parse("15:04", c.ReportTime)
	if err != nil {
		logrus.Errorf("notifier: time.Parse failed: %v\n", err)
		return nil, err
	}
	notifier := &Notifier{Chat: chat, DB: conn, Config: c, ReportTime: r, CheckInterval: c.NotifierCheckInterval}
	logrus.Infof("notifier: Created Notifier: %v", notifier)
	return notifier, nil
}

// Start starts all notifier treads
func (n *Notifier) Start() error {
	gocron.Every(n.CheckInterval).Seconds().Do(n.NotifyChannels)
	channel := gocron.Start()
	for {
		report := <-channel
		logrus.Println(report)
	}
}

// NotifyChannels reminds users of channels about upcoming or missing standups
func (n *Notifier) NotifyChannels() {

	standupTimes, err := n.DB.ListAllStandupTime()
	if err != nil {
		logrus.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
	}
	// For each standup time, if standup time is now, start reminder
	for _, standupTime := range standupTimes {
		channelID := standupTime.ChannelID
		standupTime := time.Unix(standupTime.Time, 0)
		currTime := time.Now()
		if standupTime.Hour() == currTime.Hour() && standupTime.Minute() == currTime.Minute() {
			n.SendChannelNotification(channelID)
		}
	}
}

//SendChannelNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendChannelNotification(channelID string) {
	nonReporters, err := getNonReporters(n.DB, channelID)
	if err != nil {
		logrus.Errorf("notifier: getNonReporters failed: %v\n", err)
	}
	if len(nonReporters) == 0 {
		notifyAllDone, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyAllDone"})
		if err != nil {
			logrus.Errorf("notifier: Localize failed: %v\n", err)
		}
		n.Chat.SendMessage(channelID, notifyAllDone)
		return
	}

	n.SendWarning(channelID, nonReporters)
	n.DMNonReporters(nonReporters)

	nonReportersSlackIDs := []string{}
	for _, nonReporter := range nonReporters {
		nonReportersSlackIDs = append(nonReportersSlackIDs, nonReporter.SlackUserID)
	}
	logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

	notifyNotAll := func() error {
		// Comedian will notify non reporters 5 times with 30 minutes interval.
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyNotAll"})
		if err != nil {
			logrus.Errorf("notifier: Localize failed: %v\n", err)
			return err
		}
		n.Chat.SendMessage(channelID, fmt.Sprintf(text, strings.Join(nonReportersSlackIDs, ", ")))
		return nil
		logrus.Info("notifier: notifyNotAll finished!")
		return nil
	}
	for i := 0; i <= 5; i++ {
		b := backoff.NewConstantBackOff(30 * time.Minute)
		err = backoff.Retry(notifyNotAll, b)
		if err != nil {
			logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
		}
		if i == 3 {
			// after 3 reminders Comedian sends direct message to Manager notifiing about missed standups
			notifyManager, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyManagerNotAll"})
			if err != nil {
				logrus.Errorf("notifier: Localize failed: %v\n", err)
			}
			err = n.Chat.SendUserMessage(n.Config.ManagerSlackUserID, fmt.Sprintf(notifyManager, n.Config.ManagerSlackUserID, channelID, strings.Join(nonReportersSlackIDs, ", ")))
			if err != nil {
				logrus.Errorf("notifier: n.Chat.SendUserMessage failed: %v\n", err)
			}
		}
	}

}

// SendWarning reminds users in chat about upcoming standups
func (n *Notifier) SendWarning(channelID string, nonReporters []model.StandupUser) {
	slackUserID := []string{}
	for _, user := range nonReporters {
		slackUserID = append(slackUserID, "<@"+user.SlackUserID+">")
	}
	notifyUsersWarning, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyUsersWarning"})
	if err != nil {
		logrus.Errorf("notifier: Localize failed: %v\n", err)
	}
	err = n.Chat.SendMessage(channelID, fmt.Sprintf(notifyUsersWarning, strings.Join(slackUserID, ", ")))

}

// DMNonReporters writes DM to users who did not write standups
func (n *Notifier) DMNonReporters(nonReporters []model.StandupUser) error {
	//send each non reporter direct message
	for _, nonReporter := range nonReporters {
		logrus.Infof("notifier: Notifier Send Message to non reporter: %v", nonReporter)
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "notifyDirectMessage"})
		if err != nil {
			logrus.Errorf("notifier: Localize failed: %v\n", err)
		}
		n.Chat.SendUserMessage(nonReporter.SlackUserID, fmt.Sprintf(text, nonReporter.SlackName, nonReporter.ChannelID))
	}
	return nil
}

// getNonReporters returns a list of standupers that did not write standups
func getNonReporters(db storage.Storage, channelID string) ([]model.StandupUser, error) {
	currentTime := time.Now()
	timeFrom := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)

	nonReporters, err := db.GetNonReporters(channelID, timeFrom, currentTime)
	if err != nil {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}
