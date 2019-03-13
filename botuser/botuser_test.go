package botuser

import (
	"errors"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jarcoal/httpmock"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

type MockedDB struct {
	storage.Storage

	SelectedUser        model.User
	FoundStanduper      model.Standuper
	SelectedUserError   error
	FoundStanduperError error

	SelectedChannel      model.Channel
	CreatedChannel       model.Channel
	SelectedChannelError error
	CreatedChannelError  error
}

func (m MockedDB) SelectUser(string) (model.User, error) {
	return m.SelectedUser, m.SelectedUserError
}

func (m MockedDB) FindStansuperByUserID(string, string) (model.Standuper, error) {
	return m.FoundStanduper, m.FoundStanduperError
}

func (m MockedDB) SelectChannel(string) (model.Channel, error) {
	return m.SelectedChannel, m.SelectedChannelError
}

func (m MockedDB) CreateChannel(model.Channel) (model.Channel, error) {
	return m.CreatedChannel, m.CreatedChannelError
}

func TestNew(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(bundle, settings, db)
	assert.Equal(t, "TESTUSERID", bot.properties.UserID)

}

func TestAnalizeStandup(t *testing.T) {
	testCases := []struct {
		Message string
		Problem string
	}{
		{"", ":warning: No 'yesterday' related keywords detected! Please, use one of the following: 'yesterday' or weekdays such as 'friday' etc."},
		{"yesterday", ":warning: No 'today' related keywords detected! Please, use one of the following: 'today', 'going', 'plan'"},
		{"yesterday, today", ":warning: No 'problems' related keywords detected! Please, use one of the following: 'problem', 'difficult', 'stuck', 'question', 'issue'"},
		{"yesterday, today, problems", ""},
	}

	properties := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: properties,
	}

	for _, tt := range testCases {
		problem := bot.analizeStandup(tt.Message)
		assert.Equal(t, tt.Problem, problem)
	}

	testCasesErr := []struct {
		Message string
		Problem string
	}{
		{"", ""},
		{"yesterday", ""},
		{"yesterday, today", ""},
		{"yesterday, today, problems", ""},
	}

	wrongBundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err = wrongBundle.LoadMessageFile("active.en.toml")
	assert.Error(t, err)

	bot = &Bot{
		bundle:     wrongBundle,
		properties: properties,
	}

	for _, tt := range testCasesErr {
		problem := bot.analizeStandup(tt.Message)
		assert.Equal(t, tt.Problem, problem)
	}
}

func TestGetAccessLevel(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		SelectedUser        model.User
		FoundStanduper      model.Standuper
		AccessLevel         int
		SelectedUserError   error
		FoundStanduperError error
	}{
		{model.User{}, model.Standuper{}, 0, errors.New("not found"), nil},
		{model.User{UserID: "Foo", Role: "admin"}, model.Standuper{}, 2, nil, nil},
		{model.User{UserID: "Foo", Role: ""}, model.Standuper{}, 4, nil, errors.New("not found")},
		{model.User{UserID: "Foo", Role: ""}, model.Standuper{UserID: "Foo", RoleInChannel: "pm"}, 3, nil, nil},
		{model.User{UserID: "Foo", Role: ""}, model.Standuper{UserID: "Foo", RoleInChannel: "developer"}, 4, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			SelectedUser:        tt.SelectedUser,
			FoundStanduper:      tt.FoundStanduper,
			SelectedUserError:   tt.SelectedUserError,
			FoundStanduperError: tt.FoundStanduperError,
		})

		accessLevel, err := bot.getAccessLevel("Foo", "Bar")
		if err != nil {
			assert.Equal(t, tt.SelectedUserError, err)
		}
		assert.Equal(t, tt.AccessLevel, accessLevel)
	}
}

