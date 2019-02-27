package translation

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func Translate(lang, messageID string, count int, templateData map[string]string) (string, error) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err := bundle.LoadMessageFile("translation/active.en.toml")
	if err != nil {
		return "", err
	}
	_, err = bundle.LoadMessageFile("translation/active.ru.toml")
	if err != nil {
		return "", err
	}

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
