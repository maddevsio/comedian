package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCRUDLStandup(t *testing.T) {

	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	// clean up table before tests
	standups, _ := db.ListStandups()
	for _, standup := range standups {
		db.DeleteStandup(standup.ID)
	}

	s, err := db.CreateStandup(model.Standup{
		ChannelID: "QWERTY123",
		Comment:   "work hard",
		UserID:    "userID1",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	_, err = db.CreateStandup(model.Standup{
		ChannelID: "QWERTY123",
		Comment:   "",
		UserID:    "userID1",
		MessageTS: "",
	})
	assert.Error(t, err)

	assert.Equal(t, s.Comment, "work hard")
	s2, err := db.CreateStandup(model.Standup{
		ChannelID: "ASDF098",
		Comment:   "stubComment",
		UserID:    "illidan",
		MessageTS: "you are not prepared",
	})
	assert.NoError(t, err)
	assert.Equal(t, s2.Comment, "stubComment")
	upd := model.StandupEditHistory{
		Created:     s.Modified,
		StandupID:   s.ID,
		StandupText: s.Comment,
	}
	upd, err = db.AddToStandupHistory(upd)
	assert.NoError(t, err)

	upd1 := model.StandupEditHistory{
		Created:     s.Modified,
		StandupID:   s.ID,
		StandupText: "",
	}
	upd1, err = db.AddToStandupHistory(upd1)
	assert.Error(t, err)

	sps, err := db.SelectStandupsFiltered("userID1", "QWERTY123", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(sps))

	assert.Equal(t, s.ID, upd.StandupID)
	assert.Equal(t, s.Modified, upd.Created)
	assert.Equal(t, s.Comment, upd.StandupText)
	s.Comment = "Rest"
	s, err = db.UpdateStandup(s)
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "Rest")
	items, err := db.ListStandups()
	assert.NoError(t, err)
	assert.Equal(t, items[0], s)
	selectedByMessageTS, err := db.SelectStandupByMessageTS(s2.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s2.MessageTS, selectedByMessageTS.MessageTS)
	selectedByMessageTS, err = db.SelectStandupByMessageTS(s.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageTS, selectedByMessageTS.MessageTS)
	assert.Equal(t, s.UserID, selectedByMessageTS.UserID)

	timeNow := time.Now()
	dateTo := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), timeNow.Second(), 0, time.UTC)
	dateFrom := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, time.UTC)

	_, err = db.SelectStandupsByChannelIDForPeriod(s.ChannelID, dateFrom, dateTo)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStandup(s.ID))
	assert.NoError(t, db.DeleteStandup(s2.ID))
}

func TestCRUDChannelMember(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	// clean up table before tests
	ChannelMembers, _ := db.ListAllChannelMembers()
	for _, user := range ChannelMembers {
		db.DeleteChannelMember(user.UserID, user.ChannelID)
	}

	su1, err := db.CreateChannelMember(model.ChannelMember{
		UserID:      "userID1",
		ChannelID:   "123qwe",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "123qwe", su1.ChannelID)

	su2, err := db.CreateChannelMember(model.ChannelMember{
		UserID:      "userID2",
		ChannelID:   "qwe123",
		StandupTime: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "userID2", su2.UserID)

	listOfChannels, err := db.GetUserChannels(su2.UserID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOfChannels))

	su3, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID3",
		ChannelID: "123qwe",
	})
	assert.NoError(t, err)

	isNonReporter, err := db.IsNonReporter(su3.UserID, su3.ChannelID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	assert.NoError(t, err)
	assert.Equal(t, true, isNonReporter)

	su4, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    "",
		ChannelID: "",
	})
	assert.Error(t, err)
	assert.NoError(t, db.DeleteChannelMember(su4.UserID, su4.ChannelID))
	assert.Equal(t, "userID3", su3.UserID)

	nonReporters, err := db.GetNonReporters(su3.ChannelID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nonReporters))

	user, err := db.FindChannelMemberByUserID(su2.UserID, su2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, su2.UserID, user.UserID)

	users, err := db.ListChannelMembers(su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, users[0].UserID, su1.UserID)

	users, err = db.ListAllChannelMembers()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

	users, err = db.ListChannelMembers(su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

	channels, err := db.GetAllChannels()
	assert.NoError(t, err)
	fmt.Println(channels)
	assert.Equal(t, 2, len(channels))

	assert.NoError(t, db.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(su3.UserID, su3.ChannelID))
}

func TestCRUDStandupTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	ch, err := db.CreateChannel(model.Channel{
		ChannelID:   "CHANNEL1",
		ChannelName: "chanName1",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	ch2, err := db.CreateChannel(model.Channel{
		ChannelID:   "CHANNEL2",
		ChannelName: "chanName2",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)
	assert.Equal(t, "CHANNEL1", ch.ChannelID)
	assert.Equal(t, "chanName1", ch.ChannelName)

	err = db.CreateStandupTime(int64(12), ch.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), ch.StandupTime)

	assert.NoError(t, db.DeleteStandupTime(ch.ChannelID))
	assert.Equal(t, 0, ch.StandupTime)

	time, err := db.GetChannelStandupTime(ch.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, 0, time)

	err = db.CreateStandupTime(int64(12), ch2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), ch2.StandupTime)

	time, err = db.GetChannelStandupTime(ch2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), time)

	allStandupTimes, err := db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(allStandupTimes))

	assert.NoError(t, db.DeleteStandupTime(ch.ChannelID))
	assert.NoError(t, db.DeleteStandupTime(ch2.ChannelID))

	allStandupTimes, err = db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(allStandupTimes))
}
