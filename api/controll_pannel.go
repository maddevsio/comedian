package api

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (ba *BotAPI) renderControllPannel(c echo.Context) error {

	logrus.Info(ba.Bot.CP)
	logrus.Info("ba.Bot.CP.SprintReportChannel: ", ba.Bot.CP.SprintReportChannel)
	logrus.Info("ba.Bot.CP.SprintReportTime: ", ba.Bot.CP.SprintReportTime)
	logrus.Info("ba.Bot.CP.SprintReportTurn: ", ba.Bot.CP.SprintReportTurn)
	logrus.Info("ba.Bot.CP.Monday: ", ba.Bot.CP.Monday)
	logrus.Info("ba.Bot.CP.Tuesday: ", ba.Bot.CP.Tuesday)
	logrus.Info("ba.Bot.CP.Wednesday: ", ba.Bot.CP.Wednesday)
	logrus.Info("ba.Bot.CP.Thursday: ", ba.Bot.CP.Thursday)
	logrus.Info("ba.Bot.CP.Friday: ", ba.Bot.CP.Friday)
	logrus.Info("ba.Bot.CP.Saturday: ", ba.Bot.CP.Saturday)
	logrus.Info("ba.Bot.CP.Sunday: ", ba.Bot.CP.Sunday)

	data := map[string]interface{}{
		"manager_slack_user_id": ba.Bot.CP.ManagerSlackUserID,
		"reporting_channel":     ba.Bot.CP.ReportingChannel,
		"report_time":           ba.Bot.CP.ReportTime,
		"notifier_interval":     ba.Bot.CP.NotifierInterval,
		"reminder_time":         ba.Bot.CP.ReminderTime,
		"reminder_repeats_max":  ba.Bot.CP.ReminderRepeatsMax,
		"language":              ba.Bot.CP.Language,
		"collector_enabled":     ba.Bot.CP.CollectorEnabled,
		"sprint_report_turn":    ba.Bot.CP.SprintReportTurn,
		"sprint_report_time":    ba.Bot.CP.SprintReportTime,
		"sprint_report_channel": ba.Bot.CP.SprintReportChannel,
		"monday":                ba.Bot.CP.Monday,
		"tuesday":               ba.Bot.CP.Tuesday,
		"wednesday":             ba.Bot.CP.Wednesday,
		"thursday":              ba.Bot.CP.Thursday,
		"friday":                ba.Bot.CP.Friday,
		"saturday":              ba.Bot.CP.Saturday,
		"sunday":                ba.Bot.CP.Sunday,
	}
	return c.Render(http.StatusOK, "admin", data)
}

func (ba *BotAPI) updateConfig(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Errorf("BotAPI: c.FormParams failed: %v\n", err)
		return err
	}

	logrus.Info(form)

	cp := ba.Bot.CP

	ni, err := strconv.Atoi(form.Get("notifier_interval"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	rrm, err := strconv.Atoi(form.Get("reminder_repeats_max"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	rt, err := strconv.ParseInt(form.Get("reminder_time"), 10, 64)
	if err != nil {
		logrus.Error(err)
		return err
	}
	ce, err := strconv.ParseBool(form.Get("collector_enabled"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	srt, err := strconv.ParseBool(form.Get("sprint_report_turn"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	cp.NotifierInterval = ni
	cp.ManagerSlackUserID = form.Get("manager_slack_user_id")
	cp.ReportingChannel = form.Get("reporting_channel")
	cp.ReportTime = form.Get("report_time")
	cp.Language = form.Get("language")
	cp.ReminderRepeatsMax = rrm
	cp.ReminderTime = rt
	cp.CollectorEnabled = ce
	cp.SprintReportTurn = srt
	cp.SprintReportTime = form.Get("sprint_report_time")
	cp.SprintReportChannel = form.Get("sprint_report_channel")
	cp.Monday = form.Get("monday")
	cp.Tuesday = form.Get("tuesday")
	cp.Wednesday = form.Get("wednesday")
	cp.Thursday = form.Get("thursday")
	cp.Friday = form.Get("friday")
	cp.Saturday = form.Get("saturday")
	cp.Sunday = form.Get("sunday")

	_, err = ba.Bot.DB.UpdateControllPannel(*cp)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("UpdateControllPannel success")

	return ba.renderControllPannel(c)
}
