package botuser

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/team-monitoring/comedian/translation"
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

	message, err := bot.generateHelpText(helpText[command])
	if err != nil {
		return bot.displayDefaultHelpText()
	}
	return message
}

func (bot *Bot) generateHelpText(messageID string) (string, error) {

	payload := translation.Payload{bot.bundle, bot.properties.Language, messageID, 0, nil}
	message, err := translation.Translate(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"TeamName":     bot.properties.TeamName,
			"Language":     payload.Lang,
			"MessageID":    payload.MessageID,
			"Count":        payload.Count,
			"TemplateData": payload.TemplateData,
		}).Error("Failed to translate help message!")
		return "", err
	}

	return message, nil
}

func (bot *Bot) displayDefaultHelpText() string {
	var message string
	helpOptions := []string{"AddMembers", "ShowMembers", "RemoveMembers", "AddDeadline", "ShowDeadline", "RemoveDeadline"}

	for _, msg := range helpOptions {
		text, err := bot.generateHelpText(msg)
		if err != nil {
			log.Error(err)
		}
		message += text + "\n"
	}

	return message
}
