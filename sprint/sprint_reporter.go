package sprint

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/reporting"
	"gitlab.com/team-monitoring/comedian/utils"
)

//ReporterSprint is used to send report about sprint progress
type ReporterSprint struct {
	bot      *bot.Bot
	reporter reporting.Reporter
}

//NewReporterSprint creates a new string reporter
func NewReporterSprint(Bot *bot.Bot, r reporting.Reporter) ReporterSprint {
	reporterSprint := ReporterSprint{
		bot:      Bot,
		reporter: r,
	}
	return reporterSprint
}

// Start starts all sprint reports treads
func (r *ReporterSprint) Start() {
	sprintReport := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-sprintReport:
			r.SendSprintReport()
		}
	}
}

//SendSprintReport send report about sprint
func (r *ReporterSprint) SendSprintReport() {
	if r.bot.CP.SprintReportStatus {
		sprintWeekdays := strings.Split(r.bot.CP.SprintWeekdays, ",")
		weekdays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		var sprintdays []string
		if len(sprintWeekdays) == 7 {
			for i := 0; i < 7; i++ {
				if sprintWeekdays[i] == "on" {
					sprintdays = append(sprintdays, weekdays[i])
				}
			}
		}
		if bot.InList(time.Now().Weekday().String(), sprintdays) {
			hour, minute, err := utils.FormatTime(r.bot.CP.SprintReportTime)
			if err != nil {
				logrus.Errorf("sprint_reporter: Error parsing report time: %v", err)
				return
			}
			if time.Now().Hour() == hour && time.Now().Minute() == minute {
				channels, err := r.bot.DB.GetAllChannels()
				if err != nil {
					logrus.Errorf("sprint.GetAllChannels failed: %v", err)
					return
				}
				for _, channel := range channels {
					logrus.Infof("GetSprintData by channel: %v", channel.ChannelName)
					collectorInfo, err := GetSprintData(r.bot, channel.ChannelName)
					if err != nil {
						logrus.Errorf("sprint_reporter: GetSprintData failed: %v", err)
						continue
					}
					logrus.Info("collectorInfo: ", collectorInfo)
					activeSprint := MakeActiveSprint(collectorInfo)
					logrus.Info("activeSprint: ", activeSprint)
					message, attachments, err := MakeMessage(r.bot, activeSprint, channel.ChannelName, r.reporter)
					if err != nil {
						return
					}
					r.bot.SendMessage(r.bot.CP.SprintReportChannel, message, attachments)
				}
			}
		}
	}
}
