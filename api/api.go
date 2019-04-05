package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/crypto"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"gitlab.com/team-monitoring/comedian/utils"
)

// ComedianAPI struct used to handle slack requests (slash commands)
type ComedianAPI struct {
	echo     *echo.Echo
	comedian *comedianbot.Comedian
	db       storage.Storage
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

var echoRouteRegex = regexp.MustCompile(`(?P<start>.*):(?P<param>[^\/]*)(?P<end>.*)`)

// New creates API
func New(config *config.Config, db storage.Storage, comedian *comedianbot.Comedian) ComedianAPI {

	echo := echo.New()
	echo.Use(middleware.CORS())
	echo.Pre(middleware.RemoveTrailingSlash())
	echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
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
		log.WithFields(log.Fields{"error": err}).Error("HandleCallbackEvent failed to unmarshar incomingevent")
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

		go func(event slackevents.EventsAPICallbackEvent) {
			err = api.comedian.HandleCallbackEvent(event)
			if err != nil {
				log.WithFields(log.Fields{"event": event, "error": err}).Error("HandleCallbackEvent failed")
			}
		}(event)

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

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	slachCommand, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !slachCommand.ValidateToken(api.config.SlackVerificationToken) {
		return c.String(http.StatusBadRequest, "could not understand request")
	}

	if slachCommand.Command != "/comedian" {
		log.WithFields(log.Fields(map[string]interface{}{"slachCommand": slachCommand})).Warning("Command is not /comedian")
		return c.String(http.StatusBadRequest, "slash command should be `/comedian`")
	}

	bot, err := api.comedian.SelectBot(slachCommand.TeamID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"slachCommand": slachCommand, "error": err})).Error("handleCommands failed on Select Bot")
		return err
	}

	accessLevel, err := bot.GetAccessLevel(slachCommand.UserID, slachCommand.ChannelID)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"bot": bot, "slachCommand": slachCommand, "error": err})).Error("handleCommands failed on GetAccessLevel")
		return err
	}

	command, params := utils.CommandParsing(slachCommand.Text)

	message := bot.ImplementCommands(slachCommand.ChannelID, command, params, accessLevel)

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

	cp, err := api.db.CreateBotSettings(resp.Bot.BotAccessToken, encriptedPass, resp.Bot.BotUserID, resp.TeamID, resp.TeamName)
	if err != nil {
		log.WithFields(log.Fields(map[string]interface{}{"resp": resp, "error": err})).Error("auth failed on CreateBotSettings")
		return err
	}

	bot := botuser.New(api.comedian.Bundle, cp, api.comedian.DB)
	message := fmt.Sprintf("Thank you for adding me to your workspace! Login at %v with: \n username: `%v`\n password: `%v`", api.config.UIurl, resp.TeamName, pass)
	err = bot.SendUserMessage(resp.UserID, message)
	if err != nil {
		log.Error("SendUserMessage failed in Auth: ", err)
	}
	api.comedian.AddBot(bot)

	return c.Redirect(http.StatusMovedPermanently, api.config.UIurl)
}
