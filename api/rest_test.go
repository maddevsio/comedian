package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestHandleCommands(t *testing.T) {

	noneCommand := ""
	emptyCommand := "command=/"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"command not allowed", noneCommand, http.StatusMethodNotAllowed, "\"Command not allowed\""},
		{"empty command", emptyCommand, http.StatusNotImplemented, "Not implemented"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}
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

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty channel ID", AddUserEmptyChannelID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"empty channel name", AddUserEmptyChannelName, http.StatusBadRequest, "`channel_name` cannot be empty"},
		{"empty text", AddEmptyText, http.StatusBadRequest, "`text` cannot be empty"},
		{"add user no standup time", AddUser, http.StatusOK, "<@test> added, but there is no standup time for this channel"},
		{"list users", ListUsers, http.StatusOK, "Standupers in this channel: <@test>"},
		{"add user with user exist", AddUser, http.StatusOK, "User already exists!"},
		{"delete user", DelUser, http.StatusOK, "<@test> deleted"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleUserCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	st, err := rest.db.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid",
		Channel:   "channame",
		Time:      int64(12),
	})

	testCases = []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"add user with standup time", AddUser, http.StatusOK, "<@test> added"},
		{"delete user", DelUser, http.StatusOK, "<@test> deleted"},
		{"list no users", ListUsers, http.StatusOK, "No standupers in this channel! To add one, please, use /comedianadd slash command"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleUserCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
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

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"list time no time added", ListTime, http.StatusOK, "No standup time set for this channel yet! Please, add a standup time using `/standuptimeset` command!"},
		{"add time (no users)", AddTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel>", timeInt)},
		{"add time no text", AddEmptyTime, http.StatusBadRequest, "`text` cannot be empty"},
		{"add time no channelName", AddTimeEmptyChannelName, http.StatusBadRequest, "`channel_name` cannot be empty"},
		{"add time (no channelID)", AddTimeEmptyChannelID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"list time no chan ID", ListTimeNoChanID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"list time", ListTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time is {time}|Standup time set at 12:00>", timeInt)},
		{"del time no chan ID", DelTimeNoChanID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"del time no chan Name", DelTimeNoChanName, http.StatusBadRequest, "`channel_name` cannot be empty"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleTimeCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	su1, err := rest.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "chanid",
		Channel:     "channame",
	})
	assert.NoError(t, err)

	testCases = []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"list time no time added", AddTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time set at {time}|Standup time set at 12:00>", timeInt)},
		{"remove time no text", DelTime, http.StatusOK, "standup time for this channel removed, but there are people marked as a standuper."},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleTimeCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

	//delete time
	context, rec := getContext(DelTime)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())

}

func TestHandleReportByProjectCommands(t *testing.T) {
	ReportByProjectEmptyText := "command=/report_by_project&channel_id=<#CBA2M41Q8|chanid>&text="
	ReportByProjectEmptyChanID := "command=/report_by_project&channel_id=&text=2018-06-25 2018-06-26"
	ReportByProject := "command=/report_by_project&channel_id=chanid&text= <#CBA2M41Q8|chanid> 2018-06-25 2018-06-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/projects/chanid/2018-06-25/2018-06-26", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0}]`))

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByProjectEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"empty channel ID", ReportByProjectEmptyChanID, http.StatusOK, "`channel_id` cannot be empty"},
		{"correct", ReportByProject, http.StatusOK, "Full Standup Report by project <#CBA2M41Q8>:\n\nNo data for this period\n\nCommits for period: 0 \nMerges for period: 0\n"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("ReportByProject: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

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

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/users/userID1/2018-06-25/2018-06-26", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0, "worklogs": 0}]`))

	su1, err := rest.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByUserEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"user mess up", ReportByUserMessUser, http.StatusOK, "sql: no rows in result set"},
		{"date from mess up", ReportByUserMessDateF, http.StatusOK, "parsing time \"2018-6-25\": month out of range"},
		{"date to mess up", ReportByUserMessDateT, http.StatusOK, "parsing time \"2018-6-26\": month out of range"},
		{"correct", ReportByUser, http.StatusOK, "Full Standup Report for user <@user1>:\n\nNo data for this period\n\nCommits for period: 0 \nMerges for period: 0\nWorklogs: 0 hours"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("ReportByUser: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))

}

func TestHandleReportByProjectAndUserCommands(t *testing.T) {
	ReportByProjectAndUserEmptyText := "command=/report_by_project_and_user&channel_id=<#CBA2M41Q8|chanid>&text="
	ReportByProjectAndUser := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= <#CBA2M41Q8|chanid> <@USERID|user1> 2018-06-25 2018-06-26"
	ReportByProjectAndUserNameMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= <#CBA2M41Q8|chanid> <@USERID|nouser> 2018-06-25 2018-06-26"
	ReportByProjectAndUserDateToMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= <#CBA2M41Q8|chanid> <@USERID|user1> 2018-6-25 2018-06-26"
	ReportByProjectAndUserDateFromMessUp := "command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= <#CBA2M41Q8|chanid> <@USERID|user1> 2018-06-25 2018-6-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/projects-users/chanid/USERID/2018-06-25/2018-06-26", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0}]`))

	su1, err := rest.db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByProjectAndUserEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"user name mess up", ReportByProjectAndUserNameMessUp, http.StatusOK, "This user is not set as a standup user in this channel. Please, first add user with `/comdeidanadd` command"},
		{"date from mess up", ReportByProjectAndUserDateFromMessUp, http.StatusOK, "parsing time \"2018-6-26\": month out of range"},
		{"date to mess up", ReportByProjectAndUserDateToMessUp, http.StatusOK, "parsing time \"2018-6-25\": month out of range"},
		{"correct", ReportByProjectAndUser, http.StatusOK, "This user is not set as a standup user in this channel. Please, first add user with `/comdeidanadd` command"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("ReportByProjectAndUser: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	assert.NoError(t, rest.db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))
}

func getContext(command string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/command", strings.NewReader(command))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)

	return context, rec
}
