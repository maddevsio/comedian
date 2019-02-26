package api

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/botuser"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (api *ComedianAPI) renderLoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

func (api *ComedianAPI) renderControlPannel(c echo.Context) error {

	form, err := c.FormParams()
	if err != nil {
		return err
	}

	logrus.Info(form)

	cp, err := api.DB.GetControlPannel(form.Get("team_name"))
	if err != nil {
		return c.Render(http.StatusNotFound, "login", nil)
	}

	if form.Get("password") != cp.Password {
		return c.Render(http.StatusForbidden, "login", nil)
	}

	sprintweekdays := strings.Split(cp.SprintWeekdays, ",")
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
	if cp.IndividualReportingStatus {
		status = "checked"
	}

	//selected in <select> tag
	var languageSelectedRUS, languageSelectedEN string
	if cp.Language == "ru_RU" {
		languageSelectedRUS = "selected"
	} else if cp.Language == "en_US" {
		languageSelectedEN = "selected"
	}
	var sprintReportStatusSelectedTrue, sprintReportStatusSelectedFalse string
	if cp.SprintReportStatus == true {
		sprintReportStatusSelectedTrue = "selected"
	} else {
		sprintReportStatusSelectedFalse = "selected"
	}
	var collectorSelectedEnabled, collectorSelectedDisabled string
	if cp.CollectorEnabled == true {
		collectorSelectedEnabled = "selected"
	} else {
		collectorSelectedDisabled = "selected"
	}

	data := map[string]interface{}{
		"team_name":                           cp.TeamName,
		"password":                            cp.Password,
		"manager_slack_user_id":               cp.ManagerSlackUserID,
		"reporting_channel":                   cp.ReportingChannel,
		"individual_reporting_status":         status,
		"report_time":                         cp.ReportTime,
		"notifier_interval":                   cp.NotifierInterval,
		"reminder_time":                       cp.ReminderTime,
		"reminder_repeats_max":                cp.ReminderRepeatsMax,
		"language":                            cp.Language,
		"languageSelectedRUS":                 languageSelectedRUS,
		"languageSelectedEN":                  languageSelectedEN,
		"collector_enabled":                   cp.CollectorEnabled,
		"collector_selected_enabled":          collectorSelectedEnabled,
		"collector_selected_disabled":         collectorSelectedDisabled,
		"sprint_report_status":                cp.SprintReportStatus,
		"sprint_report_status_selected_true":  sprintReportStatusSelectedTrue,
		"sprint_report_status_selected_false": sprintReportStatusSelectedFalse,
		"sprint_report_time":                  cp.SprintReportTime,
		"sprint_report_channel":               cp.SprintReportChannel,
		"task_done_status":                    cp.TaskDoneStatus,
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

func (api *ComedianAPI) updateConfig(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Errorf("BotAPI: c.FormParams failed: %v\n", err)
		return err
	}

	bot := &botuser.Bot{}
	for _, b := range api.Comedian.Bots {
		if b.Properties.TeamName == form.Get("team_name") {
			bot = b
		}
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
	srs, err := strconv.ParseBool(form.Get("sprint_report_status"))
	if err != nil {
		logrus.Error(err)
		return err
	}
	//calculate "individual_reporting_status"
	status := form.Get("individual_reporting_status")
	if status == "on" {
		bot.Properties.IndividualReportingStatus = true
	} else {
		bot.Properties.IndividualReportingStatus = false
	}
	bot.Properties.NotifierInterval = ni
	bot.Properties.ManagerSlackUserID = form.Get("manager_slack_user_id")
	bot.Properties.ReportingChannel = form.Get("reporting_channel")
	bot.Properties.ReportTime = form.Get("report_time")
	bot.Properties.Language = form.Get("language")
	bot.Properties.ReminderRepeatsMax = rrm
	bot.Properties.ReminderTime = rt
	bot.Properties.CollectorEnabled = ce
	bot.Properties.SprintReportStatus = srs
	bot.Properties.SprintReportTime = form.Get("sprint_report_time")
	bot.Properties.SprintReportChannel = form.Get("sprint_report_channel")
	bot.Properties.Password = form.Get("password")
	bot.Properties.TaskDoneStatus = form.Get("task_done_status")

	monday := form.Get("monday")
	tuesday := form.Get("tuesday")
	wednesday := form.Get("wednesday")
	thursday := form.Get("thursday")
	friday := form.Get("friday")
	saturday := form.Get("saturday")
	sunday := form.Get("sunday")

	bot.Properties.SprintWeekdays = sunday + "," + monday + "," + tuesday + "," + wednesday + "," + thursday + "," + friday + "," + saturday

	_, err = api.DB.UpdateControlPannel(bot.Properties)
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Info("UpdateControlPannel success")

	return api.renderControlPannel(c)
}
