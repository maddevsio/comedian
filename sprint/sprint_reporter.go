package sprint

import (
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
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
		logrus.Info("Sprint Report or collector turned off ")
		return
	}

	hour, minute, err := utils.FormatTime(r.bot.CP.SprintReportTime)
	if err != nil {
		logrus.Info("Sprint Report Time messed up")
		return
	}

	if !(time.Now().Hour() == hour && time.Now().Minute() == minute) {
		logrus.Info("Not a time for Sprint Report")
		return
	}

	sprintWeekdays := strings.Split(r.bot.CP.SprintWeekdays, ",")

	//need to think about weekdays in ints Sunday should == 0
	if sprintWeekdays[time.Now().Weekday()] != "on" {
		logrus.Info("Sprint Report is not ON today!")
		return
	}

	channels, err := r.bot.DB.GetAllChannels()
	if err != nil {
		logrus.Error("Sprint Report failed: %v", err)
		return
	}

	localizer := i18n.NewLocalizer(r.bot.Bundle, r.bot.CP.Language)

	noActiveSprints := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:          "NotActiveSprint",
			Description: "Displays message if project doesn't has active sprint",
			Other:       "Project has no active sprint yet",
		},
	})

	for _, channel := range channels {

		if channel.StandupTime == 0 {
			logrus.Info("skip the channel!")
			continue
		}

		sprintInfo, err := r.GetSprintInfo(channel.ChannelName)
		if err != nil {
			logrus.Error(err)
			r.bot.SendMessage(r.bot.CP.SprintReportChannel, noActiveSprints, nil)
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

		attachments, err := r.MakeMessage(activeSprint, totalWorklogs)
		if err != nil {
			logrus.Error(err)
			continue
		}

		r.bot.SendMessage(r.bot.CP.SprintReportChannel, "", attachments)
	}
}
