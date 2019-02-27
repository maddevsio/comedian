package translation

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func Translate(bundle *i18n.Bundle, lang, messageID string, count int, templateData map[string]string) (string, error) {

	localizer := i18n.NewLocalizer(bundle, lang)

	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	if count != 0 {
		config.PluralCount = count
	}

	if templateData != nil {
		config.TemplateData = templateData
	}

	text, err := localizer.Localize(config)

	if err != nil {
		return "", err
	}

	return text, nil
}
