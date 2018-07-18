package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/reporting"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

// REST struct used to handle slack requests (slash commands)
type REST struct {
	db      storage.Storage
	e       *echo.Echo
	c       config.Config
	decoder *schema.Decoder
}

const (
	commandAddUser                = "/comedianadd"
	commandRemoveUser             = "/comedianremove"
	commandListUsers              = "/comedianlist"
	commandAddTime                = "/standuptimeset"
	commandRemoveTime             = "/standuptimeremove"
	commandListTime               = "/standuptime"
	commandReportByProject        = "/report_by_project"
	commandReportByUser           = "/report_by_user"
	commandReportByProjectAndUser = "/report_by_project_and_user"
)

var localizer *i18n.Localizer

// NewRESTAPI creates API for Slack commands
func NewRESTAPI(c config.Config) (*REST, error) {
	e := echo.New()
	conn, err := storage.NewMySQL(c)
	if err != nil {
		logrus.Errorf("connect to DB: %v", err)
		return nil, err
	}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	r := &REST{
		db:      conn,
		e:       e,
		c:       c,
		decoder: decoder,
	}
	r.initEndpoints()

	localizer, err = config.GetLocalizer()
	if err != nil {
		logrus.Errorf("get localizer: %v", err)
		return nil, err
	}

	return r, nil
}

func (r *REST) initEndpoints() {
	r.e.POST("/commands", r.handleCommands)
}

// Start starts http server
func (r *REST) Start() error {
	return r.e.Start(r.c.HTTPBindAddr)
}

func (r *REST) handleCommands(c echo.Context) error {
	form, err := c.FormParams()
	if err != nil {
		logrus.Errorf("parse form params: %v", err)
		return c.JSON(http.StatusBadRequest, nil)
	}
	if command := form.Get("command"); command != "" {
		switch command {
		case commandAddUser:
			return r.addUserCommand(c, form)
		case commandRemoveUser:
			return r.removeUserCommand(c, form)
		case commandListUsers:
			return r.listUsersCommand(c, form)
		case commandAddTime:
			return r.addTime(c, form)
		case commandRemoveTime:
			return r.removeTime(c, form)
		case commandListTime:
			return r.listTime(c, form)
		case commandReportByProject:
			return r.reportByProject(c, form)
		case commandReportByUser:
			return r.reportByUser(c, form)
		case commandReportByProjectAndUser:
			return r.reportByProjectAndUser(c, form)
		default:
			return c.String(http.StatusNotImplemented, "Not implemented")
		}
	}
	return c.JSON(http.StatusMethodNotAllowed, "Command not allowed")
}

func (r *REST) addUserCommand(c echo.Context, f url.Values) error {
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	result := strings.Split(ca.Text, "|")
	slackUserID := strings.Replace(result[0], "<@", "", -1)
	userName := strings.Replace(result[1], ">", "", -1)

	user, err := r.db.FindStandupUserInChannel(userName, ca.ChannelID)
	if err != nil {
		_, err = r.db.CreateStandupUser(model.StandupUser{
			SlackUserID: slackUserID,
			SlackName:   userName,
			ChannelID:   ca.ChannelID,
			Channel:     ca.ChannelName,
		})
		if err != nil {
			logrus.Errorf("create standup user: %v", err)
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v", err))
		}
	}
	if user.SlackName == userName && user.ChannelID == ca.ChannelID {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "userExist"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, fmt.Sprintf(text))
	}
	st, err := r.db.ListStandupTime(ca.ChannelID)
	if err != nil {
		logrus.Errorf("list standup time: %v", err)
	}
	if st.Time == int64(0) {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "addUserNoStandupTime"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, fmt.Sprintf(text, userName))
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "addUser"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, userName))
}

func (r *REST) removeUserCommand(c echo.Context, f url.Values) error {
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	userName := strings.Replace(ca.Text, "@", "", -1)
	err := r.db.DeleteStandupUserByUsername(userName, ca.ChannelID)
	if err != nil {
		logrus.Errorf("could not delete standup user: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v", err))
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "deleteUser"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, userName))
}

func (r *REST) listUsersCommand(c echo.Context, f url.Values) error {
	logrus.Printf("%+v\n", f)
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	users, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if err != nil {
		logrus.Errorf("list standup users by channel ID: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v", err))
	}

	var userNames []string
	for _, user := range users {
		userNames = append(userNames, "<@"+user.SlackName+">")
	}

	if len(userNames) < 1 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "listNoStandupers"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, text)
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "listStandupers"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, strings.Join(userNames, ", ")))
}

