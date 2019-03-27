package api

import (
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"gitlab.com/team-monitoring/comedian/crypto"
	"gitlab.com/team-monitoring/comedian/model"
)

//LoginData structure is used for parsing login data that UI sends to back end
type LoginData struct {
	TeamName string `json:"teamname"`
	Password string `json:"password"`
}

func (api *RESTAPI) healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "successful operation")
}

func (api *RESTAPI) login(c echo.Context) error {
	var data LoginData

	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	settings, err := api.db.GetBotSettingsByTeamName(data.TeamName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = crypto.Compare(settings.Password, data.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["team_id"] = settings.TeamID
	claims["bot_id"] = settings.ID
	claims["expire"] = time.Now().Add(time.Hour * 72).Unix() // do we need it?

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("COMEDIAN_SLACK_CLIENT_SECRET")))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"bot":   settings,
		"token": t,
	})
}

func (api *RESTAPI) listBots(c echo.Context) error {

	bots, err := api.db.GetAllBotSettings()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, bots)
}

func (api *RESTAPI) getBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)

	if int64(botID) != id {
		return c.JSON(http.StatusForbidden, "")
	}

	bot, err := api.db.GetBotSettings(id)
	if err != nil {
		// not sure about this error. Need to check how it works
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, bot)
}

func (api *RESTAPI) updateBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)

	if int64(botID) != id {
		return c.JSON(http.StatusForbidden, "")
	}

	bot := model.BotSettings{ID: id}

	if err := c.Bind(&bot); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	res, err := api.db.UpdateBotSettings(bot)
	if err != nil {
		// maybe it would be better to find bot first...
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) deleteBot(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	botID := claims["bot_id"].(float64)
	if int64(botID) != id {
		return c.JSON(http.StatusForbidden, "")
	}

	err = api.db.DeleteBotSettingsByID(id)
	if err != nil {
		//need to check if deleting bot with wrong id causes errors
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listStandups(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	standups, err := api.db.ListStandups()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	return c.JSON(http.StatusOK, standup)
}

func (api *RESTAPI) updateStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standup := model.Standup{ID: id}
	if err := c.Bind(&standup); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
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
		return c.JSON(http.StatusBadRequest, err)
	}

	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standup.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listUsers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	users, err := api.db.ListUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	user, err := api.db.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if user.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	return c.JSON(http.StatusOK, user)
}

func (api *RESTAPI) updateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	user := model.User{ID: id}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if user.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	res, err := api.db.UpdateUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (api *RESTAPI) listChannels(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	channels, err := api.db.ListChannels()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	return c.JSON(http.StatusOK, channel)
}

func (api *RESTAPI) updateChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	channel := model.Channel{ID: id}
	if err := c.Bind(&channel); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
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
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	channel, err := api.db.GetChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if channel.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	err = api.db.DeleteChannel(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusNoContent, id)
}

func (api *RESTAPI) listStandupers(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	claims := user.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)

	standupers, err := api.db.ListStandupers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
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
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	return c.JSON(http.StatusOK, standuper)
}

func (api *RESTAPI) updateStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standuper := model.Standuper{ID: id}
	if err := c.Bind(&standuper); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
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
		return c.JSON(http.StatusBadRequest, err)
	}

	u := c.Get("user").(*jwt.Token)
	if u == nil {
		return c.JSON(http.StatusUnauthorized, "")
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	claims := u.Claims.(jwt.MapClaims)
	teamID := claims["team_id"].(string)
	if standuper.TeamID != teamID {
		return c.JSON(http.StatusForbidden, "")
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusNoContent, id)
}
