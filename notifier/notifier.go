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

	"github.com/jasonlvhit/gocron"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/storage"
	"github.com/sirupsen/logrus"
)

// Notifier struct is used to notify users about upcoming or skipped standups
type Notifier struct {
	Chat   chat.Chat
	DB     storage.Storage
	Config config.Config
}

// NewNotifier creates a new notifier
func NewNotifier(c config.Config, chat chat.Chat) (*Notifier, error) {
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	notifier := &Notifier{Chat: chat, DB: conn, Config: c}
	return notifier, nil
}

// Start starts all notifier treads
func (n *Notifier) Start() error {
	gocron.Every(1).Day().At(n.Config.ReportTime).Do(n.RevealRooks)
	gocron.Every(60).Seconds().Do(n.NotifyChannels)
	channel := gocron.Start()
	for {
		report := <-channel
		logrus.Println(report)
	}
}

// RevealRooks displays data about rooks in channel general
func (n *Notifier) RevealRooks() {
	// check if today is not saturday or sunday. During these days no notificatoins!
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		logrus.Info("It is Weekend!!! Do not disturb!!!")
		return
	}
	timeFrom := time.Now().AddDate(0, 0, -1)
	// if today is monday, check 3 days of performance for user
	if int(time.Now().Weekday()) == 1 {
		timeFrom = time.Now().AddDate(0, 0, -3)
	}
	allUsers, err := n.DB.ListAllStandupUsers()
	if err != nil {
		logrus.Errorf("notifier: n.GetCurrentDayNonReporters failed: %v\n", err)
		return
	}
	text := ""
	for _, user := range allUsers {
		worklogs, commits, err := n.getCollectorData(user, timeFrom, time.Now())
		if err != nil {
			logrus.Errorf("notifier: getCollectorData failed: %v\n", err)
			return
		}
		isNonReporter, err := n.DB.IsNonReporter(user.SlackUserID, user.ChannelID, timeFrom, time.Now())
		if err != nil {
			logrus.Errorf("notifier: IsNonReporter failed: %v\n", err)
			return
		}

		if (worklogs < 8) || (commits == 0) || (isNonReporter == true) {
			fails := ""
			if worklogs < 8 {
				fails += fmt.Sprintf(n.Config.Translate.NoWorklogs, worklogs) + ", "
			} else {
				fails += fmt.Sprintf(n.Config.Translate.HasWorklogs, worklogs) + ", "
			}
			if commits == 0 {
				fails += n.Config.Translate.NoCommits
			} else {
				fails += fmt.Sprintf(n.Config.Translate.HasCommits, commits) + ", "
			}
			if isNonReporter == true {
				fails += n.Config.Translate.NoStandup
			} else {
				fails += n.Config.Translate.HasStandup
			}

			text += fmt.Sprintf(n.Config.Translate.IsRook, user.SlackUserID, user.ChannelID, fails)
		}
	}

	n.Chat.SendMessage(n.Config.ChanGeneral, text)

}

// NotifyChannels reminds users of channels about upcoming or missing standups
func (n *Notifier) NotifyChannels() {
	if int(time.Now().Weekday()) == 6 || int(time.Now().Weekday()) == 0 {
		logrus.Info("It is Weekend!!! No standups!!!")
		return
	}
	standupTimes, err := n.DB.ListAllStandupTime()
	if err != nil {
		logrus.Errorf("notifier: ListAllStandupTime failed: %v\n", err)
		return
	}
	// For each standup time, if standup time is now, start reminder
	for _, st := range standupTimes {
		standupTime := time.Unix(st.Time, 0)
		warningTime := time.Unix(st.Time-n.Config.ReminderTime*60, 0)
		if time.Now().Hour() == warningTime.Hour() && time.Now().Minute() == warningTime.Minute() {
			n.SendWarning(st.ChannelID)
		}
		if time.Now().Hour() == standupTime.Hour() && time.Now().Minute() == standupTime.Minute() {
			n.SendChannelNotification(st.ChannelID)
		}
	}
}

