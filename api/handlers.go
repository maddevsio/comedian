package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/model"
)

func (api *ComedianAPI) healthcheck(c echo.Context) error {
	log.Info("Status healthy!")
	return c.JSON(http.StatusOK, "successful operation")
}

func (api *ComedianAPI) getBots(c echo.Context) error {
	log.Info("Get Bots!")

	bots, err := api.db.GetAllBotSettings()
	if err != nil {
		return c.JSON(http.StatusNotFound, bots)
	}

	return c.JSON(http.StatusOK, bots)
}

func (api *ComedianAPI) getBotByID(c echo.Context) error {
	log.Info("Get Bot by ID: ", c.Param("id"))

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	bot, err := api.db.GetBotSettingsByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, bot)
	}

	return c.JSON(http.StatusOK, bot)
}

func (api *ComedianAPI) updateBotByID(c echo.Context) error {
	log.Info("Update Bot by ID: ", c.Param("id"))
	bot := &model.BotSettings{}

	if err := c.Bind(bot); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	res, err := api.db.UpdateBotSettings(*bot)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *ComedianAPI) deleteBot(c echo.Context) error {
	log.Info("Detele bot with ID: ", c.Param("id"))

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	err = api.db.DeleteBotByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, "deleted")
}
