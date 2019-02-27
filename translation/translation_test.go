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

	message, err := Translate(bundle, "en_US", "WrongProject", 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid project name!", message)

	message, err = Translate(bundle, "en_US", "Wrong", 0, nil)
	assert.Error(t, err)

	message, err = Translate(bundle, "ru_RU", "AddPMsAdded", 1, map[string]string{
		"PM":  "user1",
		"PMs": "user2, user3, user4",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user1 теперь ПМ канала.", message)

	message, err = Translate(bundle, "ru_RU", "AddPMsAdded", 2, map[string]string{
		"PM":  "user1",
		"PMs": "user2, user3, user4",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Следующие пользователи назначены ПМами: user2, user3, user4 .", message)
}
