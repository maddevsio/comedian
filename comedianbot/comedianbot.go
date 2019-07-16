package comedianbot

import (
	"errors"

	"encoding/json"
	"github.com/maddevsio/comedian/botuser"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack/slackevents"
)

//Dry is needed to not send messages to slack for testing purposes
var Dry bool

// Comedian is the main struct of the project
type Comedian struct {
	bots     []*botuser.Bot
	DB       *storage.DB
	botsChan chan *botuser.Bot
	Bundle   *i18n.Bundle
}

//New makes new Comedian
func New(bundle *i18n.Bundle, db *storage.DB) *Comedian {
	comedian := Comedian{}
	comedian.bots = []*botuser.Bot{}
	comedian.botsChan = make(chan *botuser.Bot)
	comedian.DB = db
	comedian.Bundle = bundle
	return &comedian
}

//SelectBot returns bot by its team id or teamname if found
func (comedian *Comedian) SelectBot(team string) (*botuser.Bot, error) {
	var botuser *botuser.Bot

	for _, bot := range comedian.bots {
		if bot.Suits(team) {
			botuser = bot
		}
	}

	if botuser == nil {
		return botuser, errors.New("no bot found to implement the request")
	}

	return botuser, nil
}

//AddBot sends Bot to Comedian Channel where Bot can start its Work
func (comedian *Comedian) AddBot(bot *botuser.Bot) {
	comedian.botsChan <- bot
}

//SetBots populates Comedian with bots
func (comedian Comedian) SetBots() error {
	config, err := config.Get()
	if err != nil {
		return err
	}

	settings, err := comedian.DB.GetAllBotSettings()
	if err != nil {
		return err
	}

	for _, s := range settings {
		bot := botuser.New(config, comedian.Bundle, &s, comedian.DB)
		comedian.AddBot(bot)
	}
	return nil
}

// StartBots launches the bots
func (comedian *Comedian) StartBots() {
	for bot := range comedian.botsChan {
		comedian.bots = append(comedian.bots, bot)
		bot.Start()
	}
}

//HandleEvent sends message to Slack Workspace
func (comedian *Comedian) HandleEvent(incomingEvent model.ServiceEvent) error {
	bot, err := comedian.SelectBot(incomingEvent.TeamName)
	if err != nil {
		return err
	}

	if bot.Settings().AccessToken != incomingEvent.AccessToken {
		return errors.New("Wrong access token")
	}

	if Dry {
		return nil
	}

	return bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)
}

//HandleCallbackEvent choses bot to deal with event and then handles event
func (comedian *Comedian) HandleCallbackEvent(event slackevents.EventsAPICallbackEvent) error {
	bot, err := comedian.SelectBot(event.TeamID)
	if err != nil {
		return err
	}

	ev := map[string]interface{}{}
	data, err := event.InnerEvent.MarshalJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &ev); err != nil {
		return err
	}

	return bot.HandleCallBackEvent(ev["type"].(string), data)
}