func (r *REST) addTime(c echo.Context, f url.Values) error {

	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	result := strings.Split(ca.Text, ":")
	hours, err := strconv.Atoi(result[0])
	if err != nil {
		logrus.Errorf("convert time: %v", err)
		return err
	}
	munites, err := strconv.Atoi(result[1])
	if err != nil {
		logrus.Errorf("convert time: %v", err)
		return err
	}
	currentTime := time.Now()
	timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), hours, munites, 0, 0, time.Local).Unix()

	standupTime, err := r.db.CreateStandupTime(model.StandupTime{
		ChannelID: ca.ChannelID,
		Channel:   ca.ChannelName,
		Time:      timeInt,
	})
	if err != nil {
		logrus.Errorf("create standup time: %v", err)
		return err
	}
	st, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if err != nil {
		logrus.Errorf("list standup users by channel ID: %v", err)
		return err
	}
	if len(st) == 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "addStandupTimeNoUsers"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, fmt.Sprintf(text, standupTime.Time))
	}

	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "addStandupTime"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, standupTime.Time))
}

func (r *REST) removeTime(c echo.Context, f url.Values) error {
	var ca ChannelForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	err := r.db.DeleteStandupTime(ca.ChannelID)
	if err != nil {
		logrus.Errorf("could not delete standup time: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete standup time :%v", err))
	}
	st, err := r.db.ListStandupUsersByChannelID(ca.ChannelID)
	if len(st) != 0 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "removeStandupTimeWithUsers"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, fmt.Sprintf(text))
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "removeStandupTime"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, ca.ChannelName))
}

func (r *REST) listTime(c echo.Context, f url.Values) error {
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	standupTime, err := r.db.ListStandupTime(ca.ChannelID)
	if err != nil {
		logrus.Errorf("list standup time: %v", err)
		if err.Error() == "sql: no rows in result set" {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "showNoStandupTime"})
			if err != nil {
				logrus.Errorf("localize text: %v", err)
			}
			return c.String(http.StatusOK, fmt.Sprintf(text))
		} else {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list time :%v", err))
		}
	}
	text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "showStandupTime"})
	if err != nil {
		logrus.Errorf("localize text: %v", err)
	}
	return c.String(http.StatusOK, fmt.Sprintf(text, standupTime.Time))
}

func (r *REST) reportByProject(c echo.Context, f url.Values) error {
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	commandParams := strings.Fields(ca.Text)
	logrus.Println(len(commandParams))
	if len(commandParams) != 3 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "wrongNArgs"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, text)
	}
	channelID := commandParams[0]
	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	report, err := reporting.StandupReportByProject(r.db, channelID, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("standup report by project: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}

func (r *REST) reportByUser(c echo.Context, f url.Values) error {
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 3 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "userExist"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, text)
	}
	userfull := commandParams[0]
	result := strings.Split(userfull, "|")
	userName := strings.Replace(result[1], ">", "", -1)
	user, err := r.db.FindStandupUser(userName)
	if err != nil {
		logrus.Errorf("find standup user: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	report, err := reporting.StandupReportByUser(r.db, user, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("standup report by user: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}

func (r *REST) reportByProjectAndUser(c echo.Context, f url.Values) error {
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		logrus.Errorf("decode URL values: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		logrus.Errorf("validate form: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 4 {
		text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "userExist"})
		if err != nil {
			logrus.Errorf("localize text: %v", err)
		}
		return c.String(http.StatusOK, text)
	}
	channelID := commandParams[0]
	logrus.Println("0" + channelID)
	userfull := commandParams[1]
	logrus.Println("1" + userfull)
	userName := strings.Replace(userfull, "@", "", -1)
	logrus.Println("1" + userName)
	user, err := r.db.FindStandupUser(userName)
	if err != nil {
		logrus.Errorf("find standup user: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateFrom, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[3])
	if err != nil {
		logrus.Errorf("parse time: %v", err)
		return c.String(http.StatusOK, err.Error())
	}
	report, err := reporting.StandupReportByProjectAndUser(r.db, channelID, user, dateFrom, dateTo)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			text, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: "reportByProjectAndUser"})
			if err != nil {
				logrus.Errorf("localize text: %v", err)
			}
			return c.String(http.StatusOK, fmt.Sprintf(text))
		}
	}
	return c.String(http.StatusOK, report)
}
