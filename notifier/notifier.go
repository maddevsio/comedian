package notifier

import (
	"encoding/json"
	"errors"
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
	T             Translation
}

//Translation struct to get translation data
type Translation struct {
	noWorklogs          string
	noCommits           string
	noStandup           string
	hasWorklogs         string
	hasCommits          string
	hasStandup          string
	isRook              string
	notifyAllDone       string
	notifyNotAll        string
	notifyManagerNotAll string
	notifyUsersWarning  string
	notifyDirectMessage string
}

const (
	NotificationInterval   = 10
	ReminderRepeatsMax     = 5
	RemindManager          = 3
	warningTime            = 5
	MorningRooksReportTime = "15:25"
)

// NewNotifier creates a new notifier
func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		logrus.Errorf("notifier: NewMySQL failed: %v\n", err)
		return nil, err
	}
	r, err := time.Parse("15:04", c.ReportTime)
	if err != nil {
		logrus.Errorf("notifier: time.Parse failed: %v\n", err)
		return nil, err
	}

	translation, err := getTranslation()
	if err != nil {
		logrus.Errorf("notifier: getTranslation failed: %v\n", err)
		return nil, err
	}

	notifier := &Notifier{Chat: chat, DB: conn, Config: c, ReportTime: r, CheckInterval: c.NotifierCheckInterval, T: translation}
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

		if err != nil {
			logrus.Errorf("notifier: GetMN failed: %v\n", err)
		}
		worklogs, commits, isNonReporter, err := n.checkUser(user, timeFrom, currentTime)
		fails := ""
		if err != nil {
			logrus.Errorf("notifier: checkMotherFucker failed: %v\n", err)
		}
		if worklogs < 8 {
			fails += fmt.Sprintf(n.T.noWorklogs, worklogs) + ", "
		} else {
			fails += fmt.Sprintf(n.T.hasWorklogs, worklogs) + ", "
		}
		if commits == 0 {
			fails += n.T.noCommits
		} else {
			fails += fmt.Sprintf(n.T.hasCommits, commits) + ", "
		}
		if isNonReporter == true {
			fails += n.T.noStandup
		} else {
			fails += n.T.hasStandup
		}
		if (worklogs < 8) || (commits == 0) || (isNonReporter == true) {
			text += fmt.Sprintf(n.T.isRook, user.SlackUserID, fails)
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
	for _, st := range standupTimes {
		channelID := st.ChannelID
		standupTime := time.Unix(st.Time, 0)
		warningTime := time.Unix(st.Time-warningTime*60, 0)
		currTime := time.Now()
		if currTime.Hour() == warningTime.Hour() && currTime.Minute() == warningTime.Minute() {
			n.SendWarning(channelID)
		}

		if currTime.Hour() == standupTime.Hour() && currTime.Minute() == standupTime.Minute() {
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
	// if everyone wrote their standups display all done message!
	if len(nonReporters) == 0 {
		err := n.Chat.SendMessage(channelID, n.T.notifyAllDone)
		if err != nil {
			logrus.Errorf("notifier: SendMessage failed: %v\n", err)
		}
		return
	}

	// othervise DM non reporters
	n.DMNonReporters(nonReporters)

	repeats := 0

	notifyNotAll := func() error {
		nonReporters, _ := getNonReporters(n.DB, channelID)
		if len(nonReporters) == 0 {
			n.Chat.SendMessage(channelID, n.T.notifyAllDone)
			return nil
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.SlackUserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		n.Chat.SendMessage(channelID, fmt.Sprintf(n.T.notifyNotAll, strings.Join(nonReportersSlackIDs, ", ")))

		if repeats <= ReminderRepeatsMax && len(nonReporters) > 0 {
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		logrus.Info("Stop backoff")
		return nil
	}

	b := backoff.NewConstantBackOff(NotificationInterval * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

// SendWarning reminds users in chat about upcoming standups
func (n *Notifier) SendWarning(channelID string) {
	nonReporters, err := getNonReporters(n.DB, channelID)
	if err != nil {
		logrus.Errorf("notifier: getNonReporters failed: %v\n", err)
	}
	if len(nonReporters) == 0 {
		n.Chat.SendMessage(channelID, n.T.notifyAllDone)
		return
	}

	slackUserID := []string{}
	for _, user := range nonReporters {
		slackUserID = append(slackUserID, "<@"+user.SlackUserID+">")
	}
	err = n.Chat.SendMessage(channelID, fmt.Sprintf(n.T.notifyUsersWarning, strings.Join(slackUserID, ", "), warningTime))
	if err != nil {
		logrus.Errorf("notifier: n.Chat.SendMessage failed: %v\n", err)
	}

}

// DMNonReporters writes DM to users who did not write standups
func (n *Notifier) DMNonReporters(nonReporters []model.StandupUser) error {
	//send each non reporter direct message
	for _, nonReporter := range nonReporters {
		logrus.Infof("notifier: Notifier Send Message to non reporter: %v", nonReporter)
		n.Chat.SendUserMessage(nonReporter.SlackUserID, fmt.Sprintf(n.T.notifyDirectMessage, nonReporter.SlackName, nonReporter.ChannelID))
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
	logrus.Printf("checkNonReporter: worklogs %v", dataU.Worklogs/3600)
	logrus.Printf("checkNonReporter: commits %v", dataU.TotalCommits)
	logrus.Printf("checkNonReporter: isNonReporter %v", userIsNonReporter)

	return dataU.Worklogs / 3600, dataU.TotalCommits, userIsNonReporter, nil

}

func getTranslation() (Translation, error) {
	localizer, err := config.GetLocalizer()
	if err != nil {
		logrus.Errorf("slack: GetLocalizer failed: %v\n", err)
		return Translation{}, err
	}
	m := make(map[string]string)
	r := []string{
		"noWorklogs", "noCommits", "noStandup", "hasWorklogs",
		"hasCommits", "hasStandup", "isRook", "notifyAllDone",
		"notifyNotAll", "notifyManagerNotAll", "notifyUsersWarning",
		"notifyDirectMessage",
	}

	for _, t := range r {
		translated, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: t})
		if err != nil {
			logrus.Errorf("slack: Localize failed: %v\n", err)
			return Translation{}, err
		}
		m[t] = translated
	}

	t := Translation{
		noWorklogs:          m["noWorklogs"],
		noCommits:           m["noCommits"],
		noStandup:           m["noStandup"],
		hasWorklogs:         m["hasWorklogs"],
		hasCommits:          m["hasCommits"],
		hasStandup:          m["hasStandup"],
		isRook:              m["isRook"],
		notifyAllDone:       m["notifyAllDone"],
		notifyNotAll:        m["notifyNotAll"],
		notifyManagerNotAll: m["notifyManagerNotAll"],
		notifyUsersWarning:  m["notifyUsersWarning"],
		notifyDirectMessage: m["notifyDirectMessage"],
	}
	return t, nil
}
