package translation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslate(t *testing.T) {
	message, err := Translate("en_US", "WrongProject", 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid project name!", message)

	message, err = Translate("en_US", "Wrong", 0, nil)
	assert.Error(t, err)

	message, err = Translate("ru_RU", "AddPMsAdded", 1, map[string]string{
		"PM":  "user1",
		"PMs": "user2, user3, user4",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user1 теперь ПМ канала.", message)

	message, err = Translate("ru_RU", "AddPMsAdded", 2, map[string]string{
		"PM":  "user1",
		"PMs": "user2, user3, user4",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Следующие пользователи назначены ПМами: user2, user3, user4 .", message)
}
