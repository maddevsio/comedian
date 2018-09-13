package reporting

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestStandupReportByProject(t *testing.T) {
	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	c, err := config.Get()
	assert.NoError(t, err)
	r, err := NewReporter(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	data := []byte{}

	//First test when no data
	actual, err := r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.Error(t, err)
	expected := "Channel does not exist or no users are set as standupers in the channel"
	assert.Equal(t, expected, actual)

	//create user who did not write standup
	user1, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test for no standup submitted
	actual, err = r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project <#QWERTY123|chanName> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> did not submit standup!\nReport for: 2018-06-05\n<@userID1> did not submit standup!\n"
	assert.Equal(t, expected, actual)

	//create standup for user
	standup1, err := r.DB.CreateStandup(model.Standup{
		Channel:    channelName,
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	assert.NoError(t, err)

	//test if user submitted standup success
	actual, err = r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project <#QWERTY123|chanName> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> submitted standup: my standup \n\nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n\n"
	assert.Equal(t, expected, actual)

	//create another user
	user2, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	actual, err = r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project <#QWERTY123|chanName> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!<@userID2> did not submit standup!\nReport for: 2018-06-04\n<@userID1> submitted standup: my standup \n<@userID2> did not submit standup!\nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n<@userID2> did not submit standup!\n"
	assert.Equal(t, expected, actual)

	//create standup for user2
	standup2, err := r.DB.CreateStandup(model.Standup{
		Channel:    channelName,
		ChannelID:  channelID,
		Comment:    "user2 standup",
		UsernameID: "userID2",
		Username:   "user2",
		MessageTS:  "1234",
	})
	assert.NoError(t, err)

	//test if both users had written standups
	actual, err = r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project <#QWERTY123|chanName> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!<@userID2> did not submit standup!\nReport for: 2018-06-04\n<@userID1> submitted standup: my standup \n<@userID2> submitted standup: user2 standup \n\nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n<@userID2> submitted standup: user2 standup \n\n"
	assert.Equal(t, expected, actual)

	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandup(standup2.ID))
	assert.NoError(t, r.DB.DeleteStandupUser(user1.SlackName, user1.ChannelID))
	assert.NoError(t, r.DB.DeleteStandupUser(user2.SlackName, user2.ChannelID))
}

func TestStandupReportByUser(t *testing.T) {
	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	r, err := NewReporter(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	user, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
		Created:     time.Now(),
		Modified:    time.Now(),
	})
	assert.NoError(t, err)

	data := []byte{}

	_, err = r.StandupReportByUser(user.SlackUserID, dateTo, dateFrom, data)
	assert.Error(t, err)
	_, err = r.StandupReportByUser(user.SlackUserID, dateNext, dateTo, data)
	assert.Error(t, err)
	_, err = r.StandupReportByUser(user.SlackUserID, dateFrom, dateNext, data)
	assert.Error(t, err)

	expected := "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-04\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-05\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\n"
	actual, err := r.StandupReportByUser(user.SlackUserID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	standup1, err := r.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	expected = "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-04\nIn <#QWERTY123|chanName> <@userID1> submitted standup: my standup \n\nReport for: 2018-06-05\nIn <#QWERTY123|chanName> <@userID1> submitted standup: my standup \n\n"
	actual, err = r.StandupReportByUser(user.SlackUserID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandupUser(user.SlackName, user.ChannelID))
}

func TestStandupReportByProjectAndUser(t *testing.T) {
	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	c, err := config.Get()
	assert.NoError(t, err)
	r, err := NewReporter(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	user1, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})

	data := []byte{}
	actual, err := r.StandupReportByProjectAndUser(channelID, user1.SlackUserID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected := "Report on project: <#QWERTY123|chanName>, and user: <@userID1> from 2018-06-03 to 2018-06-05\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> did not submit standup!\nReport for: 2018-06-05\n<@userID1> did not submit standup!\n"
	assert.Equal(t, expected, actual)

	standup1, err := r.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})

	assert.NoError(t, err)
	actual, err = r.StandupReportByProjectAndUser(channelID, user1.SlackUserID, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Report on project: <#QWERTY123|chanName>, and user: <@userID1> from 2018-06-03 to 2018-06-05\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> submitted standup: my standup \nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n"
	assert.Equal(t, expected, actual)

	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandupUser(user1.SlackName, user1.ChannelID))
}
