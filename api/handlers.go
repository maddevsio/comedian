package api

import (
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/model"
	log "github.com/sirupsen/logrus"
)

var (
	missingTokenErr = "missing token / unauthorized"
	accessDeniedErr = "access denied"
)

func (api *ComedianAPI) listBots(c echo.Context) error {
	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	expire := claims["expire"].(float64)

	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	bots := make([]model.BotSettings, 0)
	bots, err := api.db.GetAllBotSettings()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("GetAllBotSettings failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, bots)
}

func (api *ComedianAPI) getBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	if int64(botID) != id {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, bot)
}

func (api *ComedianAPI) updateBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	if int64(botID) != id {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := c.Bind(&bot); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := api.db.UpdateBotSettings(bot)
	if err != nil {
		log.WithFields(log.Fields{"bot": bot, "error": err}).Error("UpdateBotSettings failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	settings, err := api.comedian.SelectBot(bot.TeamName)
	if err != nil {
		log.WithFields(log.Fields{"bot": bot, "error": err}).Error("Could not select bot")
		return c.JSON(http.StatusOK, res)
	}
	settings.SetProperties(res)

	return c.JSON(http.StatusOK, res)
}

func (api *ComedianAPI) listStandups(c echo.Context) error {

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standups, err := api.db.ListStandups()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListStandups failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	result := make([]model.Standup, 0)

	for _, standup := range standups {
		if standup.TeamID == teamID {
			result = append(result, standup)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func (api *ComedianAPI) getStandup(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	return c.JSON(http.StatusOK, standup)
}

func (api *ComedianAPI) updateStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if err := c.Bind(&standup); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	standup.Modified = time.Now().UTC()

	res, err := api.db.UpdateStandup(standup)
	if err != nil {
		log.WithFields(log.Fields{"standup": standup, "error": err}).Error("UpdateStandup failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (api *ComedianAPI) deleteStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("DeleteStandup failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *ComedianAPI) listChannels(c echo.Context) error {

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	channels, err := api.db.ListChannels()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListChannels failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	result := make([]model.Channel, 0)

	for _, channel := range channels {
		if channel.TeamID == teamID {
			result = append(result, channel)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func (api *ComedianAPI) getChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	return c.JSON(http.StatusOK, channel)
}

func (api *ComedianAPI) updateChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if err := c.Bind(&channel); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	res, err := api.db.UpdateChannel(channel)
	if err != nil {
		log.WithFields(log.Fields{"channel": channel, "error": err}).Error("UpdateChannel failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (api *ComedianAPI) deleteChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("GetChannel failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	err = api.db.DeleteChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *ComedianAPI) listStandupers(c echo.Context) error {

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	user := c.Get("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standupers, err := api.db.ListStandupers()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListStandupers failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	result := make([]model.Standuper, 0)

	for _, standuper := range standupers {
		if standuper.TeamID == teamID {
			result = append(result, standuper)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func (api *ComedianAPI) getStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	return c.JSON(http.StatusOK, standuper)
}

func (api *ComedianAPI) updateStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if err := c.Bind(&standuper); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	res, err := api.db.UpdateStanduper(standuper)
	if err != nil {
		log.WithFields(log.Fields{"standuper": standuper, "error": err}).Error("UpdateStanduper failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (api *ComedianAPI) deleteStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if c.Get("user") == nil {
		return c.JSON(http.StatusUnauthorized, missingTokenErr)
	}
	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	expire := claims["expire"].(float64)
	if time.Now().Unix() > int64(expire) {
		return c.JSON(http.StatusForbidden, "Token expired")
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("deleteStanduper failed at GetStanduper ")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("DeleteStanduper failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusNoContent, id)
}
