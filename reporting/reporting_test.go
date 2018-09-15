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

	channel, err := r.db.CreateChannel(model.Channel{
		ChannelName: "channame",
		ChannelID:   "chanid",
		StandupTime: int64(0),
	})

	dateTo := time.Now()
	dateFrom := time.Now().AddDate(0, 0, -2)

	data := []byte{}

	//First test when no data
	actual, err := r.StandupReportByProject(channel, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected := "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\nNo standup data for this period\n"
	assert.Equal(t, expected, actual)

	d = time.Date(2018, 6, 4, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	//create user who did not write standup
	user1, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	standup0, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		UserID:    user1.UserID,
		Comment:   "",
		MessageTS: "1234",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	//test for no standup submitted
	actual, err = r.StandupReportByProject(channel, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-04\n<@userID1> did not submit standup!\n"
	assert.Equal(t, expected, actual)

	//create standup for user
	standup1, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "my standup",
		UserID:    user1.UserID,
		MessageTS: "123",
	})
	assert.NoError(t, err)

	//test if user submitted standup success
	actual, err = r.StandupReportByProject(channel, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-04\n<@userID1> did not submit standup!\nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n"
	assert.Equal(t, expected, actual)

	//create another user
	user2, err := r.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	//test if one user wrote standup and the other did not
	actual, err = r.StandupReportByProject(channel, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-04\n<@userID1> did not submit standup!\nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n"
	assert.Equal(t, expected, actual)

	//create standup for user2
	standup2, err := r.db.CreateStandup(model.Standup{
		ChannelID: channel.ChannelID,
		Comment:   "user2 standup",
		UserID:    "userID2",
		MessageTS: "1234",
	})
	assert.NoError(t, err)

	//test if both users had written standups
	actual, err = r.StandupReportByProject(channel, dateFrom, dateTo, data)
	assert.NoError(t, err)
	expected = "Full Report on project #channame from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-04\n<@userID1> did not submit standup!\n<@userID2> submitted standup: user2 standup \nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n<@userID2> submitted standup: user2 standup \n"

	assert.Equal(t, expected, actual)

	assert.NoError(t, r.db.DeleteStandup(standup0.ID))
	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
	assert.NoError(t, r.db.DeleteStandup(standup2.ID))
	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
	assert.NoError(t, r.db.DeleteChannelMember(user2.UserID, user2.ChannelID))
	assert.NoError(t, r.db.DeleteChannel(channel.ID))
}

// func TestStandupReportByUser(t *testing.T) {
// 	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	r, err := NewReporter(c)
// 	assert.NoError(t, err)

// 	channelID := "QWERTY123"

// 	dateNext := time.Now().AddDate(0, 0, 1)
// 	dateTo := time.Now()
// 	dateFrom := time.Now().AddDate(0, 0, -2)

// 	user, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID1",
// 		ChannelID: channelID,
// 	})
// 	assert.NoError(t, err)

// 	data := []byte{}

// 	_, err = r.StandupReportByUser(user.UserID, dateTo, dateFrom, data)
// 	assert.Error(t, err)
// 	_, err = r.StandupReportByUser(user.UserID, dateNext, dateTo, data)
// 	assert.Error(t, err)
// 	_, err = r.StandupReportByUser(user.UserID, dateFrom, dateNext, data)
// 	assert.Error(t, err)

// 	expected := "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-04\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-05\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\n"
// 	actual, err := r.StandupReportByUser(user.UserID, dateFrom, dateTo, data)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, actual)

// 	standup1, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channelID,
// 		Comment:   "my standup",
// 		UserID:    "userID1",
// 		MessageTS: "123",
// 	})
// 	expected = "Full Report on user <@userID1> from 2018-06-03 to 2018-06-05:\n\nReport for: 2018-06-03\nIn <#QWERTY123|chanName> <@userID1> did not submit standup!\nReport for: 2018-06-04\nIn <#QWERTY123|chanName> <@userID1> submitted standup: my standup \n\nReport for: 2018-06-05\nIn <#QWERTY123|chanName> <@userID1> submitted standup: my standup \n\n"
// 	actual, err = r.StandupReportByUser(user.UserID, dateFrom, dateTo, data)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expected, actual)

// 	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user.UserID, user.ChannelID))
// }

// func TestStandupReportByProjectAndUser(t *testing.T) {
// 	d := time.Date(2018, 6, 5, 0, 0, 0, 0, time.UTC)
// 	monkey.Patch(time.Now, func() time.Time { return d })
// 	c, err := config.Get()
// 	assert.NoError(t, err)
// 	r, err := NewReporter(c)
// 	assert.NoError(t, err)

// 	channel, err := r.db.CreateChannel(model.Channel{
// 		ChannelName: "channame",
// 		ChannelID:   "chanid",
// 		StandupTime: int64(0),
// 	})

// 	dateTo := time.Now()
// 	dateFrom := time.Now().AddDate(0, 0, -2)

// 	user1, err := r.db.CreateChannelMember(model.ChannelMember{
// 		UserID:    "userID1",
// 		ChannelID: channel.ChannelID,
// 	})

// 	data := []byte{}
// 	actual, err := r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo, data)
// 	assert.NoError(t, err)
// 	expected := "Report on project: <#QWERTY123|chanName>, and user: <@userID1> from 2018-06-03 to 2018-06-05\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> did not submit standup!\nReport for: 2018-06-05\n<@userID1> did not submit standup!\n"
// 	assert.Equal(t, expected, actual)

// 	standup1, err := r.db.CreateStandup(model.Standup{
// 		ChannelID: channel.ChannelID,
// 		Comment:   "my standup",
// 		UserID:    "userID1",
// 		MessageTS: "123",
// 	})

// 	assert.NoError(t, err)
// 	actual, err = r.StandupReportByProjectAndUser(channel, user1.UserID, dateFrom, dateTo, data)
// 	assert.NoError(t, err)
// 	expected = "Report on project: <#QWERTY123|chanName>, and user: <@userID1> from 2018-06-03 to 2018-06-05\n\nReport for: 2018-06-03\n<@userID1> did not submit standup!\nReport for: 2018-06-04\n<@userID1> submitted standup: my standup \nReport for: 2018-06-05\n<@userID1> submitted standup: my standup \n"
// 	assert.Equal(t, expected, actual)

// 	assert.NoError(t, r.db.DeleteStandup(standup1.ID))
// 	assert.NoError(t, r.db.DeleteChannelMember(user1.UserID, user1.ChannelID))
// }
