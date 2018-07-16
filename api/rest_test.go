package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestHandleCommands(t *testing.T) {

	noneCommand := ""
	emptyCommand := "command=/"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

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
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(emptyCommand))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Equal(t, "Not implemented", rec.Body.String())

}

func TestHandleUserCommands(t *testing.T) {
	AddUser := "command=/comedianadd&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	AddEmptyText := "command=/comedianadd&text="
	AddUserEmptyChannelID := "command=/comedianadd&text=test&channel_id=&channel_name=channame"
	AddUserEmptyChannelName := "command=/comedianadd&text=test&channel_id=chanid&channel_name="
	DelUser := "command=/comedianremove&text=@test&channel_id=chanid"
	ListUsers := "command=/comedianlist&channel_id=chanid"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)
	e := echo.New()

	//add user with no channel id
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddUserEmptyChannelID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//add user with no channel name
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddUserEmptyChannelName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	//add user empty text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//add user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "<@test> added, but there is no standup time for this channel", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Пользователь <@test> добавлен, но в этом канале не установлено время для стэндапов!", rec.Body.String())
	}

	//list users with users
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ListUsers))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)

	if c.Language == "en_US" {
		assert.Equal(t, "Standupers in this channel: <@test>", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Стэндаперы в канале: <@test>", rec.Body.String())
	}

	//add user that already exists
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)

	if c.Language == "en_US" {
		assert.Equal(t, "User already exists!", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Этот пользователь уже существует", rec.Body.String())
	}

	//delete user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)

	if c.Language == "en_US" {
		assert.Equal(t, "<@test> deleted", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Пользователь <@test> удален", rec.Body.String())
	}

	st, err := rest.db.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid",
		Channel:   "channame",
		Time:      int64(12),
	})

	//add user with time set
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "<@test> added", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Пользователь <@test> добавлен", rec.Body.String())
	}

	//delete user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "<@test> deleted", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Пользователь <@test> удален", rec.Body.String())
	}

	//list no users
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ListUsers))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "No standupers in this channel! To add one, please, use /comedianadd slash command", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "В этом канале нет стэндаперов! Чтобы добавить кого-нибдуь, используйте слэш команду `/comedianadd`", rec.Body.String())
	}
	assert.NoError(t, rest.db.DeleteStandupTime(st.ChannelID))

}

func TestHandleTimeCommands(t *testing.T) {

	AddTime := "command=/standuptimeset&text=12:05&channel_id=chanid&channel_name=channame"
	AddTimeEmptyChannelName := "command=/standuptimeset&text=12:05&channel_id=chanid&channel_name="
	AddTimeEmptyChannelID := "command=/standuptimeset&text=12:05&channel_id=&channel_name=channame"
	AddEmptyTime := "command=/standuptimeset&text=&channel_id=chanid&channel_name=channame"
	ListTime := "command=/standuptime&channel_id=chanid"
	ListTimeNoChanID := "command=/standuptime&channel_id="
	DelTime := "command=/standuptimeremove&channel_id=chanid&channel_name=channame"
	DelTimeNoChanID := "command=/standuptimeremove&channel_id=&channel_name=channame"
	DelTimeNoChanName := "command=/standuptimeremove&channel_id=chanid&channel_name="
	currentTime := time.Now()
	timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 12, 5, 0, 0, time.Local).Unix()

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)
	e := echo.New()

	//list time no time added
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ListTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "No standup time set for this channel yet! Please, add a standup time using `/standuptimeset` command!", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "У этого канала до сих пор не установленно стэндап время! Пожалуйста, установите время слэшкомандой `/standuptimeset`!", rec.Body.String())
	}
	//add time (no users)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel>", timeInt), rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Срок для стэндапов установлен на {time}, но в этом канале нет стэндаперов>", timeInt), rec.Body.String())
	}
	//add time no text
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddEmptyTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//add time no channelName
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddTimeEmptyChannelName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	//add time (no channelID)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddTimeEmptyChannelID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//list time with no ChannelID
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ListTimeNoChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//list time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ListTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Standup time is {time}|Standup time set at 12:00>", timeInt), rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Срок для стэндапов установлен на {time}|Срок для стэндапов установлен на 12:00>", timeInt), rec.Body.String())
	}
	//delete time with no channel id
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelTimeNoChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//delete time with no channel name
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelTimeNoChanName))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "`channel_name` cannot be empty", rec.Body.String())

	su1, err := rest.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "chanid",
		Channel:     "channame",
	})
	assert.NoError(t, err)

	//add time (with users)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(AddTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Standup time set at {time}|Standup time set at 12:00>", timeInt), rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, fmt.Sprintf("<!date^%v^Срок для стэндапов установлен на {time}|Срок для стэндапов установлен на 12:00>", timeInt), rec.Body.String())
	}

	//delete time (with users)
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)

	if assert.NoError(t, rest.handleCommands(context)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		if c.Language == "en_US" {
			assert.Equal(t, "standup time for this channel removed, but there are people marked as a standuper.", rec.Body.String())
		}
		if c.Language == "ru_RU" {
			assert.Equal(t, "Время для стэндапов в этом канале удалено, но остались стэндаперы!", rec.Body.String())
		}
	}

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

	//delete time
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(DelTime))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Стэндап время для канала channame удалено", rec.Body.String())
	}
}