// SendWarning reminds users in chat about upcoming standups
func (n *Notifier) SendWarning(channelID string) {
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	if len(nonReporters) == 0 {
		return
	}

	nonReportersIDs := []string{}
	for _, user := range nonReporters {
		nonReportersIDs = append(nonReportersIDs, "<@"+user.SlackUserID+">")
	}
	err = n.Chat.SendMessage(channelID, fmt.Sprintf(n.Config.Translate.NotifyUsersWarning, strings.Join(nonReportersIDs, ", "), n.Config.ReminderTime))
	if err != nil {
		logrus.Errorf("notifier: n.Chat.SendMessage failed: %v\n", err)
		return
	}

}

//SendChannelNotification starts standup reminders and direct reminders to users
func (n *Notifier) SendChannelNotification(channelID string) {
	nonReporters, err := n.getCurrentDayNonReporters(channelID)
	if err != nil {
		logrus.Errorf("notifier: n.getCurrentDayNonReporters failed: %v\n", err)
		return
	}
	// if everyone wrote their standups display all done message!
	if len(nonReporters) == 0 {
		err := n.Chat.SendMessage(channelID, n.Config.Translate.NotifyAllDone)
		if err != nil {
			logrus.Errorf("notifier: SendMessage failed: %v\n", err)
		}
		return
	}

	// othervise Direct Message non reporters
	for _, nonReporter := range nonReporters {
		err := n.Chat.SendUserMessage(nonReporter.SlackUserID, fmt.Sprintf(n.Config.Translate.NotifyDirectMessage, nonReporter.SlackName, nonReporter.ChannelID))
		if err != nil {
			logrus.Errorf("notifier: SendMessage failed: %v\n", err)
		}
	}

	repeats := 0

	// think about his code and refactor more!
	notifyNotAll := func() error {
		nonReporters, _ := n.getCurrentDayNonReporters(channelID)
		if len(nonReporters) == 0 {
			n.Chat.SendMessage(channelID, n.Config.Translate.NotifyAllDone)
			return nil
		}

		nonReportersSlackIDs := []string{}
		for _, nonReporter := range nonReporters {
			nonReportersSlackIDs = append(nonReportersSlackIDs, fmt.Sprintf("<@%v>", nonReporter.SlackUserID))
		}
		logrus.Infof("notifier: Notifier non reporters: %v", nonReporters)

		n.Chat.SendMessage(channelID, fmt.Sprintf(n.Config.Translate.NotifyNotAll, strings.Join(nonReportersSlackIDs, ", ")))

		if repeats <= n.Config.ReminderRepeatsMax && len(nonReporters) > 0 {
			repeats++
			err := errors.New("Continue backoff")
			return err
		}
		logrus.Info("Stop backoff")
		return nil
	}

	b := backoff.NewConstantBackOff(time.Duration(n.Config.NotifierInterval) * time.Minute)
	err = backoff.Retry(notifyNotAll, b)
	if err != nil {
		logrus.Errorf("notifier: backoff.Retry failed: %v\n", err)
	}
}

// getNonReporters returns a list of standupers that did not write standups
func (n *Notifier) getCurrentDayNonReporters(channelID string) ([]model.StandupUser, error) {
	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	nonReporters, err := n.DB.GetNonReporters(channelID, timeFrom, time.Now())
	if err != nil && err != errors.New("no rows in result set") {
		logrus.Errorf("notifier: GetNonReporters failed: %v\n", err)
		return nil, err
	}
	return nonReporters, nil
}

func (n *Notifier) getCollectorData(user model.StandupUser, timeFrom, timeTo time.Time) (int, int, error) {
	date := fmt.Sprintf("%d-%02d-%02d", timeTo.Year(), timeTo.Month(), timeTo.Day())
	linkURL := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s", n.Config.CollectorURL, "users", user.SlackUserID, date, date)

	req, err := http.NewRequest("GET", linkURL, nil)
	if err != nil {
		logrus.Errorf("notifier: Get Request failed: %v\n", err)
		return 0, 0, err
	}
	token := n.Config.CollectorToken
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("notifier: Authorization failed: %v\n", err)
		return 0, 0, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("notifier: ioutil.ReadAll failed: %v\n", err)
		return 0, 0, err
	}
	var collectorData reporting.CollectorData
	json.Unmarshal(body, &collectorData)

	return collectorData.Worklogs / 3600, collectorData.TotalCommits, nil
}
