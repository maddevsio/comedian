package botuser

import (
	"errors"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"golang.org/x/text/language"
)

/*
FoundStanduper             model.Standuper
FoundStanduperError        error
CreatedStanduper           model.Standuper
CreateStanduperError       error
SelectedUser      model.User
SelectedUserError error
UpdatedUser       model.User
UpdatedUserError  error
ChannelStandupers          []model.Standuper
ListChannelStandupersError error
ListedUsers       []model.User
ListedUsersError  error
DeleteStanduperError       error
*/

func TestAddCommand(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText         string
		accessLevel          int
		params               string
		FoundStanduper       model.Standuper
		FoundStanduperError  error
		CreatedStanduper     model.Standuper
		CreateStanduperError error
		SelectedUser         model.User
		SelectedUserError    error
		UpdatedUser          model.User
		UpdatedUserError     error
	}{
		{"Access Denied! You need to be at least PM in this project to use this command!", 4, "", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Could not assign member: @foo . Wrong username.", 3, "@foo", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, default is a developer role, if the role is not selected. ", 3, "@foo / bar", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"User <@foo|bar> is assigned as PM.", 3, "<@foo|bar> /pm", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar> /designer", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Access Denied! You need to be at least admin in this slack to use this command!", 3, "<@foo|bar> /admin", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Could not assign user as admin: @foo!\n", 2, "@foo /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
	}

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			FoundStanduper:       tt.FoundStanduper,
			FoundStanduperError:  tt.FoundStanduperError,
			CreatedStanduper:     tt.CreatedStanduper,
			CreateStanduperError: tt.CreateStanduperError,
			SelectedUser:         tt.SelectedUser,
			SelectedUserError:    tt.SelectedUserError,
			UpdatedUser:          tt.UpdatedUser,
			UpdatedUserError:     tt.UpdatedUserError,
		})

		text := bot.addCommand(tt.accessLevel, "Foo", tt.params)
		assert.Equal(t, tt.expectedText, text)
	}
}
