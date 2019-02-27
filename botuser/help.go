package botuser

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

//DisplayHelpText displays help text
func (bot *Bot) DisplayHelpText(command string) string {

	helpText := map[string]string{
		"":                "",
		"add":             "AddMembers",
		"show":            "ShowMembers",
		"remove":          "RemoveMembers",
		"add_deadline":    "AddDeadline",
		"show_deadline":   "ShowDeadline",
		"remove_deadline": "RemoveDeadline",
	}

	message, err := bot.generateHelpText(helpText[command])
	if err != nil {
		logrus.Error(err)
		return displayDefaultHelpText()
	}
	return message
}

func (bot *Bot) generateHelpText(messageID string) (string, error) {
	localizer := i18n.NewLocalizer(bot.bundle, bot.Properties.Language)

	message, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return "", err
	}

	return message, nil
}

func displayDefaultHelpText() string {
	return "All help!"
}
