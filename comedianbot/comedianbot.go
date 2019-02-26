package comedianbot

import (
	"errors"

	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

// Comedian is the main struct of the project
type Comedian struct {
	bots     []*botuser.Bot
	db       *storage.MySQL
	botsChan chan *botuser.Bot
}

//New makes new Comedian
func New(db *storage.MySQL) *Comedian {
	comedian := Comedian{}
	comedian.bots = []*botuser.Bot{}
	comedian.botsChan = make(chan *botuser.Bot)
	comedian.db = db
	return &comedian
}

//SetBots populates Comedian with bots
func (comedian Comedian) SetBots() error {
	controllPannels, err := comedian.db.GetControlPannels()
	if err != nil {
		return err
	}

	for _, cp := range controllPannels {
		comedian.AddBot(cp)
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

func (comedian *Comedian) HandleEvent(incomingEvent model.ServiceEvent) error {
	bot, err := comedian.SelectBot(incomingEvent.TeamName)
	if err != nil {
		return err
	}

	bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)

	return nil
}

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

func (comedian *Comedian) AddBot(cp model.ControlPannel) *botuser.Bot {
	bot := botuser.New(cp, comedian.db)
	comedian.botsChan <- bot
	return bot
}
