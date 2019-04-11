package translation

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

type Payload struct {
	TeamName     string
	Bundle       *i18n.Bundle
	Lang         string
	MessageID    string
	Count        int
	TemplateData map[string]interface{}
}

func Translate(payload Payload) string {

	localizer := i18n.NewLocalizer(payload.Bundle, payload.Lang)

	config := &i18n.LocalizeConfig{
		MessageID: payload.MessageID,
	}

	if payload.Count != 0 {
		config.PluralCount = payload.Count
	}

	if payload.TemplateData != nil {
		config.TemplateData = payload.TemplateData
	}

	text, err := localizer.Localize(config)

	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     payload.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate message!")
		return ""
	}

	return text
}
