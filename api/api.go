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
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/comedianbot"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

// ComedianAPI struct used to handle slack requests (slash commands)
type ComedianAPI struct {
	echo     *echo.Echo
	DB       *storage.MySQL
	Comedian *comedianbot.Comedian
}

// NewComedianAPI creates API for Slack commands
func NewComedianAPI(comedian *comedianbot.Comedian) (ComedianAPI, error) {

	echo := echo.New()

	api := ComedianAPI{
		echo:     echo,
		DB:       comedian.DB,
		Comedian: comedian,
	}

	t := &Template{
		templates: template.Must(template.ParseGlob(os.Getenv("GOPATH") + "/src/gitlab.com/team-monitoring/comedian/controll_pannel/*.html")),
	}

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
	return api.echo.Start(api.Comedian.Config.HTTPBindAddr)
}

func (api *ComedianAPI) handleEvent(c echo.Context) error {
	type Event struct {
		Token     string `json:"token"`
		Challenge string `json:"challenge"`
		Type      string `json:"type"`
	}

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

		err = api.DB.DeleteControlPannel(event.TeamID)
		if err != nil {
			return err
		}
		log.Info("Controll Pannel was deleted")
		return c.String(http.StatusOK, "Success")
	}

	return c.String(http.StatusOK, "Success")
}

func (api *ComedianAPI) handleServiceMessage(c echo.Context) error {
	type ServiceEvent struct {
		TeamName    string             `json:"team_name"`
		AccessToken string             `json:"bot_access_token"`
		Channel     string             `json:"channel"`
		Message     string             `json:"message"`
		Attachments []slack.Attachment `json:"attachments"`
	}

	var incomingEvent ServiceEvent

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Error(err)
		return err
	}

	err = json.Unmarshal(body, &incomingEvent)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("Service request: ", incomingEvent)

	for _, bot := range api.Comedian.Bots {
		log.Info("Bot that can handle request: ", bot.Properties)
		if bot.Properties.TeamName == incomingEvent.TeamName {
			log.Info(bot.Properties.AccessToken != incomingEvent.AccessToken)
			if bot.Properties.AccessToken != incomingEvent.AccessToken {
				return c.String(http.StatusForbidden, "wrong access token")
			}
			bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)
		}
	}

	return nil

}

func (api *ComedianAPI) handleCommands(c echo.Context) error {
	var form model.FullSlackForm

	urlValues, err := c.FormParams()
	if err != nil {
		log.Errorf("ComedianAPI: c.FormParams failed: %v\n", err)
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

	var bot *botuser.Bot

	for _, b := range api.Comedian.Bots {
		if form.TeamID == b.Properties.TeamID {
			bot = b
		}
	}

	if bot == nil {
		log.Error("No bot found to inmplement the request!")
		return c.String(http.StatusOK, "No bot is willing to implement the request. Something went wrong!")
	}

	log.Info(bot)

	_, err = bot.DB.SelectChannel(form.ChannelID)
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

	resp, err := slack.GetOAuthResponse(api.Comedian.Config.SlackClientID, api.Comedian.Config.SlackClientSecret, code, "", false)
	if err != nil {
		log.Error(err)
		return err
	}

	cp, err := api.DB.CreateControlPannel(resp.Bot.BotAccessToken, resp.TeamID, resp.TeamName)

	if err != nil {
		log.Error(err)
		return err
	}

	bot := &botuser.Bot{}

	bot.API = slack.New(resp.Bot.BotAccessToken)
	bot.Properties = cp
	bot.DB = api.Comedian.DB

	api.Comedian.BotsChan <- bot

	return api.renderLoginPage(c)
}
