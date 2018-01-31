package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type REST struct {
	db storage.Storage
	e  *echo.Echo
	c  config.Config
}

const (
	commandAdd    = "/comedianadd"
	commandRemove = "/comedianremove"
	commandList   = "/comedianlist"
	commandAddTime = "/standuptimeset"
	commandRemoveTime = "/standuptimeremove"
)

func NewRESTAPI(c config.Config) (*REST, error) {
	e := echo.New()
	conn, err := storage.NewMySQL(c)
	if err != nil {
		return nil, err
	}
	r := &REST{
		db: conn,
		e:  e,
		c:  c,
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
		case commandAdd:
			username := form.Get("text")
			if username == "" {
				return c.String(http.StatusBadRequest, "username cannot be empty")
			}
			channelID := form.Get("channel_id")
			channel := form.Get("channel_name")
			if channelID == "" || channel == "" {
				return c.String(http.StatusBadRequest, "channel cannot be empty")
			}
			_, err := r.db.CreateStandupUser(model.StandupUser{
				SlackName: username,
				ChannelID: channelID,
				Channel:   channel,
			})
			if err != nil {
				log.Println(err)
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to create user :%v", err))
			}
			return c.String(http.StatusOK, fmt.Sprintf("%s added", username))
		case commandRemove:
			username := form.Get("text")
			if username == "" {
				return c.String(http.StatusBadRequest, "username cannot be empty")
			}
			channelID := form.Get("channel_id")
			if channelID == "" {
				return c.String(http.StatusBadRequest, "channel cannot be empty")
			}
			err := r.db.DeleteStandupUserByUsername(username, channelID)
			if err != nil {
				log.Println(err)
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete user :%v", err))
			}
			return c.String(http.StatusOK, fmt.Sprintf("%s deleted", username))
		case commandList:
			channelID := form.Get("channel_id")
			if channelID == "" {
				return c.String(http.StatusBadRequest, "channel cannot be empty")
			}
			users, err := r.db.ListStandupUsers(channelID)
			if err != nil {
				log.Println(err)
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to list users :%v", err))
			}

			return c.JSON(http.StatusOK, &users)
		case commandAddTime:
			timeString := form.Get("text")
			if timeString == "" {
				return c.String(http.StatusBadRequest, "standup time cannot be empty")
			}
			timeInt, err := strconv.Atoi(timeString)
			if err != nil {
				log.Println(err)
			}
			channelID := form.Get("channel_id")
			channel := form.Get("channel_name")
			if channelID == "" || channel == "" {
				return c.String(http.StatusBadRequest, "channel cannot be empty")
			}
			_, err = r.db.CreateStandupTime(model.StandupTime{
				ChannelID: channelID,
				Channel:   channel,
				Time: int64(timeInt),
			})
			if err != nil {
				log.Println(err)
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to add standup time :%v", err))
			}
			return c.String(http.StatusOK, fmt.Sprintf("standup time at %d added", timeInt))
		case commandRemoveTime:
			channelID := form.Get("channel_id")
			channel := form.Get("channel_name")
			if channelID == "" || channel == "" {
				return c.String(http.StatusBadRequest, "channel cannot be empty")
			}
			err := r.db.DeleteStandupTime(channelID)
			if err != nil {
				log.Println(err)
				return c.String(http.StatusBadRequest, fmt.Sprintf("failed to delete standup time :%v", err))
			}
			return c.String(http.StatusOK, fmt.Sprintf("standup time for %s channel deleted", channel))

		default:
			return c.String(http.StatusNotImplemented, "Not implemented")
		}
	}
	return c.JSON(http.StatusMethodNotAllowed, "Command not allowed")
}
