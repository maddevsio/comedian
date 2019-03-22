package api

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
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

type RESTAPI struct {
	db storage.Storage
}

var echoRouteRegex = regexp.MustCompile(`(?P<start>.*):(?P<param>[^\/]*)(?P<end>.*)`)

// New creates API
func New(config *config.Config, db storage.Storage, comedian *comedianbot.Comedian) ComedianAPI {

	echo := echo.New()
	echo.Use(middleware.CORS())
	echo.Pre(middleware.RemoveTrailingSlash())

	api := ComedianAPI{
		echo:     echo,
		comedian: comedian,
		db:       db,
		config:   config,
	}

	t := &Template{
		templates: template.Must(template.ParseGlob(os.Getenv("GOPATH") + "/src/gitlab.com/team-monitoring/comedian/templates/*.html")),
	}

	restAPI := RESTAPI{api.db}

	echo.GET("/v1/healthcheck", restAPI.healthcheck)

	echo.GET("/v1/standups", restAPI.listStandups)
	echo.GET("/v1/standups/:id", restAPI.getStandup)
	echo.POST("/v1/standups/:id", restAPI.updateStandup)
	echo.DELETE("/v1/standups/:id", restAPI.deleteStandup)

	echo.GET("/v1/users", restAPI.listUsers)
	echo.GET("/v1/users/:id", restAPI.getUser)
	echo.POST("/v1/users/:id", restAPI.updateUser)

	echo.GET("/v1/channels", restAPI.listChannels)
	echo.GET("/v1/channels/:id", restAPI.getChannel)
	echo.POST("/v1/channels/:id", restAPI.updateChannel)
	echo.DELETE("/v1/channels/:id", restAPI.deleteChannel)

	echo.GET("/v1/standupers", restAPI.listStandupers)
	echo.GET("/v1/standupers/:id", restAPI.getStanduper)
	echo.POST("/v1/standupers/:id", restAPI.updateStanduper)
	echo.DELETE("/v1/standupers/:id", restAPI.deleteStanduper)

	echo.GET("/v1/bots", restAPI.listBots)
	echo.GET("/v1/bots/:id", restAPI.getBot)
	echo.POST("/v1/bots/:id", restAPI.updateBot)
	echo.DELETE("/v1/bots/:id", restAPI.deleteBot)

	echo.Renderer = t
	echo.GET("/login", api.renderLoginPage)
	echo.POST("/event", api.handleEvent)
	echo.GET("/admin", api.renderControlPannel)
	echo.POST("/config", api.updateConfig)
	echo.POST("/service-message", api.handleServiceMessage)
	echo.POST("/commands", api.handleCommands)
	echo.GET("/auth", api.auth)
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
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	//Need for enabling of Event Subscriptions.
	if incomingEvent.Type == slackevents.URLVerification {
		return c.String(http.StatusOK, incomingEvent.Challenge)
	}

	if incomingEvent.Type == slackevents.CallbackEvent {
		var event slackevents.EventsAPICallbackEvent
		err = json.Unmarshal(body, &event)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		go func(event slackevents.EventsAPICallbackEvent) {
			log.Info(event)
			err = api.comedian.HandleCallbackEvent(event)
			if err != nil {
				log.Error(err)
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
		return err
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return err
	}

	err = api.comedian.HandleEvent(incomingEvent)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "Message handled!")
}

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	var form model.FullSlackForm

	urlValues, err := c.FormParams()
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	if err := decoder.Decode(&form, urlValues); err != nil {
		return c.String(http.StatusOK, err.Error())
	}

	if form.Command != "/comedian" {
		return c.String(http.StatusBadRequest, "slash command should be `/comedian`")
	}

	bot, err := api.comedian.SelectBot(form.TeamID)
	if err != nil {
		return err
	}

	accessLevel, err := bot.GetAccessLevel(form.UserID, form.ChannelID)
	if err != nil {
		return err
	}

	command, params := utils.CommandParsing(form.Text)

	message := bot.ImplementCommands(form.ChannelID, command, params, accessLevel)

	return c.String(http.StatusOK, message)

}

func (api *ComedianAPI) auth(c echo.Context) error {

	urlValues, err := c.FormParams()
	if err != nil {
		log.Errorf("ComedianAPI: c.FormParams failed: %v\n", err)
		return c.String(http.StatusUnauthorized, err.Error())
	}

	code := urlValues.Get("code")

	resp, err := slack.GetOAuthResponse(api.config.SlackClientID, api.config.SlackClientSecret, code, "", false)
	if err != nil {
		log.Error(err)
		return err
	}

	cp, err := api.db.CreateBotSettings(resp.Bot.BotAccessToken, resp.Bot.BotUserID, resp.TeamID, resp.TeamName)
	if err != nil {
		log.Errorf("ComedianAPI: CreateBotSettings failed: %v\n", err)
		return err
	}

	api.comedian.AddBot(cp)

	return api.renderLoginPage(c)
}