func TestHandleJoin(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		SelectedChannel      model.Channel
		CreatedChannel       model.Channel
		SelectedChannelError error
		CreatedChannelError  error
	}{
		{model.Channel{}, model.Channel{}, nil, nil},
		{model.Channel{}, model.Channel{}, errors.New("not found"), nil},
	}

	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://slack.com/api/conversations.info", httpmock.NewStringResponder(200, `{"error": true, "channel": {}}`))

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			SelectedChannel:      tt.SelectedChannel,
			CreatedChannel:       tt.CreatedChannel,
			SelectedChannelError: tt.SelectedChannelError,
			CreatedChannelError:  tt.CreatedChannelError,
		})

		ch, err := bot.HandleJoin("Foo", "Bar")
		if err != nil {
			assert.Error(t, err)
		}
		assert.Equal(t, tt.CreatedChannel.ID, ch.ID)
	}

	httpmock.DeactivateAndReset()

	testCases = []struct {
		SelectedChannel      model.Channel
		CreatedChannel       model.Channel
		SelectedChannelError error
		CreatedChannelError  error
	}{
		{model.Channel{}, model.Channel{}, errors.New("not found"), errors.New("could not create")},
		{model.Channel{}, model.Channel{}, errors.New("not found"), nil},
	}

	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://slack.com/api/conversations.info", httpmock.NewStringResponder(200, `{"ok": true, "channel": {"id": "CBAPFA2J2", "name": "general"}}`))

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			SelectedChannel:      tt.SelectedChannel,
			CreatedChannel:       tt.CreatedChannel,
			SelectedChannelError: tt.SelectedChannelError,
			CreatedChannelError:  tt.CreatedChannelError,
		})

		ch, err := bot.HandleJoin("Foo", "Bar")
		if err != nil {
			assert.Error(t, err)
		}
		assert.Equal(t, tt.CreatedChannel.ID, ch.ID)
	}

	httpmock.DeactivateAndReset()
}

func TestSuits(t *testing.T) {
	properties := model.BotSettings{
		TeamID:   "Foo",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: properties,
	}

	ok := bot.Suits("Foo")
	assert.Equal(t, true, ok)
}

func TestSetProperties(t *testing.T) {
	settings := model.BotSettings{
		TeamID:   "Foo",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.ru.toml")
	assert.NoError(t, err)

	bot := &Bot{
		bundle:     bundle,
		properties: settings,
	}

	newSettings := model.BotSettings{
		TeamID:   "Bar",
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot.SetProperties(newSettings)
	assert.Equal(t, "Bar", bot.properties.TeamID)
}

//httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", r)
/*
r := httpmock.NewStringResponder(200, `{
	"ok": true,
	"members": [
		{
			"id": "USER1D1",
			"team_id": "TEAMID1",
			"name": "UserAdmin",
			"deleted": false,
			"color": "9f69e7",
			"real_name": "admin",
			"is_admin": true,
			"is_owner": true,
			"is_primary_owner": true,
			"is_restricted": false,
			"is_ultra_restricted": false,
			"is_bot": false
		},
		{
			"id": "BOTID",
			"team_id": "TEAMID1",
			"name": "comedian",
			"deleted": false,
			"color": "4bbe2e",
			"real_name": "comedian",
			"tz": "America\/Los_Angeles",
			"tz_label": "Pacific Daylight Time",
			"tz_offset": -25200,
			"is_admin": false,
			"is_owner": false,
			"is_primary_owner": false,
			"is_restricted": false,
			"is_ultra_restricted": false,
			"is_bot": true,
			"is_app_user": false,
			"updated": 1529488035
		},
		{
			"id": "UBEGJBB9A",
			"team_id": "TEAMID1",
			"name": "anot",
			"deleted": false,
			"color": "674b1b",
			"real_name": "Anot",
			"is_restricted": false,
			"is_ultra_restricted": false,
			"is_bot": false,
			"is_app_user": false
		},
		{
			"id": "xxx",
			"team_id": "TEAMID1",
			"name": "deleted user",
			"deleted": true,
			"color": "674b1b",
			"real_name": "test user",
			"is_restricted": false,
			"is_ultra_restricted": false,
			"is_bot": false,
			"is_app_user": false
		}
	]
}`)

*/
