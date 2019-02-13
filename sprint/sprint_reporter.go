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
			logrus.Errorf("Failed to GetSprintInfo in channel [%v], reason [%v]. Skipping!", channel.ChannelName, err)
			continue
		}

		logrus.Infof("Channlel [%v], Sprint Info: [%v]", channel.ChannelName, sprintInfo)

		activeSprint, err := r.MakeActiveSprint(sprintInfo)
		if err != nil {
			logrus.Errorf("Failed to MakeActiveSprint in channel [%v], reason [%v]. Skipping!", channel.ChannelName, err)
			continue
		}
		logrus.Infof("Channel [%v], Active Sprint: [%v]", channel.ChannelName, activeSprint)

		totalWorklogs, err := r.CountTotalWorklogs(channel.ChannelID, activeSprint.StartDate)
		if err != nil {
			logrus.Errorf("Failed to CountTotalWorklogs in channel [%v], reason [%v]. Skipping!", channel.ChannelName, err)
			continue
		}
		logrus.Infof("Channel [%v], Total Worklogs: [%v]", channel.ChannelName, totalWorklogs)

		message, attachments, err := r.MakeMessage(activeSprint, totalWorklogs)
		if err != nil {
			logrus.Errorf("Failed to MakeMessage in channel [%v], reason [%v]. Skipping!", channel.ChannelName, err)
			continue
		}

		r.bot.SendMessage(r.bot.CP.SprintReportChannel, message, attachments)
	}
}
