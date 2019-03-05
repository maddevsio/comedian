package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
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

	bots, err := api.db.GetBotSettingsByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, bots)
	}

	return c.JSON(http.StatusOK, bots)
}
