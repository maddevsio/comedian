package api

import (
	"fmt"
	"net/http"
	"net/url"

	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/reporting"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
)

// REST struct used to handle slack requests (slash commands)
type REST struct {
	db      storage.Storage
	e       *echo.Echo
	c       config.Config
	decoder *schema.Decoder
}

const (
	commandAddUser         = "/comedianadd"
	commandRemoveUser      = "/comedianremove"
	commandListUsers       = "/comedianlist"
	commandAddTime         = "/standuptimeset"
	commandRemoveTime      = "/standuptimeremove"
	commandListTime        = "/standuptime"
	commandReportByProject = "/comedian_report_by_project"
)

// NewRESTAPI creates API for Slack commands
func NewRESTAPI(c config.Config) (*REST, error) {
	e := echo.New()
	conn, err := storage.NewMySQL(c)
	if err != nil {
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
		default:
			return c.String(http.StatusNotImplemented, "Not implemented")
		}
	}
	return c.JSON(http.StatusMethodNotAllowed, "Command not allowed")
}

func (r *REST) addUserCommand(c echo.Context, f url.Values) error {
	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	cleanName := strings.TrimLeft(ca.Text, "@")
	_, err := r.db.CreateStandupUser(model.StandupUser{
		SlackName: cleanName,
		ChannelID: ca.ChannelID,
		Channel:   ca.ChannelName,
	})
	if err != nil {
		log.Errorf("could not create standup user: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v", err))
	}
	st, err := r.db.ListStandupTime(ca.ChannelID)
	if err != nil {
		log.Errorf("could not list standup time: %v", err)
	}
	if st.Time == int64(0) {
		return c.String(http.StatusOK, fmt.Sprintf("%s added, but there is no standup time for this channel", ca.Text))
	}
	return c.String(http.StatusOK, fmt.Sprintf("%s added", ca.Text))
}

func (r *REST) removeUserCommand(c echo.Context, f url.Values) error {
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	cleanName := strings.TrimLeft(ca.Text, "@")
	err := r.db.DeleteStandupUserByUsername(cleanName, ca.ChannelID)
	if err != nil {
		log.Errorf("could not delete standup user: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v", err))
	}
	return c.String(http.StatusOK, fmt.Sprintf("%s deleted", ca.Text))
}

func (r *REST) listUsersCommand(c echo.Context, f url.Values) error {
	log.Printf("%+v\n", f)
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	users, err := r.db.ListStandupUsers(ca.ChannelID)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v", err))
	}

	return c.JSON(http.StatusOK, &users)
}

func (r *REST) addTime(c echo.Context, f url.Values) error {

	var ca FullSlackForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	t, err := time.Parse("15:04", ca.Text)
	if err != nil {
		log.Error(err.Error())
		return c.String(http.StatusBadRequest, fmt.Sprintf("could not convert time: %v", err))
	}
	timeInt := t.Unix()

	_, err = r.db.CreateStandupTime(model.StandupTime{
		ChannelID: ca.ChannelID,
		Channel:   ca.ChannelName,
		Time:      timeInt,
	})
	if err != nil {
		log.Errorf("could not create standup time: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to add standup time :%v", err))
	}
	st, err := r.db.ListStandupUsers(ca.ChannelID)
	if err != nil {
		log.Errorf("could not list standup time: %v", err)
	}
	if len(st) == 0 {
		return c.String(http.StatusOK, fmt.Sprintf("standup time at %s (UTC) added, but there is no standup "+
			"users for this channel", ca.Text))
	}

	return c.String(http.StatusOK, fmt.Sprintf("standup time at %s (UTC) added",
		time.Unix(timeInt, 0).In(time.UTC).Format("15:04")))
}

func (r *REST) removeTime(c echo.Context, f url.Values) error {
	var ca ChannelForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err := r.db.DeleteStandupTime(ca.ChannelID)
	if err != nil {
		log.Errorf("could not delete standup time: %v", err)
		return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete standup time :%v", err))
	}
	st, err := r.db.ListStandupUsers(ca.ChannelID)
	if err != nil {
		log.Errorf("could not list standup time: %v", err)
	}
	if len(st) != 0 {
		return c.String(http.StatusOK, fmt.Sprintf("standup time for this channel removed, but there are "+
			"people marked as a standuper."))
	}
	return c.String(http.StatusOK, fmt.Sprintf("standup time for %s channel deleted", ca.ChannelName))
}

func (r *REST) listTime(c echo.Context, f url.Values) error {
	var ca ChannelIDForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	suTime, err := r.db.ListStandupTime(ca.ChannelID)
	if err != nil {
		log.Println(err)
		if err.Error() == "sql: no rows in result set" {
			return c.String(http.StatusOK, fmt.Sprintf("No standup time set for this channel yet! Please, add a standup time using `/standuptimeset` command!"))
		} else {
			return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list time :%v", err))
		}
	}
	return c.String(http.StatusOK, fmt.Sprintf("standup time at %s (UTC)",
		time.Unix(suTime.Time, 0).In(time.UTC).Format("15:04")))
}

func (r *REST) reportByProject(c echo.Context, f url.Values) error {
	var ca ChannelIDTextForm
	if err := r.decoder.Decode(&ca, f); err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	if err := ca.Validate(); err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	commandParams := strings.Fields(ca.Text)
	if len(commandParams) != 3 {
		return c.String(http.StatusOK, "Wrong number of arguments")
	}
	channelID := commandParams[0]
	dateFrom, err := time.Parse("2006-01-02", commandParams[1])
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	dateTo, err := time.Parse("2006-01-02", commandParams[2])
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	report, err := reporting.StandupReportByProject(r.db, channelID, dateFrom, dateTo)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	return c.String(http.StatusOK, report)
}
