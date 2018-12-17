package api

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/chat"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (r *REST) renderControllPannel(c echo.Context) error {

	logrus.Info(r.slack.CP)

	data := map[string]interface{}{
		"manager_slack_user_id": r.slack.CP.ManagerSlackUserID,
		"reporting_channel":     r.slack.CP.ReportingChannel,
		"report_time":           r.slack.CP.ReportTime,
		"notifier_interval":     r.slack.CP.NotifierInterval,
		"reminder_time":         r.slack.CP.ReminderTime,
		"reminder_repeats_max":  r.slack.CP.ReminderRepeatsMax,
		"language":              r.slack.CP.Language,
		"collector_enabled":     r.slack.CP.CollectorEnabled,
	}
	return c.Render(http.StatusOK, "admin", data)
}

func (r *REST) updateConfig(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Errorf("rest: c.FormParams failed: %v\n", err)
		return err
	}

	logrus.Info(form)

	cp, err := r.db.GetControllPannel()
	if err != nil {
		logrus.Error(err)
		return err
	}

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

	_, err = r.db.UpdateControllPannel(cp)
	if err != nil {
		logrus.Error(err)
		return err
	}

	s, err := chat.NewSlack(r.conf)
	if err != nil {
		return err
	}
	r.slack = s

	logrus.Info("UpdateControllPannel success")

	return r.renderControllPannel(c)
}
