package comedianbot

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/botuser"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

// Comedian is the main struct of the project
type Comedian struct {
	Bots     []*botuser.Bot
	Config   config.Config
	DB       *storage.MySQL
	BotsChan chan *botuser.Bot
}

//New makes new Comedian
func New(c config.Config) (*Comedian, error) {
	comedian := Comedian{}
	comedian.Config = c
	comedian.Bots = []*botuser.Bot{}
	comedian.BotsChan = make(chan *botuser.Bot)

	db, err := storage.NewMySQL(c)
	if err != nil {
		return &comedian, err
	}

	comedian.DB = db

	return &comedian, nil

}

//SetBots populates Comedian with bots
func (comedian Comedian) SetBots() error {
	listOfProperties, err := comedian.DB.GetControlPannels()
	if err != nil {
		return err
	}

	for _, properties := range listOfProperties {
		bot := botuser.Bot{}

		bot.API = slack.New(properties.AccessToken)
		bot.Properties = properties
		bot.DB = comedian.DB

		comedian.BotsChan <- &bot

	}
	return nil
}

// StartBots launches the bots
func (comedian *Comedian) StartBots() {
	for bot := range comedian.BotsChan {
		bundle := &i18n.Bundle{DefaultLanguage: language.English}
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		bundle.LoadMessageFile("comedianbot/active.en.toml")
		bundle.LoadMessageFile("comedianbot/active.ru.toml")

		bot.Bundle = bundle

		comedian.Bots = append(comedian.Bots, bot)

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
