package botuser

import (
	log "github.com/sirupsen/logrus"
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
		return displayDefaultHelpText()
	}
	return message
}

func (bot *Bot) generateHelpText(messageID string) (string, error) {

	payload := translation.Payload{bot.bundle, bot.Properties.Language, messageID, 0, nil}
	message, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.Properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate help message!")
		return "", err
	}

	return message, nil
}

func displayDefaultHelpText() string {
	return "All help!"
}
