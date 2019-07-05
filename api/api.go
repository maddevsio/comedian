package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/comedian/botuser"
	"github.com/maddevsio/comedian/comedianbot"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/crypto"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/sethvargo/go-password/password"
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

type Event struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type TeamMember struct {
	standuper    model.Standuper
	teamWorklogs int
}

var echoRouteRegex = regexp.MustCompile(`(?P<start>.*):(?P<param>[^\/]*)(?P<end>.*)`)

// New creates API instance
func New(config *config.Config, db *storage.DB, comedian *comedianbot.Comedian) ComedianAPI {

	echo := echo.New()
	echo.Use(middleware.CORS())
	echo.Pre(middleware.RemoveTrailingSlash())
	echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method:${method}, uri:${uri}, status:${status}, error:${error}\n",
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
	echo.POST("/info-message", api.handleInfoMessage)
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
	r.GET("/users/:id", api.getUser)
	r.PATCH("/users/:id", api.updateUser)

	r.GET("/channels", api.listChannels)
	r.GET("/channels/:id", api.getChannel)
	r.PATCH("/channels/:id", api.updateChannel)
	r.DELETE("/channels/:id", api.deleteChannel)
	r.GET("/channels/:id/standupers", api.getStandupersOfChannel)

	r.GET("/standupers", api.listStandupers)
	r.GET("/standupers/:id", api.getStanduper)
	r.PATCH("/standupers/:id", api.updateStanduper)
	r.DELETE("/standupers/:id", api.deleteStanduper)

	r.GET("/bots", api.listBots)
	r.GET("/bots/:id", api.getBot)
	r.PATCH("/bots/:id", api.updateBot)
	r.POST("/bots/:id/update-password", api.changePassword)
	r.DELETE("/bots/:id", api.deleteBot)

	r.POST("/logout", api.logout)
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
		return c.String(http.StatusOK, incomingEvent.Challenge)
	}

	if incomingEvent.Type == slackevents.CallbackEvent {
		var event slackevents.EventsAPICallbackEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("HandleCallbackEvent failed to unmarshar callbackevent")
			return c.JSON(http.StatusBadRequest, err)
		}

		err = api.comedian.HandleCallbackEvent(event)
		if err != nil {
			log.WithFields(log.Fields{"event": event, "error": err}).Error("HandleCallbackEvent failed")
		}

		return c.String(http.StatusOK, "Success")
	}

	return c.String(http.StatusOK, "Success")
}

func (api *ComedianAPI) handleServiceMessage(c echo.Context) error {

	var incomingEvent model.ServiceEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on ReadAll")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on Unmarshal body")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = api.comedian.HandleEvent(incomingEvent)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on HandleEvent")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Message handled!")
}

func (api *ComedianAPI) handleInfoMessage(c echo.Context) error {

	var incomingEvent model.InfoEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on ReadAll")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on Unmarshal body")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = api.comedian.HandleInfoEvent(incomingEvent)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("handleServiceMessage failed on HandleEvent")
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "Info Message handled!")
}

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.String(http.StatusBadRequest, "wrong verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"slashCommand": slashCommand, "error": err})).Error("handleCommands failed on Select Bot")
		return err
	}

	accessLevel, err := bot.GetAccessLevel(slashCommand.UserID, slashCommand.ChannelID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"bot": bot, "slashCommand": slashCommand, "error": err})).Error("handleCommands failed on GetAccessLevel")
		return err
	}

	command, params := utils.CommandParsing(slashCommand.Text)

	message := bot.ImplementCommands(slashCommand.ChannelID, command, params, accessLevel)

	return c.String(http.StatusOK, message)
}

func (api *ComedianAPI) handleUsersCommands(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.String(http.StatusBadRequest, "Invalid verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"slashCommand": slashCommand, "error": err})).Error("handleCommands failed on Select Bot")
		return c.String(http.StatusOK, "No bot found to implement the request. Please, report this to Comedian development team")
	}

	today := time.Now()
	dateFrom := fmt.Sprintf("%d-%02d-%02d", today.Year(), today.Month(), 1)
	dateTo := fmt.Sprintf("%d-%02d-%02d", today.Year(), today.Month(), today.Day())
	dataOnUser, err := bot.GetCollectorData("users", slashCommand.UserID, dateFrom, dateTo)
	if err != nil {
		return c.String(http.StatusOK, "Failed to get data from Collector. Make sure you were added to Collector database and try again")
	}

	message := fmt.Sprintf("You have logged %v from the begining of the month", utils.SecondsToHuman(dataOnUser.Worklogs))

	return c.String(http.StatusOK, message)
}

