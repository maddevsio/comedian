package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/stretchr/testify/assert"
)

var (
	stubCommandAddUser      = "command=/comedianadd&text=@test&channel_id=chanid&channel_name=channame"
	stubCommandAddEmptyText = "command=/comedianadd&text="
	stubCommandDelUser      = "command=/comedianremove&text=@test&channel_id=chanid"
	stubCommandListUsers    = "command=/comedianlist&channel_id=chanid"
	stubCommandAddTime      = "command=/standuptimeset&text=12:10&channel_id=chanid&channel_name=channame"
	stubCommandAddEmptyTime = "command=/standuptimeset&text=&channel_id=chanid&channel_name=channame"
	stubCommandListTime     = "command=/standuptime&channel_id=chanid"
	stubCommandDelTime      = "command=/standuptimeremove&channel_id=chanid&channel_name=channame"
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
		assert.Equal(t, "`text` cannot be empty", rec.Body.String())
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

	//add time
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "standup time at 12:10 (UTC) added", rec.Body.String())
	}

	//add time empty text
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddEmptyTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "`text` cannot be empty", rec.Body.String())
	}

	//list time
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandListTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "standup time at 12:10 (UTC)", rec.Body.String())
	}

	//delete time
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())
	}
}
