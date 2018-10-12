package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/labstack/echo"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/sirupsen/logrus"
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

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

	for _, tt := range testCases {
		context, rec := getContext(tt.command)
		err := rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("TestHandleCommands: %s failed. Error: %v\n", tt.title, err)
		}
		assert.Equal(t, tt.statusCode, rec.Code)
		assert.Equal(t, tt.responseBody, rec.Body.String())
	}

	assert.NoError(t, rest.db.DeleteUser(admin.ID))

}

func TestHandleUserCommands(t *testing.T) {
	AddUser := "user_id=UB9AE7CL9&command=/comedian_add&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	AddEmptyText := "user_id=UB9AE7CL9&command=/comedian_add&channel_id=chanid&channel_name=channame&text="
	DelUser := "user_id=UB9AE7CL9&command=/comedian_remove&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	ListUsers := "user_id=UB9AE7CL9&command=/comedian_list&channel_id=chanid&channel_name=channame"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
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
		{"empty text", AddEmptyText, http.StatusOK, "Seems like you misspelled username. Please, check and try command again!"},
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
		{"list no users", ListUsers, http.StatusOK, "No standupers in this channel! To add one, please, use `/comedian_add` slash command"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
}

func TestHandleAdminCommands(t *testing.T) {
	AddAdmin := "user_id=UB9AE7CL9&command=/admin_add&text=<@userid|test>&channel_id=chanid&channel_name=channame"
	AddEmptyText := "user_id=UB9AE7CL9&command=/admin_add&text=&channel_id=chanid&channel_name=channame"
	DelAdmin := "user_id=UB9AE7CL9&command=/admin_remove&text=<@userid|test>&&channel_id=chanid&channel_name=channame"
	ListAdmins := "user_id=UB9AE7CL9&command=/admin_list&&channel_id=chanid&channel_name=channame"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})
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
		{"empty text", AddEmptyText, http.StatusOK, "Seems like you misspelled username. Please, check and try command again!"},
		{"add admin no standup time", AddAdmin, http.StatusOK, "<@test> was granted admin access\n"},
		{"list admins", ListAdmins, http.StatusOK, "Admins in this workspace: <@Admin>, <@test>"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
}

func TestHandleTimeCommands(t *testing.T) {

	AddTime := "user_id=UB9AE7CL9&command=/standup_time_set&text=12:05&channel_id=chanid&channel_name=channame"
	AddEmptyTime := "user_id=UB9AE7CL9&command=/standup_time_set&text=&channel_id=chanid&channel_name=channame"
	ListTime := "user_id=UB9AE7CL9&command=/standup_time&channel_id=chanid&channel_name=channame"
	DelTime := "user_id=UB9AE7CL9&command=/standup_time_remove&channel_id=chanid&channel_name=channame"
	currentTime := time.Now()
	timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 12, 5, 0, 0, time.Local).Unix()

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
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
		{"list time no time added", ListTime, http.StatusOK, "No standup time set for this channel yet! Please, add a standup time using `/standup_time_set` command!"},
		{"add time (no users)", AddTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>", timeInt)},
		{"add time no text", AddEmptyTime, http.StatusOK, "Could not understand how you mention time. Please, use 24:00 hour format and try again!"},
		{"list time", ListTime, http.StatusOK, fmt.Sprintf("<!date^%v^Standup time is {time}|Standup time set at 12:00>", timeInt)},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))

}

