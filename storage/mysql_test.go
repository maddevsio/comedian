package storage

import (
	"database/sql"
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
		Channel:   "First Channel",
		ChannelID: "QWERTY123",
		Comment:   "work hard",
		Username:  "user",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	nots, err := db.CreateStandup(model.Standup{
		Channel:   "First Channel",
		ChannelID: "QWERTY123",
		Comment:   "",
		Username:  "user",
		MessageTS: "",
	})
	assert.Error(t, err)
	assert.NoError(t, db.DeleteStandupUserByUsername(nots.Username, nots.ChannelID))

	assert.Equal(t, s.Comment, "work hard")
	s2, err := db.CreateStandup(model.Standup{
		Channel:   "Second Channel",
		ChannelID: "ASDF098",
		Comment:   "stubComment",
		Username:  "illidan",
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
	selected, err := db.SelectStandup(s.ID)
	assert.NoError(t, err)
	assert.Equal(t, s, selected)
	selectedByChannelID, err := db.SelectStandupsByChannelID(s2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, s2.Comment, selectedByChannelID[0].Comment)
	assert.Equal(t, s2.Username, selectedByChannelID[0].Username)
	selectedByChannelID, err = db.SelectStandupsByChannelID(s.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, selectedByChannelID[0].Comment)
	assert.Equal(t, s.Username, selectedByChannelID[0].Username)
	selectedByMessageTS, err := db.SelectStandupByMessageTS(s2.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s2.MessageTS, selectedByMessageTS.MessageTS)
	assert.Equal(t, s2.Username, selectedByMessageTS.Username)
	selectedByMessageTS, err = db.SelectStandupByMessageTS(s.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageTS, selectedByMessageTS.MessageTS)
	assert.Equal(t, s.Username, selectedByMessageTS.Username)

	timeNow := time.Now()
	dateTo := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), timeNow.Hour(), timeNow.Minute(), timeNow.Second(), 0, time.UTC)
	dateFrom := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, time.UTC)

	_, err = db.SelectStandupsForPeriod(dateFrom, dateTo)
	assert.NoError(t, err)
	//assert.Equal(t, 2, len(SelectStandupsForPeriod))

	_, err = db.SelectStandupsByChannelIDForPeriod(s.ChannelID, dateFrom, dateTo)
	assert.NoError(t, err)
	//assert.Equal(t, 1, len(SelectStandupsByChannelIDForPeriod))

	_, err = db.SelectStandupByUserNameForPeriod(s.Username, dateFrom, dateTo)
	assert.NoError(t, err)
	//assert.Equal(t, 1, len(SelectStandupByUserNameForPeriod))

	assert.NoError(t, db.DeleteStandup(s.ID))
	assert.NoError(t, db.DeleteStandup(s2.ID))
	s, err = db.SelectStandup(s.ID)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Equal(t, s.ID, int64(0))
}

func TestCRUDStandupUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	// clean up table before tests
	standupUsers, _ := db.ListAllStandupUsers()
	for _, user := range standupUsers {
		db.DeleteStandupUserByUsername(user.SlackName, user.ChannelID)
	}

	su1, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID1",
		SlackName:   "user1",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})

	assert.NoError(t, err)

	assert.Equal(t, "channel1", su1.Channel)

	su2, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID2",
		SlackName:   "user2",
		ChannelID:   "qwe123",
		Channel:     "channel2",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user2", su2.SlackName)

	su3, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "userID3",
		SlackName:   "user3",
		ChannelID:   "123qwe",
		Channel:     "channel1",
	})
	assert.NoError(t, err)

	su4, err := db.CreateStandupUser(model.StandupUser{
		SlackUserID: "",
		SlackName:   "",
		ChannelID:   "",
		Channel:     "",
	})
	assert.Error(t, err)
	assert.NoError(t, db.DeleteStandupUserByUsername(su4.SlackName, su4.ChannelID))

	assert.Equal(t, "userID3", su3.SlackUserID)

	user, err := db.FindStandupUser(su1.SlackName)
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.SlackName)

	user, err = db.FindStandupUserInChannel(su1.SlackName, su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.SlackName)
	assert.Equal(t, "123qwe", user.ChannelID)

	user, err = db.FindStandupUserInChannel(su2.SlackName, su1.ChannelID)
	assert.Error(t, err)

	users, err := db.ListStandupUsersByChannelID(su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, users[0].SlackName, su1.SlackName)

	users, err = db.ListAllStandupUsers()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))

	users, err = db.ListStandupUsersByChannelID(su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

	assert.NoError(t, db.DeleteStandupUserByUsername(su1.SlackName, su1.ChannelID))
	assert.NoError(t, db.DeleteStandupUserByUsername(su2.SlackName, su2.ChannelID))
	assert.NoError(t, db.DeleteStandupUserByUsername(su3.SlackName, su3.ChannelID))

}

func TestCRUDStandupTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	//clean up table before tests
	ast, err := db.ListAllStandupTime()
	for _, st := range ast {
		db.DeleteStandupTime(st.ChannelID)
	}
	st, err := db.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid",
		Channel:   "chanName",
		Time:      int64(12),
	})
	assert.NoError(t, err)

	nost, err := db.CreateStandupTime(model.StandupTime{
		ChannelID: "",
		Channel:   "",
		Time:      0,
	})
	assert.Error(t, err)
	assert.NoError(t, db.DeleteStandupTime(nost.ChannelID))

	assert.Equal(t, "chanid", st.ChannelID)
	assert.Equal(t, "chanName", st.Channel)
	assert.Equal(t, int64(12), st.Time)

	time, err := db.ListStandupTime(st.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, time.Time, st.Time)

	st2, err := db.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid222",
		Channel:   "chanName2",
		Time:      int64(13),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(13), st2.Time)

	st.ChannelID = "'"
	time, err = db.ListStandupTime(st.ChannelID)
	assert.Error(t, err)
	st.ChannelID = "chanid"

	allStandupTimes, err := db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(allStandupTimes))

	assert.NoError(t, db.DeleteStandupTime(st.ChannelID))
	assert.NoError(t, db.DeleteStandupTime(st2.ChannelID))

	allStandupTimes, err = db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(allStandupTimes))

	time, err = db.ListStandupTime(st.ChannelID)
	assert.Error(t, err)
	assert.Equal(t, int64(0), time.Time)
}
