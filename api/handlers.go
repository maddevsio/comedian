package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
)

func (api *RESTAPI) healthcheck(c echo.Context) error {
	log.Info("Status healthy!")
	return c.JSON(http.StatusOK, "successful operation")
}

func (api *RESTAPI) listBots(c echo.Context) error {

	bots, err := api.db.GetAllBotSettings()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, bots)
}

func (api *RESTAPI) getBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, bot)
}

func (api *RESTAPI) updateBot(c echo.Context) error {
	bot := model.BotSettings{}

	if err := c.Bind(&bot); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateBotSettings(bot)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	err = api.db.DeleteBotSettingsByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, "deleted")
}

func (api *RESTAPI) listStandups(c echo.Context) error {

	standups, err := api.db.ListStandups()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, standups)
}

func (api *RESTAPI) getStandup(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, standup)
}

func (api *RESTAPI) updateStandup(c echo.Context) error {
	standup := model.Standup{}

	if err := c.Bind(&standup); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateStandup(standup)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "deleted")
}

func (api *RESTAPI) listUsers(c echo.Context) error {
	users, err := api.db.ListUsers()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, users)
}

func (api *RESTAPI) getUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user, err := api.db.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (api *RESTAPI) updateUser(c echo.Context) error {
	user := model.User{}

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) listChannels(c echo.Context) error {
	channels, err := api.db.ListChannels()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, channels)
}

func (api *RESTAPI) getChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, channel)
}

func (api *RESTAPI) updateChannel(c echo.Context) error {
	channel := model.Channel{}

	if err := c.Bind(&channel); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateChannel(channel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	err = api.db.DeleteChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "deleted")
}

func (api *RESTAPI) listStandupers(c echo.Context) error {
	standupers, err := api.db.ListStandupers()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, standupers)
}

func (api *RESTAPI) getStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, standuper)
}

func (api *RESTAPI) updateStanduper(c echo.Context) error {
	standuper := model.Standuper{}

	if err := c.Bind(standuper); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateStanduper(standuper)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "deleted")
}
