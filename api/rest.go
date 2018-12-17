package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/utils"
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/chat"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/reporting"
	"gitlab.com/team-monitoring/comedian/storage"
)

// REST struct used to handle slack requests (slash commands)
type REST struct {
	db     storage.Storage
	echo   *echo.Echo
	conf   config.Config
	report *reporting.Reporter
	slack  *chat.Slack
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

// NewRESTAPI creates API for Slack commands
func NewRESTAPI(slack *chat.Slack) (*REST, error) {
	e := echo.New()
	rep := reporting.NewReporter(slack)

	r := &REST{
		echo:   e,
		report: rep,
		db:     slack.DB,
		slack:  slack,
		conf:   slack.Conf,
	}

	endPoint := fmt.Sprintf("/commands%s", r.conf.SecretToken)
	r.echo.POST(endPoint, r.handleCommands)

	return r, nil
}

// Start starts http server
func (r *REST) Start() error {
	return r.echo.Start(r.conf.HTTPBindAddr)
}

func (r *REST) handleCommands(c echo.Context) error {
	var form FullSlackForm

	urlValues, err := c.FormParams()
	if err != nil {
		logrus.Errorf("rest: c.FormParams failed: %v\n", err)
		return c.String(http.StatusOK, err.Error())
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := decoder.Decode(&form, urlValues); err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	return c.String(http.StatusOK, r.implementCommands(form))

}

func (r *REST) implementCommands(form FullSlackForm) string {
	_, err := r.db.SelectChannel(form.ChannelID)
	if err != nil {
		logrus.Errorf("SelectChannel failed: %v", err)
		return err.Error()
	}

	if form.Command != "/comedian" {
		return err.Error()
	}

	accessLevel, err := r.getAccessLevel(form.UserID, form.ChannelID)
	if err != nil {
		return err.Error()
	}

	command, params := utils.CommandParsing(form.Text)

	switch command {
	case "add":
		return r.addCommand(accessLevel, form.ChannelID, params)
	case "show":
		return r.listCommand(form.ChannelID, params)
	case "remove":
		return r.deleteCommand(accessLevel, form.ChannelID, params)
	case "add_deadline":
		return r.addTime(accessLevel, form.ChannelID, params)
	case "remove_deadline":
		return r.removeTime(accessLevel, form.ChannelID)
	case "show_deadline":
		return r.showTime(form.ChannelID)
	case "add_timetable":
		return r.addTimeTable(accessLevel, form.ChannelID, params)
	case "remove_timetable":
		return r.removeTimeTable(accessLevel, form.ChannelID, params)
	case "show_timetable":
		return r.showTimeTable(accessLevel, form.ChannelID, params)
	case "report_on_user":
		return r.generateReportOnUser(accessLevel, params)
	case "report_on_project":
		return r.generateReportOnProject(accessLevel, params)
	case "report_on_user_in_project":
		return r.generateReportOnUserInProject(accessLevel, params)
	default:
		return r.displayHelpText("")
	}
}

func (r *REST) getAccessLevel(userID, channelID string) (int, error) {
	user, err := r.db.SelectUser(userID)
	if err != nil {
		return 0, err
	}
	if userID == r.slack.CP.ManagerSlackUserID {
		return 1, nil
	}
	if user.IsAdmin() {
		return 2, nil
	}
	if r.db.UserIsPMForProject(userID, channelID) {
		return 3, nil
	}
	return 4, nil
}
