package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

type MockedDB struct {
	storage.Storage
	AllBotSettings     []model.BotSettings
	BotSettings        model.BotSettings
	UpdatedBotSettings model.BotSettings

	AllStandups    []model.Standup
	Standup        model.Standup
	UpdatedStandup model.Standup

	AllUsers    []model.User
	User        model.User
	UpdatedUser model.User

	AllChannels    []model.Channel
	Channel        model.Channel
	UpdatedChannel model.Channel

	AllStandupers    []model.Standuper
	Standuper        model.Standuper
	UpdatedStanduper model.Standuper

	Error error
}

func (m MockedDB) GetAllBotSettings() ([]model.BotSettings, error) {
	return m.AllBotSettings, m.Error
}

func (m MockedDB) GetBotSettings(id int64) (model.BotSettings, error) {
	return m.BotSettings, m.Error
}

func (m MockedDB) UpdateBotSettings(input model.BotSettings) (model.BotSettings, error) {
	return m.UpdatedBotSettings, m.Error
}

func (m MockedDB) DeleteBotSettingsByID(id int64) error {
	return m.Error
}

func (m MockedDB) ListStandups() ([]model.Standup, error) {
	return m.AllStandups, m.Error
}

func (m MockedDB) GetStandup(id int64) (model.Standup, error) {
	return m.Standup, m.Error
}

func (m MockedDB) UpdateStandup(input model.Standup) (model.Standup, error) {
	return m.UpdatedStandup, m.Error
}

func (m MockedDB) DeleteStandup(id int64) error {
	return m.Error
}

func (m MockedDB) ListUsers() ([]model.User, error) {
	return m.AllUsers, m.Error
}

func (m MockedDB) GetUser(id int64) (model.User, error) {
	return m.User, m.Error
}

func (m MockedDB) UpdateUser(input model.User) (model.User, error) {
	return m.UpdatedUser, m.Error
}

func (m MockedDB) ListChannels() ([]model.Channel, error) {
	return m.AllChannels, m.Error
}

func (m MockedDB) GetChannel(id int64) (model.Channel, error) {
	return m.Channel, m.Error
}

func (m MockedDB) UpdateChannel(input model.Channel) (model.Channel, error) {
	return m.UpdatedChannel, m.Error
}

func (m MockedDB) DeleteChannel(id int64) error {
	return m.Error
}

func (m MockedDB) ListStandupers() ([]model.Standuper, error) {
	return m.AllStandupers, m.Error
}

func (m MockedDB) GetStanduper(id int64) (model.Standuper, error) {
	return m.Standuper, m.Error
}

func (m MockedDB) UpdateStanduper(input model.Standuper) (model.Standuper, error) {
	return m.UpdatedStanduper, m.Error
}

func (m MockedDB) DeleteStanduper(id int64) error {
	return m.Error
}

func (m MockedDB) GetBotSettingsByTeamName(teamName string) (model.BotSettings, error) {
	return m.BotSettings, m.Error
}

func TestHealthCheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", strings.NewReader(""))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	r := &RESTAPI{db: MockedDB{}}

	if assert.NoError(t, r.healthcheck(c)) {
		assert.Equal(t, 200, rec.Code)
	}
}

