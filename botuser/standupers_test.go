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
		{"Could not assign member: @foo.", 3, "@foo", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"To add members use `add` command. Here is an example: `add @user @user1 / admin` You can add members with _admin, pm, developer, designer_ roles, default is a developer role, if the role is not selected. ", 3, "@foo / bar", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"User <@foo|bar> is assigned as PM.", 3, "<@foo|bar> /pm", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar> /designer", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Access Denied! You need to be at least admin in this slack to use this command!", 3, "<@foo|bar> /admin", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Could not assign user as admin: @foo!\n", 2, "@foo /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> already has role.", 3, "<@foo|bar>", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Could not assign member: <@foo|bar>.", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, errors.New("create standuper"), model.User{}, nil, model.User{}, nil},
		{"Could not assign user as admin: <@foo|bar>!\n", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, errors.New("select user"), model.User{}, nil},
		{"User <@foo|bar> was already assigned as admin.\n", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{Role: "admin"}, nil, model.User{}, nil},
		{"Could not assign user as admin: <@foo|bar>!\n", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, errors.New("updated user")},
		{"User <@foo|bar> is assigned as admin.", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
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

func TestAddCommandBundleFail(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

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
		{"", 4, "", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "@foo", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"All help!", 3, "@foo / bar", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar> /pm", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar> /designer", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar> /admin", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 2, "@foo /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar>", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, errors.New("create standuper"), model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar> /pm", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar> /pm", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, errors.New("create standuper"), model.User{}, nil, model.User{}, nil},
		{"", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, errors.New("select user"), model.User{}, nil},
		{"", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{Role: "admin"}, nil, model.User{}, nil},
		{"", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, errors.New("updated user")},
		{"", 2, "<@foo|bar> /admin", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
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

func TestShowCommand(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	assert.NoError(t, err)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		params                     string
		ListedUsers                []model.User
		ListedUsersError           error
		ChannelStandupers          []model.Standuper
		ListChannelStandupersError error
	}{
		{"To view members use `show` command. If you provide a role name, you will see members with this role. _admin, pm, developer, designer_. ", "foo", []model.User{}, nil, []model.Standuper{}, nil},
		{"could not list users", "admin", []model.User{}, errors.New("err"), []model.Standuper{}, nil},
		{"could not list members", "developer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"could not list members", "designer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"could not list members", "pm", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"No admins in this workspace! To add one, please, use `/comedian add` slash command", "admin", []model.User{}, nil, []model.Standuper{}, nil},
		{"Admin in this workspace: <@Foo>", "admin", []model.User{{UserName: "Foo", Role: "admin"}}, nil, []model.Standuper{}, nil},
		{"Standuper in this channel: <@FOO>", "developer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "developer", UserID: "FOO"}}, nil},
		{"Standuper in this channel: <@FOO>", "designer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "designer", UserID: "FOO"}}, nil},
		{"PM in this channel: <@FOO>", "pm", []model.User{}, nil, []model.Standuper{{RoleInChannel: "pm", UserID: "FOO"}}, nil},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "designer", []model.User{}, nil, []model.Standuper{}, nil},
		{"No PMs in this channel! To add one, please, use `/comedian add` slash command", "pm", []model.User{}, nil, []model.Standuper{}, nil},
	}

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			ListedUsers:                tt.ListedUsers,
			ListedUsersError:           tt.ListedUsersError,
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		text := bot.showCommand("Foo", tt.params)
		assert.Equal(t, tt.expectedText, text)
	}
}

func TestShowCommandBundleFail(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	settings := model.BotSettings{
		UserID:   "TESTUSERID",
		Language: "en_US",
	}

	testCases := []struct {
		expectedText               string
		params                     string
		ListedUsers                []model.User
		ListedUsersError           error
		ChannelStandupers          []model.Standuper
		ListChannelStandupersError error
	}{
		{"All help!", "foo", []model.User{}, nil, []model.Standuper{}, nil},
		{"could not list users", "admin", []model.User{}, errors.New("err"), []model.Standuper{}, nil},
		{"could not list members", "developer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"could not list members", "designer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"could not list members", "pm", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"", "admin", []model.User{}, nil, []model.Standuper{}, nil},
		{"", "admin", []model.User{{UserName: "Foo", Role: "admin"}}, nil, []model.Standuper{}, nil},
		{"", "developer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "developer", UserID: "FOO"}}, nil},
		{"", "designer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "designer", UserID: "FOO"}}, nil},
		{"", "pm", []model.User{}, nil, []model.Standuper{{RoleInChannel: "pm", UserID: "FOO"}}, nil},
		{"", "designer", []model.User{}, nil, []model.Standuper{}, nil},
		{"", "pm", []model.User{}, nil, []model.Standuper{}, nil},
	}

	for _, tt := range testCases {
		bot := New(bundle, settings, MockedDB{
			ListedUsers:                tt.ListedUsers,
			ListedUsersError:           tt.ListedUsersError,
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		text := bot.showCommand("Foo", tt.params)
		assert.Equal(t, tt.expectedText, text)
	}
}
