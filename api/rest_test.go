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
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestHandleCommands(t *testing.T) {

	c, err := config.Get()
	assert.NoError(t, err)
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})

	admin, err := rest.db.CreateUser(model.User{
		UserName: "adminUser",
		UserID:   "SuperAdminID",
		Role:     "user",
	})
	assert.NoError(t, err)

	user1, err := rest.db.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "userID",
		Role:     "user",
	})
	assert.NoError(t, err)

	user2, err := rest.db.CreateUser(model.User{
		UserName: "userName",
		UserID:   "userID2",
		Role:     "user",
	})
	assert.NoError(t, err)

	testCases := []struct {
		senderID     string
		channelID    string
		channelTitle string
		command      string
		text         string
		response     string
	}{
		{"SuperAdminID", "TestChannelID", "TestChannel", "", "", "Not implemented"},

		{"", "TestChannelID", "TestChannel", "add", "<@userID1|userName>", "Something went wrong. Please, try again later or report the problem to chatbot support!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@userID1|userName>", "Users are assigned as developers: <@userID1|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@userID2|userName> / admin", "Users are assigned as admins: <@userID2|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@userID2|userName> / wrongUserRole", "Please, check correct role name (admin, developer, pm)"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@userID3|userName> / developer", "Users are assigned as developers: <@userID3|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@userID4|userName> / pm", "Users are assigned as PMs: <@userID4|userName>\n"},

		{"userID", "TestChannelID", "TestChannel", "add", "<@userID1|userName>", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "add", "<@userID2|userName> / admin", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "add", "<@userID3|userName> / developer", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "add", "<@userID4|userName> / pm", "Access Denied! You need to be at least admin in this slack to use this command!"},

		{"SuperAdminID", "", "", "delete", "<@userID1|userName>", "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@userID1|userName>", "The following users were removed as developers: <@userID1|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@userID2|userName> / admin", "Users are removed as admins: <@userID2|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@userID2|userName> / wrongUserRole", "Please, check correct role name (admin, developer, pm)"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@userID3|userName> / developer", "The following users were removed as developers: <@userID3|userName>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@userID4|userName> / pm", "Users are removed as PMs: <@userID4|userName>\n"},

		{"userID", "TestChannelID", "TestChannel", "delete", "<@userID1|userName>", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "delete", "<@userID2|userName> / admin", "Access Denied! You need to be at least admin in this slack to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "delete", "<@userID3|userName> / developer", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"userID", "TestChannelID", "TestChannel", "delete", "<@userID4|userName> / pm", "Access Denied! You need to be at least admin in this slack to use this command!"},

		{"SuperAdminID", "WrongChannelID", "TestChannel", "list", "", "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "list", "", "No standupers in this channel! To add one, please, use `/add` slash command"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "list", "developer", "No standupers in this channel! To add one, please, use `/add` slash command"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "list", "pm", "No PMs in this channel! To add one, please, use `/add` slash command"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "list", "admin", "No admins in this workspace! To add one, please, use `/add` slash command"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "list", "wrongRole", "Please, check correct role name (admin, developer, pm)"},
	}

	for _, tt := range testCases {
		request := fmt.Sprintf("user_id=%s&channel_id=%s&channel_name=%s&command=/%s&text=%s",
			tt.senderID,
			tt.channelID,
			tt.channelTitle,
			tt.command,
			tt.text,
		)

		context, response := getContext(request)
		err = rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("handleCommands failed: %v", err)
		}
		assert.Equal(t, tt.response, response.Body.String())
	}

	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteUser(admin.ID))
	assert.NoError(t, rest.db.DeleteUser(user1.ID))
	assert.NoError(t, rest.db.DeleteUser(user2.ID))

}

func TestUsers(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})

	testCases := []struct {
		function string
		users    []string
		channel  string
		output   string
	}{
		{"list", []string{}, "TestChannelID", "No standupers in this channel! To add one, please, use `/add` slash command"},
		{"add", []string{"<@userID1|userName1>", "<@userID2|userName1>", "<@userID3|userName1>"}, "TestChannelID", "Users are assigned as developers: <@userID1|userName1><@userID2|userName1><@userID3|userName1>\n"},
		{"add", []string{"<@userID1|userName1>", "@randomUser"}, "TestChannelID", "Could not assign users as developers: @randomUser\nUsers already have roles: <@userID1|userName1>\n"},
		{"list", []string{}, "TestChannelID", "Standupers in this channel: <@userID1>, <@userID2>, <@userID3>"},
		{"delete", []string{"<@userIDwrong|userName1>", "@doesNotMatchUser"}, "TestChannelID", "Could not remove the following users as developers: <@userIDwrong|userName1>@doesNotMatchUser\n"},
		{"delete", []string{"<@userID1|userName1>", "<@userID2|userName1>", "<@userID3|userName1>"}, "TestChannelID", "The following users were removed as developers: <@userID1|userName1><@userID2|userName1><@userID3|userName1>\n"},
	}

	for _, tt := range testCases {
		var output string
		switch tt.function {
		case "add":
			output = rest.addUsers(tt.users, tt.channel)
		case "list":
			output = rest.listUsers(tt.channel)
		case "delete":
			output = rest.deleteUsers(tt.users, tt.channel)
		}
		assert.Equal(t, tt.output, output)
	}

	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
}

