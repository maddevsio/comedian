package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/comedian/botuser"
	"github.com/maddevsio/comedian/comedianbot"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	log "github.com/sirupsen/logrus"
)

// ComedianAPI struct used to handle slack requests (slash commands)
type ComedianAPI struct {
	echo     *echo.Echo
	comedian *comedianbot.Comedian
	db       *storage.DB
	config   *config.Config
}

type swagger struct {
	Swagger  string
	Info     map[string]interface{}
	Host     string
	BasePath string `yaml:"basePath"`
	Tags     []struct {
		Name        string
		Description string
	}
	Schemes     []string
	Paths       map[string]interface{}
	Definitions map[string]interface{}
}

//Event represents slack challenge event
type Event struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type teamMember struct {
	standuper    model.Standuper
	teamWorklogs int
}

var echoRouteRegex = regexp.MustCompile(`(?P<start>.*):(?P<param>[^\/]*)(?P<end>.*)`)

//New creates API instance
func New(config *config.Config, db *storage.DB, comedian *comedianbot.Comedian) ComedianAPI {

	echo := echo.New()
	echo.Use(middleware.CORS())
	echo.Pre(middleware.RemoveTrailingSlash())
	echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method:${method}, uri:${uri}, status:${status}\n",
	}))

	api := ComedianAPI{
		echo:     echo,
		comedian: comedian,
		db:       db,
		config:   config,
	}

	echo.GET("/healthcheck", api.healthcheck)
	echo.POST("/login", api.login)
	echo.POST("/event", api.handleEvent)
	echo.POST("/service-message", api.handleServiceMessage)
	echo.POST("/commands", api.handleCommands)
	echo.POST("/team-worklogs", api.showTeamWorklogs)
	echo.POST("/user-commands", api.handleUsersCommands)
	echo.GET("/auth", api.auth)

	r := echo.Group("/v1")
	r.Use(middleware.JWT([]byte(config.SlackClientSecret)))

	r.GET("/standups", api.listStandups)
	r.GET("/standups/:id", api.getStandup)
	r.PATCH("/standups/:id", api.updateStandup)
	r.DELETE("/standups/:id", api.deleteStandup)

	r.GET("/users", api.listUsers)

	r.GET("/channels", api.listChannels)
	r.GET("/channels/:id", api.getChannel)
	r.PATCH("/channels/:id", api.updateChannel)
	r.DELETE("/channels/:id", api.deleteChannel)

	r.GET("/standupers", api.listStandupers)
	r.GET("/standupers/:id", api.getStanduper)
	r.PATCH("/standupers/:id", api.updateStanduper)
	r.DELETE("/standupers/:id", api.deleteStanduper)

	r.GET("/bots", api.listBots)
	r.GET("/bots/:id", api.getBot)
	r.PATCH("/bots/:id", api.updateBot)
	r.DELETE("/bots/:id", api.deleteBot)

	return api
}

// Start starts http server
func (api *ComedianAPI) Start() error {
	err := api.comedian.SetBots()
	if err != nil {
		return err
	}

	return api.echo.Start(api.config.HTTPBindAddr)
}

func (api *ComedianAPI) handleEvent(c echo.Context) error {
	var incomingEvent Event

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if incomingEvent.Token != api.config.SlackVerificationToken {
		return c.JSON(http.StatusForbidden, "verification token does not match")
	}

	if incomingEvent.Type == slackevents.URLVerification {
		return c.JSON(http.StatusOK, incomingEvent.Challenge)
	}

	if incomingEvent.Type == slackevents.CallbackEvent {
		var event slackevents.EventsAPICallbackEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		err = api.comedian.HandleCallbackEvent(event)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
	}

	return c.JSON(http.StatusOK, "Success")
}

func (api *ComedianAPI) handleServiceMessage(c echo.Context) error {

	var incomingEvent model.ServiceEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = api.comedian.HandleEvent(incomingEvent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusOK, "Message handled!")
}

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "wrong verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		return err
	}

	message := bot.ImplementCommands(slashCommand)

	return c.JSON(http.StatusOK, message)
}

func (api *ComedianAPI) handleUsersCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "Invalid verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		return c.JSON(http.StatusOK, "No bot found to implement the request. Please, report this to Comedian development team")
	}

	today := time.Now()
	dateFrom := fmt.Sprintf("%d-%02d-%02d", today.Year(), today.Month(), 1)
	dateTo := fmt.Sprintf("%d-%02d-%02d", today.Year(), today.Month(), today.Day())
	dataOnUser, err := bot.GetCollectorData("users", slashCommand.UserID, dateFrom, dateTo)
	if err != nil {
		return c.JSON(http.StatusOK, "Failed to get data from Collector. Make sure you were added to Collector database and try again")
	}

	message := fmt.Sprintf("You have logged %v from the begining of the month", botuser.SecondsToHuman(dataOnUser.Worklogs))

	return c.JSON(http.StatusOK, message)
}

