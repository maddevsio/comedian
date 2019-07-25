package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/comedian/botuser"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	log "github.com/sirupsen/logrus"
)

// ComedianAPI struct used to handle slack requests (slash commands)
type ComedianAPI struct {
	echo   *echo.Echo
	db     *storage.DB
	config *config.Config
	bundle *i18n.Bundle
	bots   []*botuser.Bot
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

//LoginPayload represents loginPayload from UI
type LoginPayload struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
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
var dbService *storage.DB

//New creates API instance
func New(config *config.Config, db *storage.DB, bundle *i18n.Bundle) *ComedianAPI {

	echo := echo.New()
	echo.Use(middleware.CORS())
	echo.Pre(middleware.RemoveTrailingSlash())
	echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method:${method}, uri:${uri}, status:${status}\n",
	}))

	dbService = db

	api := ComedianAPI{
		echo:   echo,
		db:     db,
		config: config,
		bots:   []*botuser.Bot{},
	}

	echo.GET("/healthcheck", api.healthcheck)
	echo.POST("/login", api.login)
	echo.POST("/event", api.handleEvent)
	echo.POST("/service-message", api.handleServiceMessage)
	echo.POST("/commands", api.handleCommands)
	echo.POST("/team-worklogs", api.showTeamWorklogs)
	echo.POST("/user-commands", api.handleUsersCommands)
	echo.GET("/auth", api.auth)

	g := echo.Group("/v1")
	g.Use(AuthPreRequest)

	g.GET("/bots/:id", api.getBot)
	g.PATCH("/bots/:id", api.updateBot)

	g.GET("/standups", api.listStandups)
	g.GET("/standups/:id", api.getStandup)
	g.PATCH("/standups/:id", api.updateStandup)
	g.DELETE("/standups/:id", api.deleteStandup)

	g.GET("/channels", api.listChannels)
	g.PATCH("/channels/:id", api.updateChannel)
	g.DELETE("/channels/:id", api.deleteChannel)

	g.GET("/standupers", api.listStandupers)
	g.PATCH("/standupers/:id", api.updateStanduper)
	g.DELETE("/standupers/:id", api.deleteStanduper)

	return &api
}

// AuthPreRequest is the middleware function.
func AuthPreRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		accessToken := c.Request().Header.Get(echo.HeaderAuthorization)
		if accessToken == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing or incorrect Bot Access Token")
		}

		bot, err := dbService.GetBotSettingsByBotAccessToken(accessToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing or incorrect Bot Access Token")
		}

		c.Set("teamID", bot.TeamID)

		return next(c)
	}
}

//SelectBot returns bot by its team id or teamname if found
func (api *ComedianAPI) SelectBot(team string) *botuser.Bot {
	var bot botuser.Bot

	for _, b := range api.bots {
		if b.Suits(team) {
			return b
		}
	}

	return &bot
}

func (api *ComedianAPI) removeBot(team string) {
	var index int
	for i, b := range api.bots {
		if b.Suits(team) {
			index = i
		}
	}

	api.bots = append(api.bots[:index], api.bots[index+1:]...)
}

// Start starts http server
func (api *ComedianAPI) Start() error {

	settings, err := api.db.GetAllBotSettings()
	if err != nil {
		return err
	}

	for _, bs := range settings {
		bot := botuser.New(api.config, api.bundle, &bs, api.db)
		log.Info("New bot to append: ", bot)
		api.bots = append(api.bots, bot)
		bot.Start()
	}

	log.Info("Bots after append: ", api.bots)

	return api.echo.Start(api.config.HTTPBindAddr)
}

func (api *ComedianAPI) healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "Comedian is healthy")
}

func (api *ComedianAPI) login(c echo.Context) error {
	ld := new(LoginPayload)
	if err := c.Bind(ld); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, incorrectDataFormat)
	}

	resp, err := slack.GetOAuthResponse(api.config.SlackClientID, api.config.SlackClientSecret, ld.Code, ld.RedirectURI, false)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	slackClient := slack.New(resp.AccessToken)

	userIdentity, err := slackClient.GetUserIdentity()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userInfo, err := slackClient.GetUserInfo(userIdentity.User.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bot, err := api.db.GetBotSettingsByTeamID(userIdentity.Team.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Comedian was not invited to your Slack. Please, add it and try again")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": userInfo,
		"bot":  bot,
	})
}

