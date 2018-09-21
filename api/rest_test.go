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

	noneCommand := "user_id=UB9AE7CL9"
	emptyCommand := "user_id=UB9AE7CL9&command=/"

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
	AddUser := "user_id=UB9AE7CL9&command=/comedianadd&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	AddEmptyText := "user_id=UB9AE7CL9&command=/comedianadd&text="
	AddUserEmptyChannelID := "user_id=UB9AE7CL9&command=/comedianadd&text=test&channel_id=&channel_name=channame"
	AddUserEmptyChannelName := "user_id=UB9AE7CL9&command=/comedianadd&text=test&channel_id=chanid&channel_name="
	DelUser := "user_id=UB9AE7CL9&command=/comedianremove&text=<@userid|test>&channel_id=chanid"
	ListUsers := "user_id=UB9AE7CL9&command=/comedianlist&channel_id=chanid"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/groups.info",
		httpmock.NewStringResponder(200, `{"ok": true,"group": {"id": "chanid", "name": "channame"}}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/channels.info",
		httpmock.NewStringResponder(200, `{"ok": true,"channel": {"id": "chanid", "name": "channame"}}`))

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty channel ID", AddUserEmptyChannelID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"empty channel name", AddUserEmptyChannelName, http.StatusBadRequest, "`channel_name` cannot be empty"},
		{"empty text", AddEmptyText, http.StatusBadRequest, "`text` cannot be empty"},
		{"add user no standup time", AddUser, http.StatusOK, "<@test> now submits standups in this channel, but there is no standup time set yet!\n"},
		{"list users", ListUsers, http.StatusOK, "Standupers in this channel: <@userid>"},
		{"add user with user exist", AddUser, http.StatusOK, "<@userid> already assigned as standuper in this channel\n"},
		{"delete user", DelUser, http.StatusOK, "<@test> no longer submits standups in this channel\n"},
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
	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
	err = rest.db.CreateStandupTime(int64(12), channel.ChannelID)
	assert.NoError(t, err)

	testCases = []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"add user with standup time", AddUser, http.StatusOK, "<@test> assigned to submit standups in this channel\n"},
		{"delete user", DelUser, http.StatusOK, "<@test> no longer submits standups in this channel\n"},
		{"list no users", ListUsers, http.StatusOK, "No standupers in this channel! To add one, please, use `/comedianadd` slash command"},
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

	assert.NoError(t, rest.db.DeleteStandupTime("chanid"))
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
}

func TestHandleAdminCommands(t *testing.T) {
	AddAdmin := "user_id=UB9AE7CL9&command=/adminadd&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	AddEmptyText := "user_id=UB9AE7CL9&command=/adminadd&text="
	AddAdminEmptyChannelID := "user_id=UB9AE7CL9&command=/adminadd&text=test&channel_id=&channel_name=channame"
	AddAdminEmptyChannelName := "user_id=UB9AE7CL9&command=/adminadd&text=test&channel_id=chanid&channel_name="
	DelAdmin := "user_id=UB9AE7CL9&command=/adminremove&text=<@userid|test>&channel_id=chanid"
	ListAdmins := "user_id=UB9AE7CL9&command=/adminlist&channel_id=chanid"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://slack.com/api/groups.info",
		httpmock.NewStringResponder(200, `{"ok": true,"group": {"id": "chanid", "name": "channame"}}`))
	httpmock.RegisterResponder("POST", "https://slack.com/api/channels.info",
		httpmock.NewStringResponder(200, `{"ok": true,"channel": {"id": "chanid", "name": "channame"}}`))

	u1, err := rest.db.CreateUser(model.User{
		UserName: "test",
		UserID:   "userid",
		Role:     "",
	})
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty channel ID", AddAdminEmptyChannelID, http.StatusBadRequest, "`channel_id` cannot be empty"},
		{"empty channel name", AddAdminEmptyChannelName, http.StatusBadRequest, "`channel_name` cannot be empty"},
		{"empty text", AddEmptyText, http.StatusBadRequest, "`text` cannot be empty"},
		{"add admin no standup time", AddAdmin, http.StatusOK, "<@test> was granted admin access\n"},
		{"list admins", ListAdmins, http.StatusOK, "Admins in this workspace: <@test>"},
		{"add admin with admin exist", AddAdmin, http.StatusOK, "User is already admin!"},
		{"delete admin", DelAdmin, http.StatusOK, "<@test> was removed as admin\n"},
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

	err = rest.db.CreateStandupTime(int64(12), "chanid")
	assert.NoError(t, err)

	testCases = []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"add admin with standup time", AddAdmin, http.StatusOK, "<@test> was granted admin access\n"},
		{"delete admin", DelAdmin, http.StatusOK, "<@test> was removed as admin\n"},
		{"list no admins", ListAdmins, http.StatusOK, "No admins in this channel! To add one, please, use `/adminadd` slash command"},
	}

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleAdminCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	assert.NoError(t, rest.db.DeleteStandupTime("chanid"))
	assert.NoError(t, rest.db.DeleteUser(u1.ID))
}

func TestHandleTimeCommands(t *testing.T) {

	AddTime := "user_id=UB9AE7CL9&command=/standuptimeset&text=12:05&channel_id=chanid&channel_name=channame"
	AddTimeEmptyChannelName := "user_id=UB9AE7CL9&command=/standuptimeset&text=12:05&channel_id=chanid&channel_name="
	AddTimeEmptyChannelID := "user_id=UB9AE7CL9&command=/standuptimeset&text=12:05&channel_id=&channel_name=channame"
	AddEmptyTime := "user_id=UB9AE7CL9&command=/standuptimeset&text=&channel_id=chanid&channel_name=channame"
	ListTime := "user_id=UB9AE7CL9&command=/standuptime&channel_id=chanid"
	ListTimeNoChanID := "user_id=UB9AE7CL9&command=/standuptime&channel_id="
	DelTime := "user_id=UB9AE7CL9&command=/standuptimeremove&channel_id=chanid&channel_name=channame"
	DelTimeNoChanID := "user_id=UB9AE7CL9&command=/standuptimeremove&channel_id=&channel_name=channame"
	DelTimeNoChanName := "user_id=UB9AE7CL9&command=/standuptimeremove&channel_id=chanid&channel_name="
	currentTime := time.Now()
	timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 12, 5, 0, 0, time.Local).Unix()

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"list time no time added", ListTime, http.StatusOK, "No standup time set for this channel yet! Please, add a standup time using `/standuptimeset` command!"},
		{"add time (no users)", AddTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>", timeInt)},
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

	su1, err := rest.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: "chanid",
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

	assert.NoError(t, rest.db.DeleteChannelMember(su1.UserID, su1.ChannelID))

	//delete time
	context, rec := getContext(DelTime)
	assert.NoError(t, rest.handleCommands(context))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "standup time for channame channel deleted", rec.Body.String())
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))

}

func TestHandleReportByProjectCommands(t *testing.T) {
	ReportByProjectEmptyText := "user_id=UB9AE7CL9&command=/report_by_project&channel_id=<#CBA2M41Q8|chanid>&text="
	ReportByProjectEmptyChanID := "user_id=UB9AE7CL9&command=/report_by_project&channel_id=&text=2018-06-25 2018-06-26"
	ReportByProject := "user_id=UB9AE7CL9&command=/report_by_project&channel_id=chanid&text=#chanName 2018-06-25 2018-06-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/projects/chanName/2018-06-25/2018-06-26/", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0}]`))

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByProjectEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"empty channel ID", ReportByProjectEmptyChanID, http.StatusOK, "`channel_id` cannot be empty"},
		{"correct", ReportByProject, http.StatusOK, "Full Report on project #chanName from 2018-06-25 to 2018-06-26:\n\nNo standup data for this period\n"},
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

	assert.NoError(t, rest.db.DeleteChannel(channel.ID))

}