func TestPMs(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})

	chm, err := rest.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID3",
		ChannelID: "TestChannelID",
	})
	assert.NoError(t, err)

	testCases := []struct {
		function string
		users    []string
		channel  string
		output   string
	}{
		{"list", []string{}, "TestChannelID", "No PMs in this channel! To add one, please, use `/add` slash command"},
		{"add", []string{"<@userID1|userName1>", "<@userID2|userName1>"}, "TestChannelID", "Users are assigned as PMs: <@userID1|userName1><@userID2|userName1>\n"},
		{"add", []string{"<@userID1|userName1>", "@randomUser"}, "TestChannelID", "Could not assign users as PMs: @randomUser\nUsers already have roles: <@userID1|userName1>\n"},
		{"list", []string{}, "TestChannelID", "PMs in this channel: <@userID1>, <@userID2>"},
		{"delete", []string{"<@userIDwrong|userName1>", "@doesNotMatchUser"}, "TestChannelID", "Could not remove users as PMs: <@userIDwrong|userName1>@doesNotMatchUser\n"},
		{"delete", []string{"<@userID1|userName1>", "<@userID2|userName1>", "<@userID3|userName1>"}, "TestChannelID", "Could not remove users as PMs: <@userID3|userName1>\nUsers are removed as PMs: <@userID1|userName1><@userID2|userName1>\n"},
	}

	for _, tt := range testCases {
		var output string
		switch tt.function {
		case "add":
			output = rest.addPMs(tt.users, tt.channel)
		case "list":
			output = rest.listPMs(tt.channel)
		case "delete":
			output = rest.deletePMs(tt.users, tt.channel)
		}
		assert.Equal(t, tt.output, output)
	}

	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteChannelMember(chm.UserID, chm.ChannelID))

}

func TestAdmins(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	user1, err := rest.db.CreateUser(model.User{
		UserName: "userName1",
		UserID:   "userID1",
		Role:     "user",
	})
	assert.NoError(t, err)

	user2, err := rest.db.CreateUser(model.User{
		UserName: "userName2",
		UserID:   "userID2",
		Role:     "user",
	})
	assert.NoError(t, err)

	user3, err := rest.db.CreateUser(model.User{
		UserName: "userName3",
		UserID:   "userID3",
		Role:     "user",
	})
	assert.NoError(t, err)

	testCases := []struct {
		function string
		users    []string
		output   string
	}{
		{"list", []string{}, "No admins in this workspace! To add one, please, use `/add` slash command"},
		{"add", []string{"<@userID1|userName1>", "<@userID2|userName1>"}, "Users are assigned as admins: <@userID1|userName1><@userID2|userName1>\n"},
		{"add", []string{"<@userID1|userName1>", "@randomUser"}, "Could not assign users as admins: @randomUser\nUsers were already assigned as admins: <@userID1|userName1>\n"},
		{"list", []string{}, "Admins in this workspace: <@userName1>, <@userName2>"},
		{"delete", []string{"<@userIDwrong|userName1>", "@doesNotMatchUser"}, "Could not remove users as admins: <@userIDwrong|userName1>@doesNotMatchUser\n"},
		{"delete", []string{"<@userID1|userName1>", "<@userID2|userName1>", "<@userID3|userName1>"}, "Could not remove users as admins: <@userID3|userName1>\nUsers are removed as admins: <@userID1|userName1><@userID2|userName1>\n"},
	}

	for _, tt := range testCases {
		var output string
		switch tt.function {
		case "add":
			output = rest.addAdmins(tt.users)
		case "list":
			output = rest.listAdmins()
		case "delete":
			output = rest.deleteAdmins(tt.users)
		}
		assert.Equal(t, tt.output, output)
	}

	assert.NoError(t, rest.db.DeleteUser(user1.ID))

	assert.NoError(t, rest.db.DeleteUser(user2.ID))

	assert.NoError(t, rest.db.DeleteUser(user3.ID))

}

