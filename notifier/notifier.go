package notifier

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/reporting"
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

// MorningNotification (MN)
type MN struct {
	noWorklogs string
	noCommits  string
	noStandup  string

	hasWorklogs string
	hasCommits  string
	hasStandup  string

	isRook string
}

const (
	NotificationInterval   = 30
	ReminderRepeatsMax     = 5
	RemindManager          = 3
	MorningRooksReportTime = "10:30"
)

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
	gocron.Every(1).Day().At(MorningRooksReportTime).Do(n.RevealRooks)
	channel := gocron.Start()
	for {
		report := <-channel
		logrus.Println(report)
	}
}

func (n *Notifier) RevealRooks() {
	currentTime := time.Now()
	timeFrom := currentTime.AddDate(0, 0, -1)
	allUsers, err := n.DB.ListAllStandupUsers()
	if err != nil {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
	}
	text := ""
	for _, user := range allUsers {
		notification, err := getMN()
		if err != nil {
			logrus.Errorf("notifier: GetMN failed: %v\n", err)
		}
		worklogs, commits, isNonReporter, err := n.checkUser(user, timeFrom, currentTime)
		fails := ""
		if err != nil {
			logrus.Errorf("notifier: checkMotherFucker failed: %v\n", err)
		}
		if worklogs < 8 {
			fails += fmt.Sprintf(notification.noWorklogs, worklogs) + ", "
		} else {
			fails += fmt.Sprintf(notification.hasWorklogs, worklogs) + ", "
		}
		if commits == 0 {
			fails += notification.noCommits
		} else {
			fails += fmt.Sprintf(notification.hasCommits, commits) + ", "
		}
		if isNonReporter == true {
			fails += notification.noStandup
		} else {
			fails += notification.hasStandup
		}
		if (worklogs < 8) || (commits == 0) || (isNonReporter == true) {
			text += fmt.Sprintf(notification.isRook, user.SlackUserID, fails)
		}
	}
	n.Chat.SendMessage(n.Config.ChanGeneral, text)

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
	for i := 0; i <= ReminderRepeatsMax; i++ {
		b := backoff.NewConstantBackOff(NotificationInterval * time.Minute)
		err = backoff.Retry(notifyNotAll, b)
		if err != nil {
			logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
		}
		if i == RemindManager {
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

func (n *Notifier) checkUser(user model.StandupUser, timeFrom, timeTo time.Time) (int, int, bool, error) {
	date := fmt.Sprintf("%d-%02d-%02d", timeTo.Year(), timeTo.Month(), timeTo.Day())

	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s", n.Config.CollectorURL, "users", user.SlackUserID, date, date)
	logrus.Infof("rest: getCollectorData request URL: %s", linkURL)
	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		logrus.Errorf("notifier: Get Request failed: %v\n", err)
		return 0, 0, true, err
	}
	token := n.Config.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("notifier: Authorization failed: %v\n", err)
		return 0, 0, true, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("notifier: ioutil.ReadAll failed: %v\n", err)
		return 0, 0, true, err
	}
	var dataU reporting.UserData
	json.Unmarshal(body, &dataU)

	userIsNonReporter, err := n.DB.CheckNonReporter(user, timeFrom, timeTo)
	logrus.Printf("checkMotherFucker: worklogs %v", dataU.Worklogs/3600)
	logrus.Printf("checkMotherFucker: commits %v", dataU.TotalCommits)
	logrus.Printf("checkMotherFucker: isNonReporter %v", userIsNonReporter)

	return dataU.Worklogs / 3600, dataU.TotalCommits, userIsNonReporter, nil

}

func getMN() (MN, error) {
	noWorklogs, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "noWorklogs"})
	noCommits, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "noCommits"})
	noStandup, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "noStandup"})
	hasWorklogs, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "hasWorklogs"})
	hasCommits, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "hasCommits"})
	hasStandup, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "hasStandup"})
	isRook, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "isRook"})
	if err != nil {
		logrus.Errorf("slack: Localize failed: %v\n", err)
		return MN{}, err
	}
	mn := MN{
		noWorklogs:  noWorklogs,
		noCommits:   noCommits,
		noStandup:   noStandup,
		hasWorklogs: hasWorklogs,
		hasCommits:  hasCommits,
		hasStandup:  hasStandup,
		isRook:      isRook,
	}
	return mn, nil
}
