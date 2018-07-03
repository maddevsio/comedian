package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

var (
	noneCommand  = ""
	emptyCommand = "command=/"

	stubCommandAddUser                 = "command=/comedianadd&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	stubCommandAddEmptyText            = "command=/comedianadd&text="
	stubCommandAddUserEmptyChannelID   = "command=/comedianadd&text=test&channel_id=&channel_name=channame"
	stubCommandAddUserEmptyChannelName = "command=/comedianadd&text=test&channel_id=chanid&channel_name="
	stubCommandDelUser                 = "command=/comedianremove&text=@test&channel_id=chanid"
	stubCommandListUsers               = "command=/comedianlist&channel_id=chanid"

	stubCommandAddTime                 = "command=/standuptimeset&text=12:10&channel_id=chanid&channel_name=channame"
	stubCommandAddTimeEmptyChannelName = "command=/standuptimeset&text=12:05&channel_id=chanid&channel_name="
	stubCommandAddTimeEmptyChannelID   = "command=/standuptimeset&text=12:05&channel_id=&channel_name=channame"
	stubCommandAddEmptyTime            = "command=/standuptimeset&text=&channel_id=chanid&channel_name=channame"

	stubCommandListTime         = "command=/standuptime&channel_id=chanid"
	stubCommandListTimeNoChanID = "command=/standuptime&channel_id="

	stubCommandDelTime           = "command=/standuptimeremove&channel_id=chanid&channel_name=channame"
	stubCommandDelTimeNoChanID   = "command=/standuptimeremove&channel_id=&channel_name=channame"
	stubCommandDelTimeNoChanName = "command=/standuptimeremove&channel_id=chanid&channel_name="

	stubCommandReportByProjectEmptyText   = "command=/report_by_project&channel_id=chanid&text="
	stubCommandReportByProjectEmptyChanID = "command=/report_by_project&channel_id=&text=2018-06-25 2018-06-26"
	stubCommandReportByProject            = "command=/report_by_project&channel_id=chanid&text= chanid 2018-06-25 2018-06-26"

	stubCommandReportByUserEmptyText = "command=/report_by_user&text="
	stubCommandReportByUser          = "command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= <@userID1|user1> 2018-06-25 2018-06-26"

	stubCommandReportByProjectAndUserEmptyText = "command=/report_by_project_and_user&channel_id=chanid&text="
	stubCommandReportByProjectAndUser          = "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= channel1 @user1 2018-06-25 2018-06-26"
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

	//command not allowed
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/command", strings.NewReader(noneCommand))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "\"Command not allowed\"", rec.Body.String())

	//empty command
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(emptyCommand))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Equal(t, "Not implemented", rec.Body.String())

	//add time (no users)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandAddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	//assert.Equal(t, "<!date^1530511800^Standup time at {time} added, but there is no standup users for this channel>", rec.Body.String())

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
		//assert.Equal(t, "<!date^1530511800^Standup time set at {time}|Standup time set at 12:00>", rec.Body.String())
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
	//assert.Equal(t, "<!date^1530511800^Standup time at {time} added, but there is no standup users for this channel>", rec.Body.String())

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
	//assert.Equal(t, "<!date^1530511800^Standup time is {time}|Standup time set at 12:00>", rec.Body.String())

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

	//report by project
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByProject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Full Standup Report chanid:\n\nNo data for this period", rec.Body.String())

	//report by user empty
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByUserEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	su1, err := rest.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)

	//report by user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Full Standup Report for user <@user1>:\n\nNo data for this period", rec.Body.String())

	//report by project and user empty text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByProjectAndUserEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//report by project and user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(stubCommandReportByProjectAndUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "This user is not set as a standup user in this channel. Please, first add user with `/comdeidanadd` command", rec.Body.String())

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

}
