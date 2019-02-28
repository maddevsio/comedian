package translation

import (
	"log"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestTranslate(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	if err != nil {
		log.Fatal(err)
	}
	_, err = bundle.LoadMessageFile("../active.ru.toml")
	if err != nil {
		log.Fatal(err)
	}

	testCases := []struct {
		language     string
		messageID    string
		pluralCount  int
		templateData map[string]interface{}
		expected     string
	}{
		{"en_US", "WrongProject", 0, nil, "Invalid project name!"},
		{"en_US", "Wrong", 0, nil, ""},
		{"ru_RU", "AddPMsAdded", 1, map[string]interface{}{"PM": "user1", "PMs": "user2, user3, user4"}, "user1 теперь ПМ в проекте."},
		{"ru_RU", "AddPMsAdded", 2, map[string]interface{}{"PM": "user1", "PMs": "user2, user3, user4"}, "user2, user3, user4 теперь ПМы в проекте."},
	}

	for _, tt := range testCases {
		payload := Payload{bundle, tt.language, tt.messageID, tt.pluralCount, tt.templateData}
		message, err := Translate(payload)
		if err != nil {
			assert.Error(t, err)
		}
		assert.Equal(t, tt.expected, message)
	}

}
