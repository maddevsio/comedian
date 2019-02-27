package botuser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

func TestDisplayHelpText(t *testing.T) {
	config, err := config.Get()
	assert.NoError(t, err)

	db, err := storage.NewMySQL(config)
	assert.NoError(t, err)

	testCases := []struct {
		language      string
		command       string
		outputMessage string
	}{
		{"en_US", "", "All help!"},
		{"en_US", "add", "To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, default is a developer role, if the role is not selected. "},
	}

	for _, tt := range testCases {
		cp := model.ControlPannel{
			Language: tt.language,
		}

		bot := New(cp, db)
		text := bot.DisplayHelpText(tt.command)
		assert.Equal(t, tt.outputMessage, text)
	}

}
