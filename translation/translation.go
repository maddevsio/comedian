package translation

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Payload struct {
	Bundle       *i18n.Bundle
	Lang         string
	MessageID    string
	Count        int
	TemplateData map[string]interface{}
}

func Translate(payload Payload) (string, error) {

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
		return "", err
	}

	return text, nil
}
