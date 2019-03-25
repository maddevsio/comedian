package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/crypto"
	"gitlab.com/team-monitoring/comedian/model"
)

func (api *RESTAPI) healthcheck(c echo.Context) error {
	log.Info("Status healthy!")
	return c.JSON(http.StatusOK, "successful operation")
}

func (api *RESTAPI) login(c echo.Context) error {
	teamname := c.FormValue("teamname")
	password := c.FormValue("password")

	settings, err := api.db.GetBotSettingsByTeamName(teamname)
	if err != nil {
		return echo.ErrNotFound
	}

	err = crypto.Compare(settings.Password, password)
	if err != nil {
		return echo.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["team_id"] = settings.TeamID
	claims["bot_id"] = settings.ID
	claims["expire"] = time.Now().Add(time.Hour * 72).Unix() // do we need it?

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	s, err := json.Marshal(settings)
	if err != nil {
		log.Error("Marshal settings failed ", err)
	}

	log.Info(string(s))

	return c.JSON(http.StatusOK, map[string]string{
		"bot":   string(s),
		"token": t,
	})
}

func (api *RESTAPI) listBots(c echo.Context) error {

	bots, err := api.db.GetAllBotSettings()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	// temporary sequrity feature
	var fBots []model.BotSettings
	for _, bot := range bots {
		bot.AccessToken = ""
		fBots = append(fBots, bot)
	}

	log.Info(fBots)

	return c.JSON(http.StatusOK, fBots)
}

func (api *RESTAPI) getBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)

	if int64(botID) != id {
		return echo.ErrNotFound
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, bot)
}

func (api *RESTAPI) updateBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)

	if int64(botID) != id {
		return echo.ErrNotFound
	}

	bot := model.BotSettings{}

	if err := c.Bind(&bot); err != nil {
		log.Error(err)
		return c.JSON(http.StatusNotAcceptable, err)
	}
	bot.ID = id

	res, err := api.db.UpdateBotSettings(bot)
	if err != nil {
		log.Error(err)
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)

	if int64(botID) != id {
		return echo.ErrNotFound
	}

	err = api.db.DeleteBotSettingsByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listStandups(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	standups, err := api.db.ListStandups()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	var result []model.Standup

	for _, standup := range standups {
		if standup.TeamID == teamID {
			result = append(result, standup)
		}
	}

	return c.JSON(http.StatusOK, result)
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

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, standup)
}

func (api *RESTAPI) updateStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standup := model.Standup{}

	if err := c.Bind(&standup); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standup.ID = id

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return echo.ErrNotFound
	}

	res, err := api.db.UpdateStandup(standup)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return echo.ErrNotFound
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listUsers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	users, err := api.db.ListUsers()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	var result []model.User

	for _, user := range users {
		if user.TeamID == teamID {
			result = append(result, user)
		}
	}

	return c.JSON(http.StatusOK, result)
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

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if user.TeamID != teamID {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, user)
}

func (api *RESTAPI) updateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user := model.User{}

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	user.ID = id

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if user.TeamID != teamID {
		return echo.ErrNotFound
	}

	res, err := api.db.UpdateUser(user)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) listChannels(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	channels, err := api.db.ListChannels()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	var result []model.Channel

	for _, channel := range channels {
		if channel.TeamID == teamID {
			result = append(result, channel)
		}
	}

	return c.JSON(http.StatusOK, result)
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

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, channel)
}

func (api *RESTAPI) updateChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	channel := model.Channel{}

	if err := c.Bind(&channel); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	channel.ID = id

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return echo.ErrNotFound
	}

	res, err := api.db.UpdateChannel(channel)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return echo.ErrNotFound
	}

	err = api.db.DeleteChannel(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listStandupers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	standupers, err := api.db.ListStandupers()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	var result []model.Standuper

	for _, standuper := range standupers {
		if standuper.TeamID == teamID {
			result = append(result, standuper)
		}
	}

	return c.JSON(http.StatusOK, result)
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

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standuper.TeamID != teamID {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, standuper)
}

func (api *RESTAPI) updateStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standuper := model.Standuper{}

	if err := c.Bind(&standuper); err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standuper.ID = id

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	log.Info(teamID)
	log.Info(standuper.TeamID)
	if standuper.TeamID != teamID {
		return echo.ErrNotFound
	}

	res, err := api.db.UpdateStanduper(standuper)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusNotAcceptable, err)
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	u := c.Get("user").(*jwt.Token)
	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standuper.TeamID != teamID {
		return echo.ErrNotFound
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}

	return c.JSON(http.StatusNoContent, id)
}
