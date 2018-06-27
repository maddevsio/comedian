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
	stubCommandAddUser                    = "command=/comedianadd&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	stubCommandAddUserEmptyChannelID      = "command=/comedianadd&text=test&channel_id=&channel_name=channame"
	stubCommandAddUserEmptyChannelName    = "command=/comedianadd&text=test&channel_id=chanid&channel_name="
	stubCommandAddEmptyText               = "command=/comedianadd&text="
	stubCommandDelUser                    = "command=/comedianremove&text=@test&channel_id=chanid"
	stubCommandListUsers                  = "command=/comedianlist&channel_id=chanid"
	stubCommandAddTime                    = "command=/standuptimeset&text=12:10&channel_id=chanid&channel_name=channame"
	stubCommandAddTimeEmptyChannelName    = "command=/standuptimeset&text=12:05&channel_id=chanid&channel_name="
	stubCommandAddTimeEmptyChannelID      = "command=/standuptimeset&text=12:05&channel_id=&channel_name=channame"
	stubCommandAddEmptyTime               = "command=/standuptimeset&text=&channel_id=chanid&channel_name=channame"
	stubCommandListTime                   = "command=/standuptime&channel_id=chanid"
	stubCommandListTimeNoChanID           = "command=/standuptime&channel_id="
	stubCommandDelTime                    = "command=/standuptimeremove&channel_id=chanid&channel_name=channame"
	stubCommandDelTimeNoChanID            = "command=/standuptimeremove&channel_id=&channel_name=channame"
	stubCommandDelTimeNoChanName          = "command=/standuptimeremove&channel_id=chanid&channel_name="
	stubCommandReportByProjectEmptyText   = "command=/comedian_report_by_project&channel_id=chanid&text="
	stubCommandReportByProjectEmptyChanID = "command=/comedian_report_by_project&channel_id=&text=2018-06-25 2018-06-26"
	stubCommandReportByProject            = "command=/comedian_report_by_project&channel_id=chanid&text="
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

	//add time (no users)
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time at 12:10 (UTC) added, but there is no standup "+
		"users for this channel", rec.Body.String())

	//add time no text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddEmptyTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//add time (no channelName)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTimeEmptyChannelName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	//add time (no channelID)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTimeEmptyChannelID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//delete time with no channel id
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTimeNoChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//delete time with no channel name
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTimeNoChanName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	//delete time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())

	//add user with no channel id
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUserEmptyChannelID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//add user with no channel name
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUserEmptyChannelName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	//add user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<@test> added, but there is no standup time for this channel", rec.Body.String())

	//add time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "standup time at 12:10 (UTC) added", rec.Body.String())
	}

	//add user that already exists
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "User already exists!", rec.Body.String())

	//delete user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<@test> deleted", rec.Body.String())

	//add user with time set
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<@test> added", rec.Body.String())

	//delete time (with users)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "standup time for this channel removed, but there are "+
			"people marked as a standuper.", rec.Body.String())
	}

	//add user empty text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "`text` cannot be empty", rec.Body.String())
	}

	//delete user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<@test> deleted", rec.Body.String())

	//list users
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandListUsers))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "No standupers in this channel! To add one, please, use /comedianadd slash command", rec.Body.String())
	}

	//add time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time at 12:10 (UTC) added, but there is no standup users "+
		"for this channel", rec.Body.String())

	//list time with no ChannelID
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandListTimeNoChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//list time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandListTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time at 12:10 (UTC)", rec.Body.String())

	//delete time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandDelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())

	//report by project Empty text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByProjectEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//report by project Empty Chan ID
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByProjectEmptyChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())
}
