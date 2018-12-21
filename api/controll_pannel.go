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

	data := map[string]interface{}{
		"manager_slack_user_id": ba.Bot.CP.ManagerSlackUserID,
		"reporting_channel":     ba.Bot.CP.ReportingChannel,
		"report_time":           ba.Bot.CP.ReportTime,
		"notifier_interval":     ba.Bot.CP.NotifierInterval,
		"reminder_time":         ba.Bot.CP.ReminderTime,
		"reminder_repeats_max":  ba.Bot.CP.ReminderRepeatsMax,
		"language":              ba.Bot.CP.Language,
		"collector_enabled":     ba.Bot.CP.CollectorEnabled,
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
	cp.NotifierInterval = ni
	cp.ManagerSlackUserID = form.Get("manager_slack_user_id")
	cp.ReportingChannel = form.Get("reporting_channel")
	cp.ReportTime = form.Get("report_time")
	cp.Language = form.Get("language")
	cp.ReminderRepeatsMax = rrm
	cp.ReminderTime = rt
	cp.CollectorEnabled = ce

	_, err = ba.Bot.DB.UpdateControllPannel(*cp)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("UpdateControllPannel success")

	return ba.renderControllPannel(c)
}