func TestHandleReportByProjectCommands(t *testing.T) {
	ReportByProjectEmptyText := "user_id=UB9AE7CL9&command=/report_by_project&channel_name=privatechannel&channel_id=chanid&channel_name=channame&text="
	ReportByProject := "user_id=UB9AE7CL9&command=/report_by_project&channel_name=privatechannel&channel_id=chanid&channel_name=channame&text=#chanName 2018-06-25 2018-06-26"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

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
		{"empty text", ReportByProjectEmptyText, http.StatusOK, "Wrong number of arguments"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))

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

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

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

	d := time.Date(2018, 6, 25, 23, 50, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	emptyStandup, err := rest.db.CreateStandup(model.Standup{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
		Comment:   "",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 6, 27, 10, 0, 0, 0, time.Local)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"empty text", ReportByUserEmptyText, http.StatusOK, "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"user mess up", ReportByUserMessUser, http.StatusOK, "User does not exist!"},
		{"date from mess up", ReportByUserMessDateF, http.StatusOK, "parsing time \"2018-6-25\": month out of range"},
		{"date to mess up", ReportByUserMessDateT, http.StatusOK, "parsing time \"2018-6-26\": month out of range"},
		{"correct", ReportByUser, http.StatusOK, "Full Report on user <@userID1> from 2018-06-25 to 2018-06-26:\n\nReport for: 2018-06-25\nIn #chanName <@userID1> did not submit standup!"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
	assert.NoError(t, rest.db.DeleteStandup(emptyStandup.ID))
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

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
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
		{"empty text", ReportByProjectAndUserEmptyText, http.StatusOK, "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"user name mess up", ReportByProjectAndUserNameMessUp, http.StatusOK, "No such user in your slack!"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
}

func TestTimeTableCommand(t *testing.T) {
	AddTimeTable1 := "user_id=UB9AE7CL9&command=/timetable_set&channel_id=123qwe&channel_name=chanName&text=@user1 on mon tue wed at 10:00"
	RemoveTimeTable := "user_id=UB9AE7CL9&command=/timetable_remove&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"
	AddTimeTable2 := "user_id=UB9AE7CL9&command=/timetable_set&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1> on mon tue wed at 10:00"
	ShowTimeTable := "user_id=UB9AE7CL9&command=/timetable_show&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

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
		{"Add Timetable Misspelled username", AddTimeTable1, http.StatusOK, "Seems like you misspelled username. Please, check and try command again!"},
		{"Add Timetable OK", AddTimeTable2, http.StatusOK, "Timetable for <@User1> created: | Monday 10:00 | Tuesday 10:00 | Wednesday 10:00 | \n"},
		{"Show Timetable", ShowTimeTable, http.StatusOK, "Timetable for <@userID1> is: | Monday 10:00 | Tuesday 10:00 | Wednesday 10:00 |\n"},
		{"Remove Timetable", RemoveTimeTable, http.StatusOK, "Timetable removed for <@userID1>\n"},
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
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
}

func TestPMCommand(t *testing.T) {
	AddPM := "user_id=UB9AE7CL9&command=/pm_add&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"
	AddPMNoAccess := "user_id=userID1&command=/pm_add&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"
	ComedianNotInChannel := "user_id=UB9AE7CL9&command=/pm_add&channel_id=123&channel_name=channel&text=<@User1|userID1>"
	NoUsersToAdd := "user_id=UB9AE7CL9&command=/pm_add&channel_id=123qwe&channel_name=channel1&text="
	MisspelledUserName := "user_id=UB9AE7CL9&command=/pm_add&channel_id=123qwe&channel_name=channel1&text=@user1"
	NoChannelIDSet := "user_id=UB9AE7CL9&command=/pm_add&channel_id=&channel_name=channel1&text=<@User1|userID1>"

	c, err := config.Get()
	rest, err := NewRESTAPI(c)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "UB9AE7CL9",
		Role:     "admin",
	})
	assert.NoError(t, err)

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

	testCases := []struct {
		title        string
		command      string
		statusCode   int
		responseBody string
	}{
		{"Add PM", AddPM, http.StatusOK, "<@User1> is assigned as PM in this channel"},
		{"Add PM No Access", AddPMNoAccess, http.StatusOK, "This command is not allowed for you! You are not admin\n"},
		{"No Users To Add", NoUsersToAdd, http.StatusOK, "Seems like you misspelled username. Please, check and try command again!"},
		{"Misspelled Username", MisspelledUserName, http.StatusOK, "Seems like you misspelled username. Please, check and try command again!"},
		{"Comedian Not In Channel", ComedianNotInChannel, http.StatusOK, "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"No Channel ID", NoChannelIDSet, http.StatusOK, "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
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

	assert.NoError(t, rest.db.DeleteChannelMember("User1", "123qwe"))
	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteUser(user.ID))
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
}

func getContext(command string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/command", strings.NewReader(command))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)

	return context, rec
}

func TestPrepareTimetable(t *testing.T) {
	c, err := config.Get()
	r, err := NewRESTAPI(c)
	assert.NoError(t, err)

	user, err := r.db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	m, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	tt, err := r.db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: m.ID,
	})
	assert.NoError(t, err)

	timeNow := time.Date(2018, 10, 7, 10, 0, 0, 0, time.UTC)
	tt.Monday = timeNow.Unix()
	tt.Tuesday = timeNow.Unix()
	tt.Wednesday = timeNow.Unix()
	tt.Thursday = timeNow.Unix()
	tt.Friday = timeNow.Unix()

	tt, err = r.db.UpdateTimeTable(tt)

	assert.NoError(t, err)
	assert.Equal(t, timeNow.Unix(), tt.Monday)

	timeUpdate := time.Date(2018, 10, 7, 12, 0, 0, 0, time.UTC).Unix()

	tt, err = r.prepareTimeTable(tt, "mon tue wed thu fri sat sun", timeUpdate)
	assert.NoError(t, err)
	assert.Equal(t, timeUpdate, tt.Monday)
	assert.NoError(t, r.db.DeleteChannelMember(m.UserID, m.ChannelID))
	assert.NoError(t, r.db.DeleteUser(user.ID))
	assert.NoError(t, r.db.DeleteTimeTable(tt.ID))

}

func TestUserHasAccess(t *testing.T) {
	c, err := config.Get()
	r, err := NewRESTAPI(c)
	assert.NoError(t, err)

	userHasAccess := r.userHasAccess("RANDOMID", "RANDOMCHAN")
	assert.Equal(t, false, userHasAccess)

	admin, err := r.db.CreateUser(model.User{
		UserID:   "ADMINID",
		UserName: "Admin",
		Role:     "admin",
	})

	userHasAccess = r.userHasAccess(admin.UserID, "RANDOMCHAN")
	assert.Equal(t, true, userHasAccess)

	pmUser, err := r.db.CreateUser(model.User{
		UserID:   "PMID",
		UserName: "futurePM",
		Role:     "",
	})

	pm, err := r.db.CreatePM(model.ChannelMember{
		UserID:    pmUser.UserID,
		ChannelID: "RANDOMCHAN",
	})

	userHasAccess = r.userHasAccess(pm.UserID, "RANDOMCHAN")
	assert.Equal(t, true, userHasAccess)

	user, err := r.db.CreateUser(model.User{
		UserID:   "USERID",
		UserName: "User",
		Role:     "",
	})

	userHasAccess = r.userHasAccess(user.UserID, "RANDOMCHAN")
	assert.Equal(t, false, userHasAccess)

	assert.NoError(t, r.db.DeleteUser(admin.ID))
	assert.NoError(t, r.db.DeleteUser(pmUser.ID))
	assert.NoError(t, r.db.DeleteUser(user.ID))
	assert.NoError(t, r.db.DeleteChannelMember(pmUser.UserID, "RANDOMCHAN"))

}