func TestHandleTimeCommands(t *testing.T) {

	// currentTime := time.Now()
	// timeInt := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 12, 5, 0, 0, time.Local).Unix()

	c, err := config.Get()
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "testUser",
		UserID:   "testID",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "TestChannel",
		ChannelID:   "TestChannelID",
		StandupTime: int64(0),
	})

	testCases := []struct {
		senderID     string
		channelID    string
		channelTitle string
		command      string
		text         string
		response     string
	}{
		{"testID", "TestChannelID", "TestChannel", "standup_time", "", "No standup time set for this channel yet! Please, add a standup time using `/standup_time_set` command!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time", "", "No standup time set for this channel yet! Please, add a standup time using `/standup_time_set` command!"},
		{"SuperAdminID", "wrongchannel", "xyz", "standup_time", "", "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"testID", "TestChannelID", "TestChannel", "standup_time_set", "12:05", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@testID|testUser> / pm", "Users are assigned as PMs: <@testID|testUser>\n"},
		{"testID", "TestChannelID", "TestChannel", "standup_time_set", "12:05", "<!date^1540361100^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@testID|testUser> / pm", "Users are removed as PMs: <@testID|testUser>\n"},
		{"SuperAdminID", "wrongchannel", "xyz", "standup_time_set", "12:05", "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time_set", "1205", "Could not understand how you mention time. Please, use 24:00 hour format and try again!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time_set", "12:05", "<!date^1540361100^Standup time at {time} added, but there is no standup users for this channel|Standup time at 12:00 added, but there is no standup users for this channel>"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time", "", "<!date^1540361100^Standup time is {time}|Standup time set at 12:00>"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time_remove", "", "standup time for TestChannel channel deleted"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "add", "<@testID|testUser> / developer", "Users are assigned as developers: <@testID|testUser>\n"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time_set", "12:05", "<!date^1540361100^Standup time set at {time}|Standup time set at 12:00>"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time", "", "<!date^1540361100^Standup time is {time}|Standup time set at 12:00>"},
		{"SuperAdminID", "wrongchannel", "xyz", "standup_time_remove", "", "I do not have this channel in my database... Please, reinvite me if I am already here and try again!"},
		{"testID", "TestChannelID", "TestChannel", "standup_time_remove", "", "Access Denied! You need to be at least PM in this project to use this command!"},
		{"SuperAdminID", "TestChannelID", "TestChannel", "standup_time_remove", "", "standup time for this channel removed, but there are people marked as a standuper."},
		{"SuperAdminID", "TestChannelID", "TestChannel", "delete", "<@testID|testUser> / developer", "The following users were removed as developers: <@testID|testUser>\n"},
	}

	for _, tt := range testCases {
		request := fmt.Sprintf("user_id=%s&channel_id=%s&channel_name=%s&command=/%s&text=%s",
			tt.senderID,
			tt.channelID,
			tt.channelTitle,
			tt.command,
			tt.text,
		)

		context, response := getContext(request)
		err = rest.handleCommands(context)
		if err != nil {
			logrus.Errorf("handleCommands failed: %v", err)
		}
		assert.Equal(t, tt.response, response.Body.String())
	}

	assert.NoError(t, rest.db.DeleteChannel(channel.ID))
	assert.NoError(t, rest.db.DeleteUser(admin.ID))

}

