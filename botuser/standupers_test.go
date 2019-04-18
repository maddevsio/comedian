package botuser

import (
	"errors"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/require"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
	"golang.org/x/text/language"
)

func TestAddCommand(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	c, err := config.Get()
	require.NoError(t, err)

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
		{"To add members use `add` command. Here is an example: `add @user @user1 / pm` You can add members with _pm, developer, designer, tester_ roles, default is a developer role, if the role is not selected. ", 3, "@foo / bar", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"User <@foo|bar> is assigned as PM.", 3, "<@foo|bar> /pm", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar> /designer", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Member <@foo|bar> is assigned.\n", 3, "<@foo|bar>", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
		{"Could not assign member: <@foo|bar>.", 3, "<@foo|bar>", model.Standuper{}, errors.New("select standuper"), model.Standuper{}, errors.New("create standuper"), model.User{}, nil, model.User{}, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
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
		require.Equal(t, tt.expectedText, text)
	}
}

func TestAddCommandBundleFail(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	c, err := config.Get()
	require.NoError(t, err)

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
		{"", 3, "@foo / bar", model.Standuper{}, nil, model.Standuper{}, nil, model.User{}, nil, model.User{}, nil},
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
		bot := New(c, bundle, settings, MockedDB{
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
		require.Equal(t, tt.expectedText, text)
	}
}

func TestShowCommand(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	c, err := config.Get()
	require.NoError(t, err)

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
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "foo", []model.User{}, nil, []model.Standuper{}, nil},
		{"No admins in this workspace! To add one, please, add admin in your slack settings", "admin", []model.User{}, errors.New("err"), []model.Standuper{}, nil},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "developer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "designer", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "pm", []model.User{}, errors.New("err"), []model.Standuper{}, errors.New("err")},
		{"No admins in this workspace! To add one, please, add admin in your slack settings", "admin", []model.User{}, nil, []model.Standuper{}, nil},
		{"Admin in this workspace: <@Foo>", "admin", []model.User{{UserName: "Foo", Role: "admin"}}, nil, []model.Standuper{}, nil},
		{"No PMs in this channel!\nDeveloper in the project: <@FOO>\n", "developer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "developer", UserID: "FOO"}}, nil},
		{"No PMs in this channel!\nTester in the project: <@FOO>\n", "tester", []model.User{}, nil, []model.Standuper{{RoleInChannel: "tester", UserID: "FOO"}}, nil},
		{"No PMs in this channel!\nDesigner in the project: <@FOO>\n", "designer", []model.User{}, nil, []model.Standuper{{RoleInChannel: "designer", UserID: "FOO"}}, nil},
		{"PM in the project: <@FOO>\n", "pm", []model.User{}, nil, []model.Standuper{{RoleInChannel: "pm", UserID: "FOO"}}, nil},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "designer", []model.User{}, nil, []model.Standuper{}, nil},
		{"No standupers in this channel! To add one, please, use `/comedian add` slash command", "pm", []model.User{}, nil, []model.Standuper{}, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			ListedUsers:                tt.ListedUsers,
			ListedUsersError:           tt.ListedUsersError,
			ChannelStandupers:          tt.ChannelStandupers,
			ListChannelStandupersError: tt.ListChannelStandupersError,
		})

		text := bot.showCommand("Foo", tt.params)
		require.Equal(t, tt.expectedText, text)
	}
}

func TestDeleteCommand(t *testing.T) {
	bundle := &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	_, err := bundle.LoadMessageFile("../active.en.toml")
	require.NoError(t, err)

	c, err := config.Get()
	require.NoError(t, err)

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
		DeleteStanduperError error
		SelectedUser         model.User
		SelectedUserError    error
		UpdatedUser          model.User
		UpdatedUserError     error
	}{
		{"Access Denied! You need to be at least PM in this project to use this command!", 4, "", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"Could not remove the member: @foo . User is not standuper and not tracked. \n", 3, "@foo", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"Could not remove the following members: @foo, /, bar . Users are not standupers and not tracked. \n", 3, "@foo / bar", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"Could not remove the member: <@foo|bar> . User is not standuper and not tracked. \n", 3, "<@foo|bar>", model.Standuper{}, errors.New("member error"), nil, model.User{}, nil, model.User{}, nil},
		{"The member <@foo|bar> removed.", 3, "<@foo|bar>", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, nil, model.User{}, nil, model.User{}, nil},
	}

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			FoundStanduper:       tt.FoundStanduper,
			FoundStanduperError:  tt.FoundStanduperError,
			DeleteStanduperError: tt.DeleteStanduperError,
			SelectedUser:         tt.SelectedUser,
			SelectedUserError:    tt.SelectedUserError,
			UpdatedUser:          tt.UpdatedUser,
			UpdatedUserError:     tt.UpdatedUserError,
		})

		text := bot.deleteCommand(tt.accessLevel, "Foo", tt.params)
		require.Equal(t, tt.expectedText, text)
	}
}

func TestDeleteCommandBundleFailed(t *testing.T) {
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
		DeleteStanduperError error
		SelectedUser         model.User
		SelectedUserError    error
		UpdatedUser          model.User
		UpdatedUserError     error
	}{
		{"", 4, "", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "@foo", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "@foo / bar", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"", 2, "@foo /admin", model.Standuper{}, nil, nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar>", model.Standuper{}, errors.New("member error"), nil, model.User{}, nil, model.User{}, nil},
		{"", 3, "<@foo|bar>", model.Standuper{ChannelID: "Foo", UserID: "foo"}, nil, nil, model.User{}, nil, model.User{}, nil},
	}

	c, err := config.Get()
	require.NoError(t, err)

	for _, tt := range testCases {
		bot := New(c, bundle, settings, MockedDB{
			FoundStanduper:       tt.FoundStanduper,
			FoundStanduperError:  tt.FoundStanduperError,
			DeleteStanduperError: tt.DeleteStanduperError,
			SelectedUser:         tt.SelectedUser,
			SelectedUserError:    tt.SelectedUserError,
			UpdatedUser:          tt.UpdatedUser,
			UpdatedUserError:     tt.UpdatedUserError,
		})

		text := bot.deleteCommand(tt.accessLevel, "Foo", tt.params)
		require.Equal(t, tt.expectedText, text)
	}
}
