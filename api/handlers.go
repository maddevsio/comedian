package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

var (
	incorrectID         = "Incorrect value for 'id', must be integer"
	accessDenied        = "Entity belongs to a different team, access denied"
	doesNotExist        = "Entity does not yet exist"
	incorrectDataFormat = "Incorrect data format, double check request body"
	somethingWentWrong  = "Something went wrong"
)

func (api *ComedianAPI) getBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	bot, err := api.db.GetWorkspace(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if bot.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"bot": bot})
}

func (api *ComedianAPI) updateBot(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	settings, err := api.db.GetWorkspace(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if err := c.Bind(&settings); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectDataFormat)
	}

	if settings.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	res, err := api.db.UpdateWorkspace(settings)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bot, err := api.SelectBot(settings.WorkspaceName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	bot.SetProperties(&res)

	return c.JSON(http.StatusOK, map[string]interface{}{"bot": bot})
}

func (api *ComedianAPI) getStandup(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if standup.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standup": standup})
}

func (api *ComedianAPI) listStandups(c echo.Context) error {

	standups, err := api.db.ListTeamStandups(c.Get("teamID").(string))
	if err != nil {
		echo.NewHTTPError(http.StatusUnauthorized, somethingWentWrong)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standups": standups})
}

func (api *ComedianAPI) updateStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if err := c.Bind(&standup); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectDataFormat)
	}

	if standup.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	standup, err = api.db.UpdateStandup(standup)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standup": standup})
}

func (api *ComedianAPI) deleteStandup(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	standup, err := api.db.GetStandup(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if standup.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	err = api.db.DeleteStandup(id)
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, somethingWentWrong)
	}

	return c.JSON(http.StatusNoContent, "")
}

func (api *ComedianAPI) listChannels(c echo.Context) error {

	channels, err := api.db.ListWorkspaceProjects(c.Get("teamID").(string))
	if err != nil {
		echo.NewHTTPError(http.StatusInternalServerError, somethingWentWrong)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"channels": channels})
}

func (api *ComedianAPI) updateChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	channel, err := api.db.GetProject(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if err := c.Bind(&channel); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectDataFormat)
	}

	if channel.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	channel, err = api.db.UpdateProject(channel)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"channel": channel})
}

func (api *ComedianAPI) deleteChannel(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	channel, err := api.db.GetProject(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if channel.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	err = api.db.DeleteProject(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, somethingWentWrong)
	}

	return c.JSON(http.StatusNoContent, "")
}

func (api *ComedianAPI) listStandupers(c echo.Context) error {

	standupers, err := api.db.ListWorkspaceStandupers(c.Get("teamID").(string))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, somethingWentWrong)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standupers": standupers})
}

func (api *ComedianAPI) updateStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if err := c.Bind(&standuper); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectDataFormat)
	}

	if standuper.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	standuper, err = api.db.UpdateStanduper(standuper)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"standuper": standuper})
}

func (api *ComedianAPI) deleteStanduper(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 0, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectID)
	}

	standuper, err := api.db.GetStanduper(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, doesNotExist)
	}

	if standuper.WorkspaceID != c.Get("teamID") {
		return echo.NewHTTPError(http.StatusUnauthorized, accessDenied)
	}

	err = api.db.DeleteStanduper(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, somethingWentWrong)
	}

	return c.JSON(http.StatusNoContent, "")
}
