package botuser

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/translation"
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

	message, err := translation.Translate(bot.bundle, bot.Properties.Language, messageID, 0, nil)
	if err != nil {
		return "", err
	}

	return message, nil
}

func displayDefaultHelpText() string {
	return "All help!"
}