func (api *ComedianAPI) showTeamWorklogs(c echo.Context) error {
	slashCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !slashCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.String(http.StatusBadRequest, "Invalid verification token")
	}

	bot, err := api.comedian.SelectBot(slashCommand.TeamID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"slashCommand": slashCommand, "error": err})).Error("handleCommands failed on Select Bot")
		return c.String(http.StatusOK, "No bot found to implement the request. Please, report this to Comedian development team")
	}

	standupers, err := api.db.ListChannelStandupers(slashCommand.ChannelID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"slashCommand": slashCommand, "error": err})).Error("handleCommands failed on ListChannelStandupers")
		return c.String(http.StatusOK, "Could not retreve standupers of the project")
	}

	if len(standupers) == 0 {
		return c.String(http.StatusOK, "No one standups in the project. No data")
	}

	channel := standupers[0].ChannelName

	dates := strings.Split(slashCommand.Text, "-")
	var from, to time.Time

	if len(dates) == 2 {
		from, err = utils.StringToTime(dates[0])
		if err != nil {
			return c.String(http.StatusOK, err.Error())
		}

		to, err = utils.StringToTime(dates[1])
		if err != nil {
			return c.String(http.StatusOK, err.Error())
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
	members := []TeamMember{}

	for _, standuper := range standupers {
		userInProject := fmt.Sprintf("%v/%v", standuper.UserID, standuper.ChannelName)
		dataOnUserInProject, err := bot.GetCollectorData("user-in-project", userInProject, dateFrom, dateTo)
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"standuper": standuper, "error": err})).Error("failed to get data on user in project")
			continue
		}
		members = append(members, TeamMember{
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

	return c.String(http.StatusOK, message)
}

func (api *ComedianAPI) auth(c echo.Context) error {

	urlValues, err := c.FormParams()
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"error": err})).Error("auth failed on c.FormParams()")
		return c.String(http.StatusUnauthorized, err.Error())
	}

	code := urlValues.Get("code")

	resp, err := slack.GetOAuthResponse(api.config.SlackClientID, api.config.SlackClientSecret, code, "", false)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"config": api.config, "urlValues": urlValues, "error": err})).Error("auth failed on GetOAuthResponse")
		return err
	}

	pass, err := password.Generate(26, 10, 0, false, false)
	if err != nil {
		return err
	}

	encriptedPass, err := crypto.Generate(pass)
	if err != nil {
		return err
	}

	botSettings, err := api.db.GetBotSettingsByTeamID(resp.TeamID)
	if err != nil {
		cp, err := api.db.CreateBotSettings(resp.Bot.BotAccessToken, encriptedPass, resp.Bot.BotUserID, resp.TeamID, resp.TeamName)
		if err != nil {
			log.WithFields(log.Fields(map[string]interface{}{"resp": resp, "error": err})).Error("auth failed on CreateBotSettings")
			return err
		}

		bot := botuser.New(api.config, api.comedian.Bundle, cp, api.comedian.DB)
		message := fmt.Sprintf("Thank you for adding me to your workspace! Login at %v with: \n username: `%v`\n password: `%v`", api.config.UIurl, resp.TeamName, pass)
		err = bot.SendUserMessage(resp.UserID, message)
		if err != nil {
			log.Error("SendUserMessage failed in Auth: ", err)
		}
		err = bot.SendUserMessage(resp.UserID, "If you never used Comedian before, check Comedian short intro video for more information! https://youtu.be/huvmtJCCtOE")
		if err != nil {
			log.Error("SendUserMessage failed in Auth: ", err)
		}
		api.comedian.AddBot(bot)

		return c.Redirect(http.StatusMovedPermanently, api.config.UIurl)
	}

	botSettings.AccessToken = resp.Bot.BotAccessToken
	botSettings.UserID = resp.Bot.BotUserID
	botSettings.Password = encriptedPass

	settings, err := api.db.UpdateBotSettings(botSettings)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"resp": resp, "error": err})).Error("auth failed on CreateBotSettings")
		return err
	}

	bot, err := api.comedian.SelectBot(resp.TeamID)
	if err != nil {
		log.Error(err)
		return err
	}
	bot.SetProperties(settings)

	message := fmt.Sprintf("Settings updated! New Login at %v with: \n username: `%v`\n password: `%v`", api.config.UIurl, resp.TeamName, pass)
	err = bot.SendUserMessage(resp.UserID, message)
	if err != nil {
		log.Error("SendUserMessage failed in Auth: ", err)
	}
	err = bot.SendUserMessage(resp.UserID, "If you never used Comedian before, check Comedian short intro video for more information! https://youtu.be/huvmtJCCtOE")
	if err != nil {
		log.Error("SendUserMessage failed in Auth: ", err)
	}

	return c.Redirect(http.StatusMovedPermanently, api.config.UIurl)
}

func sortTeamMembers(entries []TeamMember) []TeamMember {
	var members []TeamMember

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

func sweep(entries []TeamMember, prevPasses int) bool {
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
