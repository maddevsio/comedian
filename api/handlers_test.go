package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

type MockedDBForBot struct {
	storage.Storage
	AllBotSettings         []model.BotSettings
	GetAllBotSettingsError error

	BotSettings         model.BotSettings
	GetBotSettingsError error

	UpdatedBotSettings     model.BotSettings
	UpdateBotSettingsError error

	DeleteBotSettingsError error
}

func (m MockedDBForBot) GetAllBotSettings() ([]model.BotSettings, error) {
	return m.AllBotSettings, m.GetAllBotSettingsError
}

func (m MockedDBForBot) GetBotSettings(id int64) (model.BotSettings, error) {
	return m.BotSettings, m.GetBotSettingsError
}

func (m MockedDBForBot) UpdateBotSettings(bot model.BotSettings) (model.BotSettings, error) {
	return m.UpdatedBotSettings, m.UpdateBotSettingsError
}

func (m MockedDBForBot) DeleteBotSettingsByID(id int64) error {
	return m.DeleteBotSettingsError
}

func TestBotHandlers(t *testing.T) {

	testCases := []struct {
		AllBotSettings         []model.BotSettings
		GetAllBotSettingsError error
		StatusCode             int
	}{
		{[]model.BotSettings{}, errors.New("err"), 404},
		{[]model.BotSettings{model.BotSettings{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDBForBot{
			AllBotSettings:         tt.AllBotSettings,
			GetAllBotSettingsError: tt.GetAllBotSettingsError,
		}}

		// Setup
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Assertions
		if assert.NoError(t, r.listBots(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}
