package botuser

import (
	"errors"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jarcoal/httpmock"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
	"golang.org/x/text/language"
)

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

	bot := New(c, bundle, settings, db)
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
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
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
		bot := New(c, bundle, settings, MockedDB{
			SelectedUser:        tt.SelectedUser,
			FoundStanduper:      tt.FoundStanduper,
			SelectedUserError:   tt.SelectedUserError,
			FoundStanduperError: tt.FoundStanduperError,
		})

		accessLevel, err := bot.GetAccessLevel("Foo", "Bar")
		if err != nil {
			assert.Equal(t, tt.SelectedUserError, err)
		}
		assert.Equal(t, tt.AccessLevel, accessLevel)
	}
}

func TestHandleJoin(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
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
		bot := New(c, bundle, settings, MockedDB{
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
		bot := New(c, bundle, settings, MockedDB{
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

func TestUpdateUser(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		User                 slack.User
		SelectedUser         model.User
		CreatedUser          model.User
		UpdatedUser          model.User
		Standupers           []model.Standuper
		SelectedUserError    error
		CreatedUserError     error
		UpdatedUserError     error
		ListStandupersError  error
		DeleteUserError      error
		DeleteStanduperError error
	}{
		{slack.User{IsBot: true}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, nil, nil, nil, nil, nil, nil},
		{slack.User{Name: "slackbot"}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, nil, nil, nil, nil, nil, nil},
		{slack.User{Deleted: false, IsAdmin: true, Name: "Foo"}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, nil, nil, errors.New("update user error"), nil, nil, nil},
		{slack.User{Deleted: true}, model.User{ID: int64(1)}, model.User{}, model.User{}, []model.Standuper{}, nil, nil, nil, nil, errors.New("delete user error"), nil},
		{slack.User{Deleted: true}, model.User{ID: int64(1)}, model.User{}, model.User{}, []model.Standuper{}, nil, nil, nil, errors.New("list standupers error"), nil, nil},
		{slack.User{Deleted: true}, model.User{ID: int64(1), UserID: "FOO123"}, model.User{}, model.User{}, []model.Standuper{{UserID: "FOO123"}}, nil, nil, nil, nil, nil, errors.New("delete standuper error")},
		{slack.User{Deleted: false}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, errors.New("select user error"), errors.New("create user error"), nil, nil, nil, nil},
		{slack.User{Deleted: false, IsAdmin: true}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, errors.New("select user error"), errors.New("create user error"), nil, nil, nil, nil},
		{slack.User{Deleted: false, IsAdmin: true}, model.User{}, model.User{}, model.User{}, []model.Standuper{}, errors.New("select user error"), nil, nil, nil, nil, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			SelectedUser:         tt.SelectedUser,
			CreatedUser:          tt.CreatedUser,
			UpdatedUser:          tt.UpdatedUser,
			Standupers:           tt.Standupers,
			SelectedUserError:    tt.SelectedUserError,
			CreatedUserError:     tt.CreatedUserError,
			UpdatedUserError:     tt.UpdatedUserError,
			ListStandupersError:  tt.ListStandupersError,
			DeleteUserError:      tt.DeleteUserError,
			DeleteStanduperError: tt.DeleteStanduperError,
		})

		err := bot.updateUser(tt.User)
		if err != nil {
			assert.Error(t, err)
		}
	}
}

func TestUpdateUsersList(t *testing.T) {
	success := httpmock.NewStringResponder(200, `{
		"ok": true,
		"members": [
			{
				"id": "USER1D1",
				"team_id": "TEAMID1",
				"name": "UserAdmin",
				"deleted": false,
				"color": "9f69e7",
				"real_name": "admin",
				"is_admin": false,
				"is_owner": false,
				"is_primary_owner": false,
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": false
			}
		]
	}`)

	fail := httpmock.NewStringResponder(404, `{"error":true, "members":[]}`)

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", fail)

	bot := New(c, bundle, settings, MockedDB{CreatedUserError: nil})

	bot.UpdateUsersList()
	httpmock.DeactivateAndReset()

	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", success)

	bot = New(c, bundle, settings, MockedDB{CreatedUserError: errors.New("err"), SelectedUserError: errors.New("err")})

	bot.UpdateUsersList()
	httpmock.DeactivateAndReset()
}

func TestSendMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postEphemeral", httpmock.NewStringResponder(200, `{"ok": true}`))

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(c, bundle, settings, MockedDB{})
	err = bot.SendMessage("YYYZZZVVV", "Hey!", nil)
	assert.NoError(t, err)

	err = bot.SendEphemeralMessage("YYYZZZVVV", "USER!", "Hey!")
	assert.NoError(t, err)
}

func TestSendUserMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `{"ok": true}`))

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(c, bundle, settings, MockedDB{})

	err = bot.SendUserMessage("USLACKBOT", "MSG to User!")
	assert.NoError(t, err)
}

func TestHandleMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postMessage", httpmock.NewStringResponder(200, `"ok": true`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/reactions.add", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/im.open", httpmock.NewStringResponder(200, `{"ok": true}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/chat.postEphemeral", httpmock.NewStringResponder(200, `{"ok": true}`))

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		text                          string
		subType                       string
		SelectedStandupByMessageTS    model.Standup
		FoundStanduper                model.Standuper
		UpdatedStanduper              model.Standuper
		CreatedStandup                model.Standup
		UpdatedStandup                model.Standup
		SelectStandupByMessageTSError error
		FoundStanduperError           error
		UpdateStanduperError          error
		DeleteStandupError            error
		CreateStandupError            error
		UpdateStandupError            error
		SubmittedStandupTodayResult   bool
		SubmittedStandupTodayError    error
	}{
		{"Lorem ipsum...", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},
		{"TESTUSERID Lorem ipsum...", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, true, errors.New("submitted standup")},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, true, nil},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, errors.New("create standup"), nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{ID: int64(1)}, model.Standup{}, nil, errors.New("standuper not found"), nil, nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{ID: int64(1)}, model.Standuper{}, model.Standup{ID: int64(1)}, model.Standup{}, nil, nil, errors.New("update standuper"), nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeMessage, model.Standup{}, model.Standuper{ID: int64(1)}, model.Standuper{}, model.Standup{ID: int64(1)}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},

		{"Lorem ipsum...", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},
		{"TESTUSERID Lorem ipsum...", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{ID: int64(1)}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, errors.New("update standup"), false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{ID: int64(1)}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, nil, nil, true, errors.New("submitted standup")},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, nil, nil, true, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, errors.New("create standup"), nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), errors.New("standuper not found"), nil, nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, errors.New("update standuper"), nil, nil, nil, false, nil},
		{"TESTUSERID yesterday, today, problems", typeEditMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, nil, nil, false, nil},

		{"Lorem ipsum...", typeDeleteMessage, model.Standup{}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, errors.New("err"), nil, nil, nil, nil, nil, false, nil},
		{"Lorem ipsum...", typeDeleteMessage, model.Standup{ID: int64(1)}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, errors.New("err"), nil, nil, false, nil},
		{"Lorem ipsum...", typeDeleteMessage, model.Standup{ID: int64(1)}, model.Standuper{}, model.Standuper{}, model.Standup{}, model.Standup{}, nil, nil, nil, nil, nil, nil, false, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			FoundStanduper:                tt.FoundStanduper,
			SelectedStandupByMessageTS:    tt.SelectedStandupByMessageTS,
			UpdatedStanduper:              tt.UpdatedStanduper,
			SelectStandupByMessageTSError: tt.SelectStandupByMessageTSError,
			UpdateStanduperError:          tt.UpdateStanduperError,
			DeleteStandupError:            tt.DeleteStandupError,
			FoundStanduperError:           tt.FoundStanduperError,
			CreatedStandup:                tt.CreatedStandup,
			CreateStandupError:            tt.CreateStandupError,
			UpdateStandupError:            tt.UpdateStandupError,
			SubmittedStandupTodayResult:   tt.SubmittedStandupTodayResult,
			SubmittedStandupTodayError:    tt.SubmittedStandupTodayError,
		})

		msg := &slack.MessageEvent{}
		msg.Text = tt.text
		msg.User = "Foo"
		msg.Channel = "Bar"
		msg.Timestamp = "15000"
		msg.SubType = tt.subType

		if tt.subType == typeEditMessage {
			msg = &slack.MessageEvent{
				SubMessage: &slack.Msg{
					User:      "Foo",
					Text:      tt.text,
					Timestamp: "15000",
				},
			}
			msg.Channel = "Bar"
			msg.SubType = tt.subType
		}

		bot.HandleMessage(msg)
	}
}

func TestImplementCommands(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(c, bundle, settings, MockedDB{
		SelectedChannelError: errors.New("no channel"),
	})

	testCases := []struct {
		command     string
		params      string
		accessLevel int
		expected    string
	}{
		{"foo", "", 0, "To add members use `add` command. Here is an example: `add @user @user1 / pm` You can add members with _pm, developer, designer, tester_ roles, default is a developer role, if the role is not selected. \nTo view members use `show` command\nTo remove members use `remove` command and tag members to remove\nTo set standup deadline use `add_deadline` command.\nTo view standup deadline in the channel use `show_deadline` command. \nTo remove standup deadline in the channel use `remove_deadline` command. \n"},
		{"add", "", 4, "Access Denied! You need to be at least PM in this project to use this command!"},
		{"show", "foo", 0, "No standupers in this channel! To add one, please, use `/comedian add` slash command"},
		{"remove", "", 4, "Access Denied! You need to be at least PM in this project to use this command!"},
		{"add_deadline", "", 4, "Access Denied! You need to be at least PM in this project to use this command!"},
		{"remove_deadline", "", 4, "Access Denied! You need to be at least PM in this project to use this command!"},
		{"show_deadline", "", 0, "No standup time set for this channel yet! Please, add a standup time using `/comedian add_deadline` command!"},
	}

	for _, tt := range testCases {
		message := bot.ImplementCommands("Foo", tt.command, tt.params, tt.accessLevel)
		assert.Equal(t, tt.expected, message)
	}

}

func TestStart(t *testing.T) {
	success := httpmock.NewStringResponder(200, `{
		"ok": true,
		"members": [
			{
				"id": "USER1D1",
				"team_id": "TEAMID1",
				"name": "UserAdmin",
				"deleted": false,
				"color": "9f69e7",
				"real_name": "admin",
				"is_admin": false,
				"is_owner": false,
				"is_primary_owner": false,
				"is_restricted": false,
				"is_ultra_restricted": false,
				"is_bot": false
			}
		]
	}`)

	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	c, err := config.Get()
	assert.NoError(t, err)

	_, err = bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	bot := New(c, bundle, settings, MockedDB{
		ListedUsers: []model.User{},
	})

	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://slack.com/api/users.list", success)

	Dry = true
	bot.Start()
	httpmock.DeactivateAndReset()

}
