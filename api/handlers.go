package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var accessDeniedErr = "access denied"

func (api *ComedianAPI) getBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"bot": bot})
}

func (api *ComedianAPI) updateBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	botSettings, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	if err := c.Bind(&botSettings); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	res, err := api.db.UpdateBotSettings(botSettings)
	if err != nil {
		log.WithFields(log.Fields{"bot": botSettings, "error": err}).Error("UpdateBotSettings failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})

	}

	bot, err := api.comedian.SelectBot(botSettings.TeamName)
	if err != nil {
		log.WithFields(log.Fields{"bot": bot, "error": err}).Error("Could not select bot")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	bot.SetProperties(&res)

	return c.JSON(http.StatusOK, map[string]interface{}{"bot": bot})
}

func (api *ComedianAPI) listStandups(c echo.Context) error {

	standups, err := api.db.ListTeamStandups(c.Get("teamID").(string))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListStandups failed")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standups": standups})
}

func (api *ComedianAPI) getStandup(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": err.Error()})
	}

	if standup.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standup": standup})
}

func (api *ComedianAPI) updateStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": err.Error()})
	}

	if err := c.Bind(&standup); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	if standup.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	standup.Modified = time.Now().UTC()

	standup, err = api.db.UpdateStandup(standup)
	if err != nil {
		log.WithFields(log.Fields{"standup": standup, "error": err}).Error("UpdateStandup failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})

	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standup": standup})
}

func (api *ComedianAPI) deleteStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	if standup.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("DeleteStandup failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})

	}

	return c.JSON(http.StatusNoContent, "")
}

func (api *ComedianAPI) listChannels(c echo.Context) error {

	channels, err := api.db.ListTeamChannels(c.Get("teamID").(string))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListChannels failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"channels": channels})
}

func (api *ComedianAPI) updateChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if err := c.Bind(&channel); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	if channel.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	channel, err = api.db.UpdateChannel(channel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"channel": channel})
}

func (api *ComedianAPI) deleteChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("GetChannel failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})

	}

	if channel.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	err = api.db.DeleteChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, "")
}

func (api *ComedianAPI) listStandupers(c echo.Context) error {

	standupers, err := api.db.ListTeamStandupers(c.Get("teamID").(string))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListStandupers failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standupers": standupers})
}

func (api *ComedianAPI) updateStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": err.Error()})
	}

	if err := c.Bind(&standuper); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	if standuper.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	standuper, err = api.db.UpdateStanduper(standuper)
	if err != nil {
		log.WithFields(log.Fields{"standuper": standuper, "error": err}).Error("UpdateStanduper failed")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standuper": standuper})
}

func (api *ComedianAPI) deleteStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	if standuper.TeamID != c.Get("teamID") {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": accessDeniedErr})
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, "")
}
