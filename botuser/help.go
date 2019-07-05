package botuser

import (
	"github.com/maddevsio/comedian/translation"
)

//DisplayHelpText displays help text
func (bot *Bot) DisplayHelpText(command string) string {
	helpText := map[string]string{
		"add":             "AddMembers",
		"show":            "ShowMembers",
		"remove":          "RemoveMembers",
		"add_deadline":    "AddDeadline",
		"show_deadline":   "ShowDeadline",
		"remove_deadline": "RemoveDeadline",
	}
	return bot.generateHelpText(helpText[command])
}

func (bot *Bot) generateHelpText(messageID string) string {
	payload := translation.Payload{bot.properties.TeamName, bot.bundle, bot.properties.Language, messageID, 0, nil}
	text := translation.Translate(payload)
	return text
}

func (bot *Bot) displayDefaultHelpText() string {
	var message string
	helpOptions := []string{"AddMembers", "ShowMembers", "RemoveMembers", "AddDeadline", "ShowDeadline", "RemoveDeadline"}

	for _, msg := range helpOptions {
		text := bot.generateHelpText(msg)
		message += text + "\n"
	}

	return message
}
