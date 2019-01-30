package sprint

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/utils"
)

//SprintReporter is used to send report about sprint progress
type SprintReporter struct {
	bot *bot.Bot
}

//NewSprintReporter creates a new string reporter
func NewSprintReporter(Bot *bot.Bot) SprintReporter {
	return SprintReporter{Bot}
}

// Start starts all sprint reports treads
func (r *SprintReporter) Start() {
	sprintReport := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-sprintReport:
			r.SendSprintReport()
		}
	}
}

//SendSprintReport send report about sprint
func (r *SprintReporter) SendSprintReport() {
	logrus.Info("Send sprint report begin")

	if !r.bot.CP.SprintReportStatus || !r.bot.CP.CollectorEnabled {
		return
	}

	hour, minute, err := utils.FormatTime(r.bot.CP.SprintReportTime)
	if err != nil {
		logrus.Errorf("Sprint Report Time messed up. Error: %v", err)
		return
	}

	if !(time.Now().Hour() == hour && time.Now().Minute() == minute) {
		return
	}

	sprintWeekdays := strings.Split(r.bot.CP.SprintWeekdays, ",")

	//need to think about weekdays in ints Sunday should == 0
	if sprintWeekdays[time.Now().Weekday()] != "on" {
		return
	}

	channels, err := r.bot.DB.GetAllChannels()
	if err != nil {
		logrus.Error("Sprint Report failed: %v", err)
		return
	}

	for _, channel := range channels {

		if channel.StandupTime == 0 {
			logrus.Infof("skip the channel %v. standup time is 0", channel.ChannelName)
			continue
		}

		sprintInfo, err := r.GetSprintInfo(channel.ChannelName)
		if err != nil {
			logrus.Error(err)
			continue
		}

		activeSprint, err := MakeActiveSprint(sprintInfo)
		if err != nil {
			logrus.Error(err)
			continue
		}

		totalWorklogs, err := r.CountTotalWorklogs(channel.ChannelID, activeSprint.StartDate)
		if err != nil {
			logrus.Error(err)
			continue
		}

		message, attachments, err := r.MakeMessage(activeSprint, totalWorklogs)
		if err != nil {
			logrus.Error(err)
			continue
		}

		r.bot.SendMessage(r.bot.CP.SprintReportChannel, message, attachments)
	}
}
