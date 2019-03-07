package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

type MockedDBForBot struct {
	storage.Storage
	AllBotSettings     []model.BotSettings
	BotSettings        model.BotSettings
	UpdatedBotSettings model.BotSettings
	Error              error
}

func (m MockedDBForBot) GetAllBotSettings() ([]model.BotSettings, error) {
	return m.AllBotSettings, m.Error
}

func (m MockedDBForBot) GetBotSettings(id int64) (model.BotSettings, error) {
	return m.BotSettings, m.Error
}

func (m MockedDBForBot) UpdateBotSettings(bot model.BotSettings) (model.BotSettings, error) {
	return m.UpdatedBotSettings, m.Error
}

func (m MockedDBForBot) DeleteBotSettingsByID(id int64) error {
	return m.Error
}

func TestHealthCheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	r := &RESTAPI{db: MockedDBForBot{}}

	if assert.NoError(t, r.healthcheck(c)) {
		assert.Equal(t, 200, rec.Code)
	}
}

func TestListBots(t *testing.T) {

	testCases := []struct {
		AllBotSettings []model.BotSettings
		Error          error
		StatusCode     int
	}{
		{[]model.BotSettings{}, errors.New("err"), 404},
		{[]model.BotSettings{model.BotSettings{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDBForBot{
			AllBotSettings: tt.AllBotSettings,
			Error:          tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, r.listBots(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestGetBot(t *testing.T) {

	testCases := []struct {
		BotSettings model.BotSettings
		Error       error
		ID          string
		StatusCode  int
	}{
		{model.BotSettings{}, errors.New("err"), "", 406},
		{model.BotSettings{}, errors.New("err"), "1", 404},
		{model.BotSettings{}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDBForBot{
			BotSettings: tt.BotSettings,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)

		if assert.NoError(t, r.getBot(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestUpdateBot(t *testing.T) {

	testCases := []struct {
		BotSettings model.BotSettings
		Error       error
		ID          string
		formValues  map[string]string
		StatusCode  int
	}{
		{model.BotSettings{}, errors.New("err"), "", map[string]string{}, 406},
		{model.BotSettings{}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 404},
		{model.BotSettings{}, nil, "1", map[string]string{"password": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDBForBot{
			BotSettings: tt.BotSettings,
			Error:       tt.Error,
		}}

		e := echo.New()
		f := make(url.Values)
		for k, v := range tt.formValues {
			f.Set(k, v)
		}

		req := httptest.NewRequest(http.MethodPost, "/bots", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)

		if assert.NoError(t, r.updateBot(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestDeleteBot(t *testing.T) {

	testCases := []struct {
		BotSettings model.BotSettings
		Error       error
		ID          string
		StatusCode  int
	}{
		{model.BotSettings{}, errors.New("err"), "", 406},
		{model.BotSettings{}, errors.New("err"), "1", 404},
		{model.BotSettings{}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDBForBot{
			BotSettings: tt.BotSettings,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)

		if assert.NoError(t, r.deleteBot(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}
