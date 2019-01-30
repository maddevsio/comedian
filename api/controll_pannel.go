package api

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

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

	sprintweekdays := strings.Split(ba.Bot.CP.SprintWeekdays, ",")
	weekdays := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	sprintdays := make(map[string]string)
	if len(sprintweekdays) == 7 {
		for i := 0; i < 7; i++ {
			if sprintweekdays[i] == "on" {
				sprintdays[weekdays[i]] = "checked"
			}
		}
	}
	//status of individual reporting
	var status string
	if ba.Bot.CP.IndividualReportingStatus {
		status = "checked"
	}

	//selected in <select> tag
	var languageSelectedRUS, languageSelectedEN string
	if ba.Bot.CP.Language == "ru_RU" {
		languageSelectedRUS = "selected"
	} else if ba.Bot.CP.Language == "en_US" {
		languageSelectedEN = "selected"
	}
	var sprintReportStatusSelectedTrue, sprintReportStatusSelectedFalse string
	if ba.Bot.CP.SprintReportStatus == true {
		sprintReportStatusSelectedTrue = "selected"
	} else {
		sprintReportStatusSelectedFalse = "selected"
	}
	var collectorSelectedEnabled, collectorSelectedDisabled string
	if ba.Bot.CP.CollectorEnabled == true {
		collectorSelectedEnabled = "selected"
	} else {
		collectorSelectedDisabled = "selected"
	}

	data := map[string]interface{}{
		"manager_slack_user_id":               ba.Bot.CP.ManagerSlackUserID,
		"reporting_channel":                   ba.Bot.CP.ReportingChannel,
		"individual_reporting_status":         status,
		"report_time":                         ba.Bot.CP.ReportTime,
		"notifier_interval":                   ba.Bot.CP.NotifierInterval,
		"reminder_time":                       ba.Bot.CP.ReminderTime,
		"reminder_repeats_max":                ba.Bot.CP.ReminderRepeatsMax,
		"language":                            ba.Bot.CP.Language,
		"languageSelectedRUS":                 languageSelectedRUS,
		"languageSelectedEN":                  languageSelectedEN,
		"collector_enabled":                   ba.Bot.CP.CollectorEnabled,
		"collector_selected_enabled":          collectorSelectedEnabled,
		"collector_selected_disabled":         collectorSelectedDisabled,
		"sprint_report_status":                ba.Bot.CP.SprintReportStatus,
		"sprint_report_status_selected_true":  sprintReportStatusSelectedTrue,
		"sprint_report_status_selected_false": sprintReportStatusSelectedFalse,
		"sprint_report_time":                  ba.Bot.CP.SprintReportTime,
		"sprint_report_channel":               ba.Bot.CP.SprintReportChannel,
		"monday":                              sprintdays["monday"],
		"tuesday":                             sprintdays["tuesday"],
		"wednesday":                           sprintdays["wednesday"],
		"thursday":                            sprintdays["thursday"],
		"friday":                              sprintdays["friday"],
		"saturday":                            sprintdays["saturday"],
		"sunday":                              sprintdays["sunday"],
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
	srs, err := strconv.ParseBool(form.Get("sprint_report_status"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	//calculate "individual_reporting_status"
	status := form.Get("individual_reporting_status")
	if status == "on" {
		cp.IndividualReportingStatus = true
	} else {
		cp.IndividualReportingStatus = false
	}
	cp.NotifierInterval = ni
	cp.ManagerSlackUserID = form.Get("manager_slack_user_id")
	cp.ReportingChannel = form.Get("reporting_channel")
	cp.ReportTime = form.Get("report_time")
	cp.Language = form.Get("language")
	cp.ReminderRepeatsMax = rrm
	cp.ReminderTime = rt
	cp.CollectorEnabled = ce
	cp.SprintReportStatus = srs
	cp.SprintReportTime = form.Get("sprint_report_time")
	cp.SprintReportChannel = form.Get("sprint_report_channel")

	monday := form.Get("monday")
	tuesday := form.Get("tuesday")
	wednesday := form.Get("wednesday")
	thursday := form.Get("thursday")
	friday := form.Get("friday")
	saturday := form.Get("saturday")
	sunday := form.Get("sunday")

	cp.SprintWeekdays = sunday + "," + monday + "," + tuesday + "," + wednesday + "," + thursday + "," + friday + "," + saturday

	_, err = ba.Bot.DB.UpdateControllPannel(*cp)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("UpdateControllPannel success")

	return ba.renderControllPannel(c)
}
