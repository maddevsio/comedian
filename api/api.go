package api

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/utils"
)

// BotAPI struct used to handle slack requests (slash commands)
type BotAPI struct {
	echo *echo.Echo
	Bot  *bot.Bot
}

// FullSlackForm struct used for parsing full payload from slack
type FullSlackForm struct {
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ChannelID   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
	UserID      string `schema:"user_id"`
	UserName    string `schema:"user_name"`
}

// NewBotAPI creates API for Slack commands
func NewBotAPI(bot *bot.Bot) (*BotAPI, error) {

	e := echo.New()

	ba := &BotAPI{
		echo: e,
		Bot:  bot,
	}

	t := &Template{
		templates: template.Must(template.ParseGlob(os.Getenv("GOPATH") + "/src/gitlab.com/team-monitoring/comedian/controll_pannel/index.html")),
	}

	endPoint := fmt.Sprintf("/commands%s", ba.Bot.Conf.SecretToken)

	g := ba.echo.Group("/")

	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == ba.Bot.Conf.Login && password == ba.Bot.Conf.Password {
			return true, nil
		}
		return false, nil
	}))

	ba.echo.POST(endPoint, ba.handleCommands)
	ba.echo.Renderer = t
	g.GET("admin", ba.renderControllPannel)
	g.POST("config", ba.updateConfig)

	return ba, nil
}

// Start starts http server
func (ba *BotAPI) Start() error {
	return ba.echo.Start(ba.Bot.Conf.HTTPBindAddr)
}

func (ba *BotAPI) handleCommands(c echo.Context) error {
	var form FullSlackForm

	urlValues, err := c.FormParams()
	if err != nil {
		logrus.Errorf("BotAPI: c.FormParams failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := decoder.Decode(&form, urlValues); err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	return c.String(http.StatusOK, ba.implementCommands(form))

}

func (ba *BotAPI) implementCommands(form FullSlackForm) string {
	_, err := ba.Bot.DB.SelectChannel(form.ChannelID)
	if err != nil {
		logrus.Errorf("SelectChannel failed: %v", err)
		return err.Error()
	}

	if form.Command != "/comedian" {
		return err.Error()
	}

	accessLevel, err := ba.getAccessLevel(form.UserID, form.ChannelID)
	if err != nil {
		return err.Error()
	}

	command, params := utils.CommandParsing(form.Text)

	switch command {
	case "add":
		return ba.addCommand(accessLevel, form.ChannelID, params)
	case "show":
		return ba.showCommand(form.ChannelID, params)
	case "remove":
		return ba.deleteCommand(accessLevel, form.ChannelID, params)
	case "add_deadline":
		return ba.addTime(accessLevel, form.ChannelID, params)
	case "remove_deadline":
		return ba.removeTime(accessLevel, form.ChannelID)
	case "show_deadline":
		return ba.showTime(form.ChannelID)
	case "add_timetable":
		return ba.addTimeTable(accessLevel, form.ChannelID, params)
	case "remove_timetable":
		return ba.removeTimeTable(accessLevel, form.ChannelID, params)
	case "show_timetable":
		return ba.showTimeTable(accessLevel, form.ChannelID, params)
	case "report_on_user":
		return ba.generateReportOnUser(accessLevel, params)
	case "report_on_project":
		return ba.generateReportOnProject(accessLevel, params)
	case "report_on_user_in_project":
		return ba.generateReportOnUserInProject(accessLevel, params)
	case "add_onduty_project":
		return ba.addOnDutyProject(params, form.ChannelID)
	case "add_onduty_devops":
		return ba.addOnDutyDevops(params, form.ChannelID)
	case "onduty_show":
		return ba.onDutyShow(form.ChannelID)
	default:
		return ba.DisplayHelpText("")
	}
}

func (ba *BotAPI) getAccessLevel(userID, channelID string) (int, error) {
	user, err := ba.Bot.DB.SelectUser(userID)
	if err != nil {
		return 0, err
	}
	if userID == ba.Bot.CP.ManagerSlackUserID {
		return 1, nil
	}
	if user.IsAdmin() {
		return 2, nil
	}
	if ba.Bot.DB.UserIsPMForProject(userID, channelID) {
		return 3, nil
	}
	return 4, nil
}