/* Bots functionaliy */
func TestListBots(t *testing.T) {

	testCases := []struct {
		AllBotSettings []model.BotSettings
		Error          error
		StatusCode     int
	}{
		{[]model.BotSettings{}, errors.New("err"), 500},
		{[]model.BotSettings{model.BotSettings{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			AllBotSettings: tt.AllBotSettings,
			Error:          tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"bot_id": float64(1)},
		})

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
		{model.BotSettings{}, errors.New("err"), "", 400},
		{model.BotSettings{}, errors.New("err"), "1", 404},
		{model.BotSettings{}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			BotSettings: tt.BotSettings,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"bot_id": float64(1)},
		})

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
		{model.BotSettings{}, errors.New("err"), "", map[string]string{}, 400},
		{model.BotSettings{}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 500},
		{model.BotSettings{}, nil, "1", map[string]string{"password": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
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
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"bot_id": float64(1)},
		})

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
		{model.BotSettings{}, errors.New("err"), "", 400},
		{model.BotSettings{}, errors.New("err"), "1", 500},
		{model.BotSettings{}, nil, "1", 204},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			BotSettings: tt.BotSettings,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/bots", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"bot_id": float64(1)},
		})

		if assert.NoError(t, r.deleteBot(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

/* Standups functionaliy */
func TestListStandups(t *testing.T) {

	testCases := []struct {
		AllStandups []model.Standup
		Error       error
		StatusCode  int
	}{
		{[]model.Standup{}, errors.New("err"), 500},
		{[]model.Standup{model.Standup{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			AllStandups: tt.AllStandups,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/standups", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.listStandups(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestGetStandup(t *testing.T) {

	testCases := []struct {
		Standup    model.Standup
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Standup{}, errors.New("err"), "", 400},
		{model.Standup{}, errors.New("err"), "1", 404},
		{model.Standup{TeamID: "foo"}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standup: tt.Standup,
			Error:   tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/standups", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.getStandup(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestUpdateStandup(t *testing.T) {

	testCases := []struct {
		Standup    model.Standup
		Error      error
		ID         string
		formValues map[string]string
		StatusCode int
	}{
		{model.Standup{TeamID: "foo"}, errors.New("err"), "", map[string]string{}, 400},
		{model.Standup{TeamID: "foo"}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 500},
		{model.Standup{TeamID: "foo"}, nil, "1", map[string]string{"password": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standup: tt.Standup,
			Error:   tt.Error,
		}}

		e := echo.New()
		f := make(url.Values)
		for k, v := range tt.formValues {
			f.Set(k, v)
		}

		req := httptest.NewRequest(http.MethodPost, "/standups", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": ""},
		})

		if assert.NoError(t, r.updateStandup(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestDeleteStandup(t *testing.T) {

	testCases := []struct {
		Standup    model.Standup
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Standup{}, errors.New("err"), "", 400},
		{model.Standup{}, errors.New("err"), "1", 500},
		{model.Standup{TeamID: "foo"}, nil, "1", 204},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standup: tt.Standup,
			Error:   tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/standups", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.deleteStandup(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

/* Users functionaliy */
func TestListUsers(t *testing.T) {

	testCases := []struct {
		AllUsers   []model.User
		Error      error
		StatusCode int
	}{
		{[]model.User{}, errors.New("err"), 500},
		{[]model.User{model.User{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			AllUsers: tt.AllUsers,
			Error:    tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.listUsers(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestGetUser(t *testing.T) {

	testCases := []struct {
		User       model.User
		Error      error
		ID         string
		StatusCode int
	}{
		{model.User{}, errors.New("err"), "", 400},
		{model.User{}, errors.New("err"), "1", 404},
		{model.User{TeamID: "foo"}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			User:  tt.User,
			Error: tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.getUser(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestUpdateUser(t *testing.T) {

	testCases := []struct {
		User       model.User
		Error      error
		ID         string
		formValues map[string]string
		StatusCode int
	}{
		{model.User{}, errors.New("err"), "", map[string]string{}, 400},
		{model.User{}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 500},
		{model.User{}, nil, "1", map[string]string{"password": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			User:  tt.User,
			Error: tt.Error,
		}}

		e := echo.New()
		f := make(url.Values)
		for k, v := range tt.formValues {
			f.Set(k, v)
		}

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": ""},
		})

		if assert.NoError(t, r.updateUser(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

/* Channels functionaliy */
func TestListChannels(t *testing.T) {

	testCases := []struct {
		AllChannels []model.Channel
		Error       error
		StatusCode  int
	}{
		{[]model.Channel{}, errors.New("err"), 500},
		{[]model.Channel{model.Channel{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			AllChannels: tt.AllChannels,
			Error:       tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/channels", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.listChannels(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestGetChannel(t *testing.T) {

	testCases := []struct {
		Channel    model.Channel
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Channel{}, errors.New("err"), "", 400},
		{model.Channel{}, errors.New("err"), "1", 404},
		{model.Channel{TeamID: "foo"}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Channel: tt.Channel,
			Error:   tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/channels", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.getChannel(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestUpdateChannel(t *testing.T) {

	testCases := []struct {
		Channel    model.Channel
		Error      error
		ID         string
		formValues map[string]string
		StatusCode int
	}{
		{model.Channel{}, errors.New("err"), "", map[string]string{}, 400},
		{model.Channel{}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 500},
		{model.Channel{}, nil, "1", map[string]string{"password": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Channel: tt.Channel,
			Error:   tt.Error,
		}}

		e := echo.New()
		f := make(url.Values)
		for k, v := range tt.formValues {
			f.Set(k, v)
		}

		req := httptest.NewRequest(http.MethodPost, "/channels", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": ""},
		})

		if assert.NoError(t, r.updateChannel(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestDeleteChannel(t *testing.T) {

	testCases := []struct {
		Channel    model.Channel
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Channel{}, errors.New("err"), "", 400},
		{model.Channel{}, errors.New("err"), "1", 500},
		{model.Channel{TeamID: "foo"}, nil, "1", 204},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Channel: tt.Channel,
			Error:   tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/channels", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.deleteChannel(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

/* Standupers functionaliy */
func TestListStandupers(t *testing.T) {

	testCases := []struct {
		AllStandupers []model.Standuper
		Error         error
		StatusCode    int
	}{
		{[]model.Standuper{}, errors.New("err"), 500},
		{[]model.Standuper{model.Standuper{}}, nil, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			AllStandupers: tt.AllStandupers,
			Error:         tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/standupers", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.listStandupers(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestGetStanduper(t *testing.T) {

	testCases := []struct {
		Standuper  model.Standuper
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Standuper{}, errors.New("err"), "", 400},
		{model.Standuper{}, errors.New("err"), "1", 404},
		{model.Standuper{TeamID: "foo"}, nil, "1", 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standuper: tt.Standuper,
			Error:     tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/standupers", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.getStanduper(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestUpdateStanduper(t *testing.T) {

	testCases := []struct {
		Standuper  model.Standuper
		Error      error
		ID         string
		formValues map[string]string
		StatusCode int
	}{
		{model.Standuper{}, errors.New("err"), "", map[string]string{}, 400},
		{model.Standuper{}, errors.New("err"), "1", map[string]string{"pass": "foo"}, 500},
		{model.Standuper{}, nil, "1", map[string]string{"team_id": "foo"}, 200},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standuper: tt.Standuper,
			Error:     tt.Error,
		}}

		e := echo.New()
		f := make(url.Values)
		for k, v := range tt.formValues {
			f.Set(k, v)
		}

		req := httptest.NewRequest(http.MethodPost, "/standupers", strings.NewReader(f.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": ""},
		})

		if assert.NoError(t, r.updateStanduper(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestDeleteStanduper(t *testing.T) {

	testCases := []struct {
		Standuper  model.Standuper
		Error      error
		ID         string
		StatusCode int
	}{
		{model.Standuper{}, errors.New("err"), "", 400},
		{model.Standuper{}, errors.New("err"), "1", 500},
		{model.Standuper{TeamID: "foo"}, nil, "1", 204},
	}

	for _, tt := range testCases {
		r := &RESTAPI{db: MockedDB{
			Standuper: tt.Standuper,
			Error:     tt.Error,
		}}

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/standupers", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames("id")
		c.SetParamValues(tt.ID)
		c.Set("user", &jwt.Token{
			Raw:    "token",
			Claims: jwt.MapClaims{"team_id": "foo"},
		})

		if assert.NoError(t, r.deleteStanduper(c)) {
			assert.Equal(t, tt.StatusCode, rec.Code)
		}
	}
}

func TestLogin(t *testing.T) {

	var testCase = []struct {
		BotSettings model.BotSettings
		FormValues  map[string]string
		Error       error
		StatusCode  int
	}{
		{model.BotSettings{}, map[string]string{"teamname": "team", "password": "root"}, errors.New("err"), 404},
		{model.BotSettings{Password: "$2a$10$paoqyMUatSfoQdVbOfD9PeVAFaa5o1oNou6OPNwR2hv2ikweVdnCC"}, map[string]string{"teamname": "testComedian", "password": "testComedian"}, nil, 200},
		{model.BotSettings{}, map[string]string{}, errors.New("err"), 400},
		{model.BotSettings{Password: "$2a$10$paoqyMUatSfoQdVbOfD9PeVAFaa5o1oNou6OPNwR2hv2ikweVdnCC"}, map[string]string{"teamname": "testComedian", "password": "team"}, nil, 400},
		{model.BotSettings{Password: "$2a$10$paoqyMUatSfoQdVbOfD9PeVAFaa5o1oNou6OPNwR2hv2ikweVdnCC"}, map[string]string{"teamname": "testComedian", "password": "testComedian"}, nil, 200},
	}

	for _, tt := range testCase {
		t.Run("TestLogin", func(t *testing.T) {

			r := &RESTAPI{db: MockedDB{
				BotSettings: tt.BotSettings,
				Error:       tt.Error,
			}}

			e := echo.New()
			f := make(url.Values)
			for k, v := range tt.FormValues {
				f.Set(k, v)
			}
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(f.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if assert.NoError(t, r.login(c)) {
				assert.Equal(t, tt.StatusCode, rec.Code)
			}
		})
	}
}
