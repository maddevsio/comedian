package comedianbot

import (
	"errors"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack/slackevents"
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

// Comedian is the main struct of the project
type Comedian struct {
	bots     []*botuser.Bot
	db       storage.Storage
	botsChan chan *botuser.Bot
	bundle   *i18n.Bundle
}

//New makes new Comedian
func New(bundle *i18n.Bundle, db *storage.DB) *Comedian {
	comedian := Comedian{}
	comedian.bots = []*botuser.Bot{}
	comedian.botsChan = make(chan *botuser.Bot)
	comedian.db = db
	comedian.bundle = bundle
	return &comedian
}

//SetBots populates Comedian with bots
func (comedian Comedian) SetBots() error {
	settings, err := comedian.db.GetAllBotSettings()
	if err != nil {
		return err
	}

	for _, s := range settings {
		comedian.AddBot(s)
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

	return bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)
}

//HandleCallbackEvent choses bot to deal with event and then handles event
func (comedian *Comedian) HandleCallbackEvent(event slackevents.EventsAPICallbackEvent) error {
	bot, err := comedian.SelectBot(event.TeamID)
	if err != nil {
		return err
	}

	return bot.HandleCallBackEvent(event.InnerEvent)
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
func (comedian *Comedian) AddBot(settings model.BotSettings) {
	bot := botuser.New(comedian.bundle, settings, comedian.db)
	comedian.botsChan <- bot
}
