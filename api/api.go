package api

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

// ComedianAPI struct used to handle slack requests (slash commands)
type ComedianAPI struct {
	echo     *echo.Echo
	comedian *comedianbot.Comedian
	db       storage.Storage
	config   *config.Config
}

type Event struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type RESTAPI struct {
	db storage.Storage
}

// NewComedianAPI creates API for Slack commands
func NewComedianAPI(config *config.Config, db storage.Storage, comedian *comedianbot.Comedian) (ComedianAPI, error) {

	echo := echo.New()

	api := ComedianAPI{
		echo:     echo,
		comedian: comedian,
		db:       db,
		config:   config,
	}

	t := &Template{
		templates: template.Must(template.ParseGlob(os.Getenv("GOPATH") + "/src/gitlab.com/team-monitoring/comedian/templates/*.html")),
	}

	v1 := echo.Group("/v1")

	restAPI := RESTAPI{api.db}

	v1.GET("/healthcheck", restAPI.healthcheck)

	v1.GET("/standups", restAPI.listStandups)
	v1.GET("/standups/:id", restAPI.getStandup)
	v1.POST("/standups/:id", restAPI.updateStandup)
	v1.DELETE("/standups/:id", restAPI.deleteStandup)

	v1.GET("/users", restAPI.listUsers)
	v1.GET("/users/:id", restAPI.getUser)
	v1.POST("/users/:id", restAPI.updateUser)

	v1.GET("/channels", restAPI.listChannels)
	v1.GET("/channels/:id", restAPI.getChannel)
	v1.POST("/channels/:id", restAPI.updateChannel)
	v1.DELETE("/channels/:id", restAPI.deleteChannel)

	v1.GET("/standupers", restAPI.listStandupers)
	v1.GET("/standupers/:id", restAPI.getStanduper)
	v1.POST("/standupers/:id", restAPI.updateStanduper)
	v1.DELETE("/standupers/:id", restAPI.deleteStanduper)

	v1.GET("/bots", restAPI.listBots)
	v1.GET("/bots/:id", restAPI.getBot)
	v1.POST("/bots/:id", restAPI.updateBot)
	v1.DELETE("/bots/:id", restAPI.deleteBot)

	echo.Renderer = t
	echo.GET("/login", api.renderLoginPage)
	echo.POST("/event", api.handleEvent)
	echo.GET("/admin", api.renderControlPannel)
	echo.POST("/config", api.updateConfig)
	echo.POST("/service-message", api.handleServiceMessage)

	echo.POST("/commands", api.handleCommands)
	echo.GET("/auth", api.auth)

	err := comedian.SetBots()

	return api, err
}

// Start starts http server
func (api *ComedianAPI) Start() error {
	return api.echo.Start(api.config.HTTPBindAddr)
}

func (api *ComedianAPI) handleEvent(c echo.Context) error {
	var incomingEvent Event
	var event slackevents.EventsAPICallbackEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		return err
	}

	//Need for enabling of Event Subscriptions.
	if incomingEvent.Type == slackevents.URLVerification {
		return c.String(http.StatusOK, incomingEvent.Challenge)
	}

	if incomingEvent.Type == slackevents.CallbackEvent {
		err = json.Unmarshal(body, &event)
		if err != nil {
			return err
		}

		err = api.db.DeleteBotSettings(event.TeamID)
		if err != nil {
			return err
		}

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

	return nil
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

	message := bot.ImplementCommands(form)

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

	cp, err := api.db.CreateBotSettings(resp.Bot.BotAccessToken, resp.TeamID, resp.TeamName)
	if err != nil {
		return err
	}

	api.comedian.AddBot(cp)

	return api.renderLoginPage(c)
}