func (api *ComedianAPI) showTeamWorklogs(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "Invalid verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		return c.JSON(http.StatusOK, "No bot found to implement the request. Please, report this to Comedian development team")
	}

	standupers, err := api.db.ListChannelStandupers(slashCommand.ChannelID)
	if err != nil {
		return c.JSON(http.StatusOK, "Could not retreve standupers of the project")
	}

	if len(standupers) == 0 {
		return c.JSON(http.StatusOK, "No one standups in the project. No data")
	}

	channel := standupers[0].ChannelName

	dates := strings.Split(slashCommand.Text, "-")
	var from, to time.Time

	if len(dates) == 2 {

		from, err = dateparse.ParseIn(strings.TrimSpace(dates[0]), time.Local)
		if err != nil {
			return c.JSON(http.StatusOK, err)
		}

		to, err = dateparse.ParseIn(strings.TrimSpace(dates[1]), time.Local)
		if err != nil {
			return c.JSON(http.StatusOK, err)
		}
	} else {
		today := time.Now()
		from = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
		to = today
	}

	dateFrom := fmt.Sprintf("%d-%02d-%02d", from.Year(), from.Month(), from.Day())
	dateTo := fmt.Sprintf("%d-%02d-%02d", to.Year(), to.Month(), to.Day())

	var message string
	var total int

	message += fmt.Sprintf("Worklogs of %s, from %s to %s: \n", channel, dateFrom, dateTo)
	members := []teamMember{}

	for _, standuper := range standupers {
		userInProject := fmt.Sprintf("%v/%v", standuper.UserID, standuper.ChannelName)
		dataOnUserInProject, err := bot.GetCollectorData("user-in-project", userInProject, dateFrom, dateTo)
		if err != nil {

			continue
		}
		members = append(members, teamMember{
			standuper:    standuper,
			teamWorklogs: dataOnUserInProject.Worklogs,
		})
		total += dataOnUserInProject.Worklogs
	}

	members = sortTeamMembers(members)

	for _, member := range members {
		message += fmt.Sprintf("%s - %.2f \n", member.standuper.RealName, float32(member.teamWorklogs)/3600)
	}

	message += fmt.Sprintf("In total: %.2f", float32(total)/3600)

	return c.JSON(http.StatusOK, message)
}

func (api *ComedianAPI) auth(c echo.Context) error {

	urlValues, err := c.FormParams()
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	code := urlValues.Get("code")

	resp, err := slack.GetOAuthResponse(api.config.SlackClientID, api.config.SlackClientSecret, code, "", false)
	if err != nil {
		return err
	}

	botSettings, err := api.db.GetBotSettingsByTeamID(resp.TeamID)
	if err != nil {
		cp, err := api.db.CreateBotSettings(resp.Bot.BotAccessToken, resp.Bot.BotUserID, resp.TeamID, resp.TeamName)
		if err != nil {

			return err
		}

		bot := botuser.New(api.config, api.comedian.Bundle, cp, api.comedian.DB)
		err = bot.SendUserMessage(resp.UserID, "If you never used Comedian before, check Comedian short intro video for more information! https://youtu.be/huvmtJCCtOE")
		if err != nil {
			log.Error("SendUserMessage failed in Auth: ", err)
		}
		api.comedian.AddBot(bot)

		return c.Redirect(http.StatusMovedPermanently, api.config.UIurl)
	}

	botSettings.AccessToken = resp.Bot.BotAccessToken
	botSettings.UserID = resp.Bot.BotUserID

	settings, err := api.db.UpdateBotSettings(botSettings)
	if err != nil {
		return err
	}

	bot, err := api.comedian.SelectBot(resp.TeamID)
	if err != nil {
		log.Error(err)
		return err
	}
	bot.SetProperties(settings)
	err = bot.SendUserMessage(resp.UserID, "If you never used Comedian before, check Comedian short intro video for more information! https://youtu.be/huvmtJCCtOE")
	if err != nil {
		log.Error("SendUserMessage failed in Auth: ", err)
	}

	return c.Redirect(http.StatusMovedPermanently, api.config.UIurl)
}

func sortTeamMembers(entries []teamMember) []teamMember {
	var members []teamMember

	for i := 0; i < len(entries); i++ {
		if !sweep(entries, i) {
			break
		}
	}

	for _, item := range entries {
		members = append(members, item)
	}

	return members
}

func sweep(entries []teamMember, prevPasses int) bool {
	var N = len(entries)
	var didSwap = false
	var firstIndex = 0
	var secondIndex = 1

	for secondIndex < (N - prevPasses) {

		var firstItem = entries[firstIndex]
		var secondItem = entries[secondIndex]
		if entries[firstIndex].teamWorklogs < entries[secondIndex].teamWorklogs {
			entries[firstIndex] = secondItem
			entries[secondIndex] = firstItem
			didSwap = true
		}
		firstIndex++
		secondIndex++
	}

	return didSwap
}
