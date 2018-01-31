package api

import (
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	stubCommandAddUser      = "command=/comedianadd&text=@test&channel_id=chanid&channel_name=channame"
	stubCommandAddEmptyText = "command=/comedianadd&text="
	stubCommandDelUser      = "command=/comedianremove&text=@test&channel_id=chanid"
	stubCommandListUsers    = "command=/comedianlist"
)

func TestHandleCommands(t *testing.T) {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	rest, err := NewRESTAPI(c)
	if err != nil {
		log.Fatal(err)
	}

	//add user
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "@test added", rec.Body.String())
	}

	//add user empty text
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "username cannot be empty", rec.Body.String())
	}

	//delete user
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "@test deleted", rec.Body.String())
	}

	//list users
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandListUsers))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "[]", rec.Body.String())
	}
}