func TestHandleReportByProjectCommands(t *testing.T) {
	ReportByProjectEmptyText := "user_id=SuperAdminID&command=/report_by_project&channel_name=privatechannel&channel_id=chanid&channel_name=channame&text="
	ReportByProject := "user_id=SuperAdminID&command=/report_by_project&channel_name=privatechannel&channel_id=chanid&channel_name=channame&text=#chanName 2018-06-25 2018-06-26"

	c, err := config.Get()
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	channel, err := rest.db.CreateChannel(model.Channel{
		ChannelName: "chanName",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "SuperAdminID",
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
	ReportByUserEmptyText := "user_id=SuperAdminID&command=/report_by_user&text="
	ReportByUser := "user_id=SuperAdminID&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-06-25 2018-06-26"
	ReportByUserMessUser := "user_id=SuperAdminID&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @huinya 2018-06-25 2018-06-26"
	ReportByUserMessDateF := "user_id=SuperAdminID&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-6-25 2018-06-26"
	ReportByUserMessDateT := "user_id=SuperAdminID&command=/report_by_user&channel_id=123qwe&channel_name=channel1&text= @user1 2018-06-25 2018-6-26"

	c, err := config.Get()
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "SuperAdminID",
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
		{"correct", ReportByUser, http.StatusOK, "Full Report on user <@userID1> from 2018-06-25 to 2018-06-26:\n\nReport for: 2018-06-25\nIn #chanName <@userID1> did not submit standup!================================================\n"},
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
	ReportByProjectAndUserEmptyText := "user_id=SuperAdminID&command=/report_by_user_in_project&channel_id=#chanid&text="
	ReportByProjectAndUser := "user_id=SuperAdminID&command=/report_by_user_in_project&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-06-25 2018-06-26"
	ReportByProjectAndUserNameMessUp := "user_id=SuperAdminID&command=/report_by_user_in_project&channel_id=123qwe&channel_name=channel1&text= #chanid @nouser 2018-06-25 2018-06-26"
	ReportByProjectAndUserDateToMessUp := "user_id=SuperAdminID&command=/report_by_user_in_project&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-6-25 2018-06-26"
	ReportByProjectAndUserDateFromMessUp := "user_id=SuperAdminID&command=/report_by_user_in_project&channel_id=123qwe&channel_name=channel1&text= #chanid @user1 2018-06-25 2018-6-26"

	c, err := config.Get()
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "SuperAdminID",
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
	AddTimeTable1 := "user_id=SuperAdminID&command=/timetable_set&channel_id=123qwe&channel_name=chanName&text=@user1 on mon tue wed at 10:00"
	RemoveTimeTable := "user_id=SuperAdminID&command=/timetable_remove&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"
	AddTimeTable2 := "user_id=SuperAdminID&command=/timetable_set&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1> on mon tue wed at 10:00"
	ShowTimeTable := "user_id=SuperAdminID&command=/timetable_show&channel_id=123qwe&channel_name=chanName&text=<@User1|userID1>"

	c, err := config.Get()
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	rest, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	admin, err := rest.db.CreateUser(model.User{
		UserName: "Admin",
		UserID:   "SuperAdminID",
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

func TestUserHasAccess(t *testing.T) {
	c, err := config.Get()
	c.ManagerSlackUserID = "SUPERADMINID"
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	r, err := NewRESTAPI(slack)
	assert.NoError(t, err)

	accessLevel, err := r.getAccessLevel("RANDOMID", "RANDOMCHAN")
	assert.Error(t, err)
	assert.Equal(t, 0, accessLevel)

	superAdmin, err := r.db.CreateUser(model.User{
		UserID:   "SUPERADMINID",
		UserName: "SAdmin",
		Role:     "",
	})
	assert.NoError(t, err)

	accessLevel, err = r.getAccessLevel(superAdmin.UserID, "RANDOMCHAN")
	assert.NoError(t, err)
	assert.Equal(t, 1, accessLevel)

	admin, err := r.db.CreateUser(model.User{
		UserID:   "ADMINID",
		UserName: "Admin",
		Role:     "admin",
	})
	assert.NoError(t, err)

	accessLevel, err = r.getAccessLevel(admin.UserID, "RANDOMCHAN")
	assert.NoError(t, err)
	assert.Equal(t, 2, accessLevel)

	pmUser, err := r.db.CreateUser(model.User{
		UserID:   "PMID",
		UserName: "futurePM",
		Role:     "",
	})
	assert.NoError(t, err)

	pm, err := r.db.CreatePM(model.ChannelMember{
		UserID:    pmUser.UserID,
		ChannelID: "RANDOMCHAN",
	})
	assert.NoError(t, err)

	accessLevel, err = r.getAccessLevel(pm.UserID, "RANDOMCHAN")
	assert.NoError(t, err)
	assert.Equal(t, 3, accessLevel)

	user, err := r.db.CreateUser(model.User{
		UserID:   "USERID",
		UserName: "User",
		Role:     "",
	})

	accessLevel, err = r.getAccessLevel(user.UserID, "RANDOMCHAN")
	assert.NoError(t, err)
	assert.Equal(t, 4, accessLevel)

	assert.NoError(t, r.db.DeleteUser(admin.ID))
	assert.NoError(t, r.db.DeleteUser(pmUser.ID))
	assert.NoError(t, r.db.DeleteUser(user.ID))
	assert.NoError(t, r.db.DeleteChannelMember(pmUser.UserID, "RANDOMCHAN"))

}

func getContext(command string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/command", strings.NewReader(command))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)

	return context, rec
}
