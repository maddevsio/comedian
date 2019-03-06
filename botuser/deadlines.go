package botuser

import (
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/olebedev/when/rules/ru"
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/translation"
)

func (bot *Bot) addTime(accessLevel int, channelID, params string) string {
	if accessLevel > pmAccess {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
		accessAtLeastPM, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return accessAtLeastPM
	}

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(ru.All...)

	payload := translation.Payload{bot.bundle, bot.properties.Language, "SomethingWentWrong", 0, nil}
	somethingWentWrong, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}

	r, err := w.Parse(params, time.Now())
	if err != nil {
		return somethingWentWrong
	}
	if r == nil {
		return somethingWentWrong
	}

	channel, err := bot.db.SelectChannel(channelID)
	if err != nil {
		return somethingWentWrong
	}

	channel.StandupTime = r.Time.Unix()

	_, err = bot.db.UpdateChannel(channel)
	if err != nil {
		return somethingWentWrong
	}

	channelMembers, err := bot.db.ListChannelMembers(channelID)
	if err != nil {
		log.Errorf("BotAPI: ListChannelMembers failed: %v\n", err)
	}

	if len(channelMembers) == 0 {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "AddStandupTimeNoUsers", 0, map[string]interface{}{"timeInt": r.Time.Unix()}}
		addStandupTimeNoUsers, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return addStandupTimeNoUsers
	}

	payload = translation.Payload{bot.bundle, bot.properties.Language, "AddStandupTime", 0, map[string]interface{}{"timeInt": r.Time.Unix()}}
	addStandupTime, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}
	return addStandupTime
}

func (bot *Bot) removeTime(accessLevel int, channelID string) string {
	if accessLevel > pmAccess {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "AccessAtLeastPM", 0, nil}
		accessAtLeastPM, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return accessAtLeastPM
	}

	channel, err := bot.db.SelectChannel(channelID)
	if err != nil {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "SomethingWentWrong", 0, nil}
		somethingWentWrong, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return somethingWentWrong
	}

	channel.StandupTime = int64(0)

	_, err = bot.db.UpdateChannel(channel)

	if err != nil {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "SomethingWentWrong", 0, nil}
		somethingWentWrong, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return somethingWentWrong

	}
	st, err := bot.db.ListChannelMembers(channelID)
	if len(st) != 0 {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "RemoveStandupTimeWithUsers", 0, nil}
		removeStandupTimeWithUsers, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return removeStandupTimeWithUsers
	}

	payload := translation.Payload{bot.bundle, bot.properties.Language, "RemoveStandupTime", 0, nil}
	removeStandupTime, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}
	return removeStandupTime
}

func (bot *Bot) showTime(channelID string) string {
	channel, err := bot.db.SelectChannel(channelID)
	if err != nil || channel.StandupTime == int64(0) {
		payload := translation.Payload{bot.bundle, bot.properties.Language, "ShowNoStandupTime", 0, nil}
		showNoStandupTime, err := translation.Translate(payload)
		if err != nil {
			log.WithFields(log.Fields{
				"TeamName":     bot.properties.TeamName,
				"Language":     payload.Lang,
				"MessageID":    payload.MessageID,
				"Count":        payload.Count,
				"TemplateData": payload.TemplateData,
			}).Error("Failed to translate message!")
		}
		return showNoStandupTime
	}

	payload := translation.Payload{bot.bundle, bot.properties.Language, "ShowStandupTime", 0, map[string]interface{}{"standuptime": channel.StandupTime}}
	showStandupTime, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
	}
	return showStandupTime
}
