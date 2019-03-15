package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/storage"
)

func TestRenderLoginPage(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)
	api := New(c, db, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/login", strings.NewReader(""))
	rec := httptest.NewRecorder()
	cntx := e.NewContext(req, rec)
	err = api.renderLoginPage(cntx)
	assert.Error(t, err)
}

func TestRenderControlPannel(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := storage.New(c)
	assert.NoError(t, err)
	api := New(c, db, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin", strings.NewReader(""))
	rec := httptest.NewRecorder()
	cntx := e.NewContext(req, rec)
	err = api.renderControlPannel(cntx)
	assert.Error(t, err)
}

// func TestUpdateConfig(t *testing.T) {
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	db, err := storage.New(c)
// 	assert.NoError(t, err)
// 	api := New(c, db, nil)
// 	e := echo.New()
// 	req := httptest.NewRequest(http.MethodPost, "/config", strings.NewReader(""))
// 	rec := httptest.NewRecorder()
// 	cntx := e.NewContext(req, rec)
// 	err = api.updateConfig(cntx)
// 	assert.Error(t, err)
// }