func TestHandleReportByUserCommands(t *testing.T) {
	ReportByUserEmptyText := "user_id=UB9AE7CL9&command=/report_by_user&text="
	ReportByUser := "user_id=UB9AE7CL9&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-06-25 2018-06-26"
	ReportByUserMessUser := "user_id=UB9AE7CL9&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @huinya 2018-06-25 2018-06-26"
	ReportByUserMessDateF := "user_id=UB9AE7CL9&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-6-25 2018-06-26"
	ReportByUserMessDateT := "user_id=UB9AE7CL9&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-06-25 2018-6-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/users/userID1/2018-06-25/2018-06-26/", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0, "worklogs": 0}]`))

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "123qwe",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	user, err := rest.db.CreateUser(model.User{
		UserName: "User1",
		UserID:   "userID1",
		Role:     "",
	})
	assert.NoError(t, err)

	su1, err := rest.db.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByUserEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"user mess up", ReportByUserMessUser, http.StatusOK, "User does not exist!"},
		{"date from mess up", ReportByUserMessDateF, http.StatusOK, "parsing time \"2018-6-25\": month out of range"},
		{"date to mess up", ReportByUserMessDateT, http.StatusOK, "parsing time \"2018-6-26\": month out of range"},
		{"correct", ReportByUser, http.StatusOK, "Full Report on user <@userID1> from 2018-06-25 to 2018-06-26:\n\nNo standup data for this period\n"},
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

	assert.NoError(t, rest.db.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteUser(user.ID))

}

func TestHandleReportByProjectAndUserCommands(t *testing.T) {
	ReportByProjectAndUserEmptyText := "user_id=UB9AE7CL9&command=/report_by_project_and_user&channel_id=#chanid&text="
	ReportByProjectAndUser := "user_id=UB9AE7CL9&command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-06-25 2018-06-26"
	ReportByProjectAndUserNameMessUp := "user_id=UB9AE7CL9&command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= #chanid @nouser 2018-06-25 2018-06-26"
	ReportByProjectAndUserDateToMessUp := "user_id=UB9AE7CL9&command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-6-25 2018-06-26"
	ReportByProjectAndUserDateFromMessUp := "user_id=UB9AE7CL9&command=/report_by_project_and_user&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-06-25 2018-6-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", fmt.Sprintf("%v/rest/api/v1/logger/user-in-project/userID1/chanid/2018-06-25/2018-06-26/", c.CollectorURL),
		httpmock.NewStringResponder(200, `[{"total_commits": 0, "total_merges": 0}]`))

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "chanid",
		ChannelID:   "123qwe",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	user, err := rest.db.CreateUser(model.User{
		UserName: "user1",
		UserID:   "userID1",
		Role:     "",
	})
	assert.NoError(t, err)

	su1, err := rest.db.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByProjectAndUserEmptyText, http.StatusOK, "`text` cannot be empty"},
		{"user name mess up", ReportByProjectAndUserNameMessUp, http.StatusOK, "User does not exist!"},
		{"date from mess up", ReportByProjectAndUserDateFromMessUp, http.StatusOK, "parsing time \"2018-6-26\": month out of range"},
		{"date to mess up", ReportByProjectAndUserDateToMessUp, http.StatusOK, "parsing time \"2018-6-25\": month out of range"},
		{"correct", ReportByProjectAndUser, http.StatusOK, "Report on user <@userID1> in project #chanid from 2018-06-25 to 2018-06-26\n\nNo standup data for this period\n"},
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

	assert.NoError(t, rest.db.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteUser(user.ID))
}

func getContext(command string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/command", strings.NewReader(command))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)

	return context, rec
}
