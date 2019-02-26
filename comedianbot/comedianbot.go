package comedianbot

import (
	"errors"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
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

		bot.UpdateUsersList()

		go func(bot *botuser.Bot) {
			rtm := bot.API.NewRTM()
			go rtm.ManageConnection()
			for msg := range rtm.IncomingEvents {
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:
					botUserID := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
					bot.HandleMessage(ev, botUserID)
				case *slack.ConnectedEvent:
					log.Info("Reconnected!")
				case *slack.MemberJoinedChannelEvent:
					bot.HandleJoin(ev.Channel, ev.Team)
				}
			}
		}(bot)

		go func(bot *botuser.Bot) {
			notificationForChannels := time.NewTicker(time.Second * 60).C
			for {
				select {
				case <-notificationForChannels:
					bot.NotifyChannels()
				}
			}
		}(bot)
	}
}

func (comedian *Comedian) HandleEvent(incomingEvent model.ServiceEvent) error {
	for _, bot := range comedian.bots {
		log.Info("Bot that can handle request: ", bot.Properties)
		if bot.Properties.TeamName == incomingEvent.TeamName {
			log.Info(bot.Properties.AccessToken != incomingEvent.AccessToken)
			if bot.Properties.AccessToken != incomingEvent.AccessToken {
				return errors.New("wrong access token")
			}
			bot.SendMessage(incomingEvent.Channel, incomingEvent.Message, incomingEvent.Attachments)
		}
	}
	return nil
}

func (comedian *Comedian) SelectBot(team string) (*botuser.Bot, error) {
	var bot *botuser.Bot

	for _, b := range comedian.bots {
		if team == b.Properties.TeamID || team == b.Properties.TeamName {
			bot = b
		}
	}

	if bot == nil {
		return bot, errors.New("no bot found to implement the request")
	}

	return bot, nil
}

func (comedian *Comedian) AddBot(cp model.ControlPannel) error {
	bot := &botuser.Bot{}

	bot.API = slack.New(cp.AccessToken)
	bot.Properties = cp
	bot.DB = comedian.db

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFile("comedianbot/active.en.toml")
	bundle.LoadMessageFile("comedianbot/active.ru.toml")

	bot.Bundle = bundle

	comedian.botsChan <- bot
	return nil
}
