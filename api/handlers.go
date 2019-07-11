package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/model"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

var (
	missingTokenErr = "missing token / unauthorized"
	accessDeniedErr = "access denied"
)

//LoginPayload represents loginPayload from UI
type LoginPayload struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

//Validate LoginPayload
func (lp *LoginPayload) Validate() error {
	if strings.TrimSpace(lp.Code) == "" {
		return fmt.Errorf("code is required")
	}
	return nil
}

//ChangePasswordData is used to change password
type ChangePasswordData struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" validate:"min=8,max=40"`
}

func (api *ComedianAPI) healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "successful operation")
}

func (api *ComedianAPI) login(c echo.Context) error {
	ld := new(LoginPayload)
	if err := c.Bind(ld); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	if ld.Validate() != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": ld.Validate().Error()})
	}

	resp, err := slack.GetOAuthResponse(api.config.SlackClientID, api.config.SlackClientSecret, ld.Code, ld.RedirectURI, false)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	userIdentity := &slack.UserIdentityResponse{}

	userIdentityResponse, err := http.Get(fmt.Sprintf("https://slack.com/api/users.identity?token=%s", resp.AccessToken))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	err = json.NewDecoder(userIdentityResponse.Body).Decode(userIdentity)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	user, err := api.db.SelectUser(userIdentity.User.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "No user with such ID"})
	}

	bot, err := api.db.GetBotSettingsByTeamID(user.TeamID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "No comedian bot found"})
	}
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.UserID

	// Generate encoded token and send it as response.
	tokenString, err := token.SignedString([]byte(api.config.SlackClientSecret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user":  user,
		"bot":   bot,
	})
}

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

func (api *ComedianAPI) deleteBot(c echo.Context) error {

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

	err = api.db.DeleteBotSettingsByID(id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err}).Error("DeleteBotSettingsByID failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	settings, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	bot, err := api.comedian.SelectBot(settings.TeamName)
	if err != nil {
		log.WithFields(log.Fields{"bot": bot, "error": err}).Error("Could not select bot")
		return c.JSON(http.StatusOK, id)
	}

	bot.Stop()

	return c.JSON(http.StatusNoContent, id)
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

func (api *ComedianAPI) listUsers(c echo.Context) error {

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

	users, err := api.db.ListUsers()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("ListUsers failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	result := make([]model.User, 0)

	for _, user := range users {
		if user.TeamID == teamID {
			result = append(result, user)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func (api *ComedianAPI) getUser(c echo.Context) error {
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

	user, err := api.db.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if user.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	return c.JSON(http.StatusOK, user)
}

func (api *ComedianAPI) updateUser(c echo.Context) error {
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

	user, err := api.db.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if user.TeamID != teamID {
		return c.JSON(http.StatusForbidden, accessDeniedErr)
	}

	res, err := api.db.UpdateUser(user)
	if err != nil {
		log.WithFields(log.Fields{"user": user, "error": err}).Error("UpdateUser failed")
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
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

func (api *ComedianAPI) getStandupersOfChannel(c echo.Context) error {
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
		return c.JSON(404, "Channel not found")
	}
	standupers, err := api.db.ListChannelStandupers(channel.ChannelID)
	if err != nil {
		log.Errorf("Handler of: GET /channels/:id/standupers. ListStandupersByTeamID failed: %v", err)
		return c.JSON(500, "internal server error")
	}
	var result []model.Standuper
	for _, standuper := range standupers {
		if standuper.TeamID == teamID {
			result = append(result, standuper)
		}
	}
	return c.JSON(200, result)
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

func (api *ComedianAPI) logout(c echo.Context) error {
	return c.JSON(http.StatusCreated, "logged out")
}