func TestHandleReportByProjectCommands(t *testing.T) {
	ReportByProjectEmptyText := "command=/report_by_project&channel_id=chanid&text="
	ReportByProjectEmptyChanID := "command=/report_by_project&channel_id=&text=2018-06-25 2018-06-26"
	ReportByProject := "command=/report_by_project&channel_id=chanid&text= chanid 2018-06-25 2018-06-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	e := echo.New()

	//report by project Empty text
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`text` cannot be empty", rec.Body.String())

	//report by project Empty Chan ID
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectEmptyChanID))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "`channel_id` cannot be empty", rec.Body.String())

	//report by project
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProject))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Full Standup Report chanid:\n\nNo data for this period", rec.Body.String())

}

func TestHandleReportByUserCommands(t *testing.T) {
	ReportByUserEmptyText := "command=/report_by_user&text="
	ReportByUser := "command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= <@userID1|user1> 2018-06-25 2018-06-26"
	ReportByUserMessUser := "command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= <@huiuser|huinya> 2018-06-25 2018-06-26"
	ReportByUserMessDateF := "command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= <@userID1|user1> 2018-6-25 2018-06-26"
	ReportByUserMessDateT := "command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= <@userID1|user1> 2018-06-25 2018-6-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	e := echo.New()
	//report by user empty
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByUserEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
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

	//report by user mess up username
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByUserMessUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "sql: no rows in result set", rec.Body.String())

	//report by user mess up from date
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByUserMessDateF))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "parsing time \"2018-6-25\": month out of range", rec.Body.String())

	//report by user mess up to date
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByUserMessDateT))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "parsing time \"2018-6-26\": month out of range", rec.Body.String())

	//report by user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Full Standup Report for user <@user1>:\n\nNo data for this period", rec.Body.String())

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

}

func TestHandleReportByProjectAndUserCommands(t *testing.T) {
	ReportByProjectAndUserEmptyText := "command=/report_by_project_and_user&channel_id=chanid&text="
	ReportByProjectAndUser := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= channel1 @user1 2018-06-25 2018-06-26"
	ReportByProjectAndUserNameMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= channel1 @nouser 2018-06-25 2018-06-26"
	ReportByProjectAndUserDateToMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= channel1 @user1 2018-6-25 2018-06-26"
	ReportByProjectAndUserDateFromMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= channel1 @user1 2018-06-25 2018-6-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	e := echo.New()
	//report by project and user empty text
	req := httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectAndUserEmptyText))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
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

	//report by project and user user name mess up
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectAndUserNameMessUp))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "sql: no rows in result set", rec.Body.String())

	//report by project and user date from mess up
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectAndUserDateFromMessUp))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "parsing time \"2018-6-26\": month out of range", rec.Body.String())

	//report by project and user date to mess up
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectAndUserDateToMessUp))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "parsing time \"2018-6-25\": month out of range", rec.Body.String())

	//report by project and user
	req = httptest.NewRequest(echo.POST, "/commands", strings.NewReader(ReportByProjectAndUser))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec = httptest.NewRecorder()
	context = e.NewContext(req, rec)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	if c.Language == "en_US" {
		assert.Equal(t, "This user is not set as a standup user in this channel. Please, first add user with `/comdeidanadd` command", rec.Body.String())
	}
	if c.Language == "ru_RU" {
		assert.Equal(t, "Данный пользователь не установлен как стэндапер в этом канале. Для начала добавьте его слэшкомандой `/comdeidanadd`", rec.Body.String())
	}
	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

}
