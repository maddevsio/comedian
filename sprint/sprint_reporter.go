package sprint

import (
	"time"

	"gitlab.com/team-monitoring/comedian/bot"
)

//ReporterSprint is used to send report about sprint progress
type ReporterSprint struct {
	bot *bot.Bot
}

//NewReporterSprint creates a new string reporter
func NewReporterSprint(Bot *bot.Bot) ReporterSprint {
	reporterSprint := ReporterSprint{
		bot: Bot,
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

//bot.CP.Turn
//bot.CP.PeportTime
//bot.CP.Weekdays
//bot.CP.FinStatus

//SendSprintReport send report about sprint
func (r *ReporterSprint) SendSprintReport() {

}
