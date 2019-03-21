package api

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/crypto"
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
		return c.Render(http.StatusInternalServerError, "login", nil)
	}

	settings, err := api.db.GetBotSettingsByTeamName(form.Get("team_name"))
	if err != nil {
		return c.Render(http.StatusNotFound, "login", nil)
	}

	err = crypto.Compare(settings.Password, form.Get("password"))

	if err != nil {
		return c.Render(http.StatusForbidden, "login", nil)
	}

	//selected in <select> tag
	var languageSelectedRUS, languageSelectedEN string
	if settings.Language == "ru_RU" {
		languageSelectedRUS = "selected"
	} else {
		languageSelectedEN = "selected"
	}

	data := map[string]interface{}{
		"team_name":            settings.TeamName,
		"password":             settings.Password,
		"notifier_interval":    settings.NotifierInterval,
		"reminder_time":        settings.ReminderTime,
		"reminder_repeats_max": settings.ReminderRepeatsMax,
		"language":             settings.Language,
		"languageSelectedRUS":  languageSelectedRUS,
		"languageSelectedEN":   languageSelectedEN,
	}
	return c.Render(http.StatusOK, "admin", data)
}

func (api *ComedianAPI) updateConfig(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Error(err)
		return err
	}

	bot, err := api.comedian.SelectBot(form.Get("team_name"))
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

	settings := bot.Settings()

	settings.NotifierInterval = ni
	settings.Language = form.Get("language")
	settings.ReminderRepeatsMax = rrm
	settings.ReminderTime = rt
	settings.Password = form.Get("password")

	bot.SetProperties(settings)

	_, err = api.db.UpdateBotSettings(settings)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return api.renderControlPannel(c)
}
