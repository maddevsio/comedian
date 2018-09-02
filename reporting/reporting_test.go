package reporting

import (
	"fmt"
	"testing"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestStandupReportByProject(t *testing.T) {
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
	text, err := r.StandupReportByProject(channelID, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.Equal(t, "Full Report on project <#QWERTY123>:\n\nNo standup data for this period", text)

	//create user who did not write standup
	user1, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test for no standup submitted
	text, err = r.StandupReportByProject(channelName, dateFrom, dateTo, data)
	assert.NoError(t, err)

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
	text, err = r.StandupReportByProject(channelName, dateFrom, dateTo, data)
	assert.NoError(t, err)

	//create another user
	user2, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   channelID,
		Channel:     channelName,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	text, err = r.StandupReportByProject(channelName, dateFrom, dateTo, data)
	assert.NoError(t, err)

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
	text, err = r.StandupReportByProject(channelName, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandup(standup2.ID))
	assert.NoError(t, r.DB.DeleteStandupUserByUsername(user1.SlackName, user1.ChannelID))
	assert.NoError(t, r.DB.DeleteStandupUserByUsername(user2.SlackName, user2.ChannelID))
}

func TestStandupReportByUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	r, err := NewReporter(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	dateNextText := fmt.Sprintf("%d-%02d-%02d", dateNext.Year(), dateNext.Month(), dateNext.Day())
	dateToText := fmt.Sprintf("%d-%02d-%02d", dateTo.Year(), dateTo.Month(), dateTo.Day())
	//dateFromText := fmt.Sprintf("%d-%02d-%02d", dateFrom.Year(), dateFrom.Month(), dateFrom.Day())

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

	text, err := r.StandupReportByUser(user, dateTo, dateFrom, data)
	assert.Error(t, err)
	text, err = r.StandupReportByUser(user, dateNext, dateTo, data)
	assert.Error(t, err)
	text, err = r.StandupReportByUser(user, dateFrom, dateNext, data)
	assert.Error(t, err)
	text, err = r.StandupReportByUser(user, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Full Report on user <@user1>:\n\n\n\nReport from %v to %v:\nIn channel: <#QWERTY123>\n\n<@user1>: ignored standup!\n\n", dateToText, dateNextText), text)

	standup1, err := r.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	text, err = r.StandupReportByUser(user, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Full Report on user <@user1>:\n\n\n\nReport from %v to %v:\nIn channel: <#QWERTY123>\nmy standup\n\n", dateToText, dateNextText), text)

	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandupUserByUsername(user.SlackName, user.ChannelID))
}

func TestStandupReportByProjectAndUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	r, err := NewReporter(c)
	assert.NoError(t, err)

	channelID := "QWERTY123"
	channelName := "chanName"

	dateNext := time.Now().AddDate(0, 0, 1)
	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	dateNextText := fmt.Sprintf("%d-%02d-%02d", dateNext.Year(), dateNext.Month(), dateNext.Day())
	dateToText := fmt.Sprintf("%d-%02d-%02d", dateTo.Year(), dateTo.Month(), dateTo.Day())
	//dateFromText := fmt.Sprintf("%d-%02d-%02d", dateFrom.Year(), dateFrom.Month(), dateFrom.Day())

	user1, err := r.DB.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   channelID,
		Channel:     channelName,
	})

	data := []byte{}
	text, err := r.StandupReportByProjectAndUser(channelID, user1, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Report on project: <#QWERTY123>, and user: <@user1>\n\n\n\nReport from %v to %v:\n\n<@user1>: ignored standup!\n", dateToText, dateNextText), text)
	standup1, err := r.DB.CreateStandup(model.Standup{
		ChannelID:  channelID,
		Comment:    "my standup",
		UsernameID: "userID1",
		Username:   "user1",
		MessageTS:  "123",
	})
	assert.NoError(t, err)

	text, err = r.StandupReportByProjectAndUser(channelID, user1, dateFrom, dateTo, data)
	assert.NoError(t, err)

	assert.Equal(t, fmt.Sprintf("Report on project: <#QWERTY123>, and user: <@user1>\n\n\n\nReport from %v to %v:\n\nStandup from <@userID1>:\nmy standup\n", dateToText, dateNextText), text)
	assert.NoError(t, r.DB.DeleteStandup(standup1.ID))
	assert.NoError(t, r.DB.DeleteStandupUserByUsername(user1.SlackName, user1.ChannelID))
}