func (api *ComedianAPI) handleEvent(c echo.Context) error {
	var incomingEvent Event

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if incomingEvent.Token != api.config.SlackVerificationToken {
		return echo.NewHTTPError(http.StatusUnauthorized, "verification token does not match")
	}

	if incomingEvent.Type == slackevents.URLVerification {
		return c.JSON(http.StatusOK, incomingEvent.Challenge)
	}

	if incomingEvent.Type == slackevents.CallbackEvent {
		var event slackevents.EventsAPICallbackEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = api.HandleCallbackEvent(event)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return c.JSON(http.StatusOK, "Success")
}

func (api *ComedianAPI) handleServiceMessage(c echo.Context) error {

	var incomingEvent model.ServiceEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = api.HandleEvent(incomingEvent)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Message handled!")
}

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "wrong verification token")
	}

	bot := api.SelectBot(slashCommand.TeamID)

	message := bot.ImplementCommands(slashCommand)

	return c.String(http.StatusOK, message)
}

func (api *ComedianAPI) handleUsersCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "Invalid verification token")
	}

	bot := api.SelectBot(slashCommand.TeamID)
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

//HandleEvent sends message to Slack Workspace
func (api *ComedianAPI) HandleEvent(incomingEvent model.ServiceEvent) error {
	bot := api.SelectBot(incomingEvent.TeamName)

	if bot.Settings().AccessToken != incomingEvent.AccessToken {
		return errors.New("Wrong access token")
	}

	return bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)
}

//HandleCallbackEvent choses bot to deal with event and then handles event
func (api *ComedianAPI) HandleCallbackEvent(event slackevents.EventsAPICallbackEvent) error {
	bot := api.SelectBot(event.TeamID)
	log.Info("Selected bot for HandleCallbackEvent: ", bot)

	ev := map[string]interface{}{}
	data, err := event.InnerEvent.MarshalJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &ev); err != nil {
		return err
	}

	switch ev["type"].(string) {
	case "message":
		log.Info("message!")
		message := &slack.MessageEvent{}
		if err := json.Unmarshal(data, message); err != nil {
			return err
		}
		return bot.HandleMessage(message)
	case "app_mention":
		log.Info("app_mention!")
		return nil
	case "member_joined_channel":
		join := &slack.MemberJoinedChannelEvent{}
		if err := json.Unmarshal(data, join); err != nil {
			return err
		}
		_, err := bot.HandleJoin(join.Channel, join.Team)
		return err
	case "app_uninstalled":
		bot.Stop()
		api.removeBot(event.TeamID)
		return api.db.DeleteBotSettings(event.TeamID)
	default:
		log.WithFields(log.Fields{"event": string(data)}).Warning("unrecognized event!")
		return nil
	}
}

func (api *ComedianAPI) showTeamWorklogs(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.JSON(http.StatusBadRequest, "Invalid verification token")
	}

	bot := api.SelectBot(slashCommand.TeamID)

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
		log.Error("slack.GetOAuthResponse failed: ", err)
		return err
	}

	botSettings, err := api.db.GetBotSettingsByTeamID(resp.TeamID)
	if err != nil {
		log.Error("GetBotSettingsByTeamID failed: ", err)

		bs := model.BotSettings{
			NotifierInterval:    30,
			Language:            "en_US",
			ReminderRepeatsMax:  3,
			ReminderTime:        int64(10),
			AccessToken:         resp.Bot.BotAccessToken,
			UserID:              resp.Bot.BotUserID,
			TeamID:              resp.TeamID,
			TeamName:            resp.TeamName,
			ReportingChannel:    "",
			ReportingTime:       "9:00",
			IndividualReportsOn: false,
		}

		cp, err := api.db.CreateBotSettings(bs)
		if err != nil {
			log.Error("CreateBotSettings failed: ", err)
			return err
		}

		bot := botuser.New(api.config, api.bundle, &cp, api.db)
		log.Info("botuser.New resulted in bot: ", bot)
		api.bots = append(api.bots, bot)
		log.Info("api.bots contains: ", api.bots)
		bot.Start()

		log.Info("api.bots contains: ", api.bots)
		return c.Redirect(http.StatusMovedPermanently, "maddevs.io")
	}

	botSettings.AccessToken = resp.Bot.BotAccessToken
	botSettings.UserID = resp.Bot.BotUserID

	settings, err := api.db.UpdateBotSettings(botSettings)
	if err != nil {
		return err
	}

	log.Info("New bot: ", settings)

	bot := api.SelectBot(resp.TeamID)

	bot.SetProperties(&settings)

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
