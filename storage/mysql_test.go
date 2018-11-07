package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"

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

	ch, err := db.CreateChannel(model.Channel{
		ChannelID:   "QWERTY123",
		ChannelName: "chanName1",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	submittedStandupToday := db.SubmittedStandupToday("userID1", ch.ChannelID)
	assert.Equal(t, false, submittedStandupToday)

	s, err := db.CreateStandup(model.Standup{
		ChannelID: ch.ChannelID,
		Comment:   "work hard",
		UserID:    "userID1",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	//bot is able to create empty standups
	_, err = db.CreateStandup(model.Standup{
		ChannelID: ch.ChannelID,
		Comment:   "",
		UserID:    "userID1",
		MessageTS: "",
	})
	assert.NoError(t, err)

	d := time.Date(2018, 6, 24, 9, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	assert.Equal(t, s.Comment, "work hard")
	s2, err := db.CreateStandup(model.Standup{
		ChannelID: ch.ChannelID,
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

	d = time.Date(2018, 6, 24, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	submittedStandupToday = db.SubmittedStandupToday("illidan", ch.ChannelID)
	assert.Equal(t, true, submittedStandupToday)

	upd1 := model.StandupEditHistory{
		Created:     s.Modified,
		StandupID:   s.ID,
		StandupText: "",
	}
	upd1, err = db.AddToStandupHistory(upd1)
	assert.Error(t, err)

	_, err = db.SelectStandupsFiltered("userID1", "QWERTY123", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
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
	assert.NoError(t, db.DeleteChannel(ch.ID))
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

	u1, err := db.CreateUser(model.User{
		UserName: "firstUser",
		UserID:   "userID1",
	})
	assert.NoError(t, err)

	su1, err := db.CreateChannelMember(model.ChannelMember{
		UserID:        "userID1",
		ChannelID:     "123qwe",
		StandupTime:   0,
		RoleInChannel: "developer",
	})
	assert.NoError(t, err)
	assert.Equal(t, "123qwe", su1.ChannelID)

	_, err = db.FindChannelMemberByUserName(u1.UserName, "123qwe")
	assert.NoError(t, err)

	su2, err := db.CreateChannelMember(model.ChannelMember{
		UserID:        "userID2",
		ChannelID:     "qwe123",
		StandupTime:   0,
		RoleInChannel: "developer",
	})
	assert.NoError(t, err)
	assert.Equal(t, "userID2", su2.UserID)

	listOfChannels, err := db.GetUserChannels(su2.UserID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listOfChannels))

	su3, err := db.CreateChannelMember(model.ChannelMember{
		UserID:        "userID3",
		ChannelID:     "123qwe",
		RoleInChannel: "developer",
	})
	assert.NoError(t, err)

	isNonReporter, err := db.IsNonReporter(su3.UserID, su3.ChannelID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	assert.Error(t, err)
	assert.Equal(t, false, isNonReporter)

	su4, err := db.CreateChannelMember(model.ChannelMember{
		UserID:        "",
		ChannelID:     "",
		RoleInChannel: "developer",
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

	chm, err := db.SelectChannelMember(su2.ID)
	assert.NoError(t, err)
	assert.Equal(t, su2.UserID, chm.UserID)

	_, err = db.SelectChannelMember(345)
	assert.Error(t, err)

	users, err := db.ListChannelMembersByRole(su1.ChannelID, "developer")
	assert.NoError(t, err)
	assert.Equal(t, users[0].UserID, su1.UserID)

	users, err = db.ListAllChannelMembers()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(users))

	members, err := db.FindMembersByUserID(user.UserID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(members))

	users, err = db.ListChannelMembers(su1.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(users))

	assert.NoError(t, db.DeleteChannelMember(su1.UserID, su1.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(su2.UserID, su2.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(su3.UserID, su3.ChannelID))
	assert.NoError(t, db.DeleteUser(u1.ID))

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

	err = db.UpdateChannelStandupTime(int64(120), ch.ChannelID)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStandupTime(ch.ChannelID))
	assert.Equal(t, int64(0), ch.StandupTime)

	time, err := db.GetChannelStandupTime(ch.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), time)

	err = db.CreateStandupTime(int64(12), ch2.ChannelID)
	assert.NoError(t, err)

	time, err = db.GetChannelStandupTime(ch2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), time)

	allStandupTimes, err := db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allStandupTimes))

	assert.NoError(t, db.DeleteStandupTime(ch.ChannelID))
	assert.NoError(t, db.DeleteStandupTime(ch2.ChannelID))

	allStandupTimes, err = db.ListAllStandupTime()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(allStandupTimes))

	channels, err := db.GetAllChannels()
	for _, channel := range channels {
		ch, err := db.SelectChannel(channel.ChannelID)
		assert.NoError(t, err)
		assert.NoError(t, db.DeleteChannel(ch.ID))
	}
}

func TestCRUDChannel(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	ch, err := db.CreateChannel(model.Channel{
		ChannelID:   "QWERTY123",
		ChannelName: "chanName1",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	channelName, err := db.GetChannelName(ch.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, ch.ChannelName, channelName)

	channelID, err := db.GetChannelID(ch.ChannelName)
	assert.NoError(t, err)
	assert.Equal(t, ch.ChannelID, channelID)

	chans, err := db.GetChannels()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chans))

	assert.NoError(t, db.DeleteChannel(ch.ID))
}

func TestCRUDUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	user, err := db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})

	_, err = db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "admin",
	})
	assert.NoError(t, err)

	u1, err := db.SelectUserByUserName(user.UserName)
	assert.NoError(t, err)
	assert.Equal(t, user, u1)

	u2, err := db.SelectUser(user.UserID)
	assert.NoError(t, err)
	assert.Equal(t, user, u2)

	admins, err := db.ListAdmins()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(admins))

	user.Role = "admin"

	_, err = db.UpdateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Role)

	users, err := db.ListUsers()
	assert.NoError(t, err)
	for _, user := range users {
		assert.NoError(t, db.DeleteUser(user.ID))
	}
}

func TestPMForProject(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	ch, err := db.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	m1, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    "ID1",
		ChannelID: ch.ChannelID,
	})
	assert.NoError(t, err)
	m2, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    "ID2",
		ChannelID: ch.ChannelID,
	})
	assert.NoError(t, err)
	m3, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    "ID3",
		ChannelID: ch.ChannelID,
	})
	assert.NoError(t, err)

	pm, err := db.CreateChannelMember(model.ChannelMember{
		UserID:        "ID4",
		ChannelID:     ch.ChannelID,
		RoleInChannel: "pm",
	})
	assert.NoError(t, err)

	ok1 := db.UserIsPMForProject(m1.UserID, m1.ChannelID)
	assert.Equal(t, false, ok1)
	ok2 := db.UserIsPMForProject(m2.UserID, m2.ChannelID)
	assert.Equal(t, false, ok2)
	ok3 := db.UserIsPMForProject(m3.UserID, m3.ChannelID)
	assert.Equal(t, false, ok3)
	ok4 := db.UserIsPMForProject(pm.UserID, pm.ChannelID)
	assert.Equal(t, true, ok4)

	pms, err := db.ListChannelMembersByRole(ch.ChannelID, "pm")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pms))

	assert.NoError(t, db.DeleteChannelMember(m1.UserID, m1.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(m2.UserID, m2.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(m3.UserID, m3.ChannelID))
	assert.NoError(t, db.DeleteChannelMember(pm.UserID, pm.ChannelID))
	assert.NoError(t, db.DeleteChannel(ch.ID))

}

func TestCRUDTimeTable(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	db, err := NewMySQL(c)
	assert.NoError(t, err)

	user, err := db.CreateUser(model.User{
		UserID:   "QWERTY123",
		UserName: "chanName1",
		Role:     "",
	})
	assert.NoError(t, err)

	channel, err := db.CreateChannel(model.Channel{
		ChannelID:   "XYZ",
		ChannelName: "chan",
		StandupTime: int64(0),
	})
	assert.NoError(t, err)

	m, err := db.CreateChannelMember(model.ChannelMember{
		UserID:    user.UserID,
		ChannelID: channel.ChannelID,
	})
	assert.NoError(t, err)

	memberHasTimeTable := db.MemberHasTimeTable(m.ID)
	assert.Equal(t, false, memberHasTimeTable)

	_, err = db.CreateTimeTable(model.TimeTable{
		ChannelMemberID: m.ID,
	})
	assert.NoError(t, err)

	memberHasTimeTable = db.MemberHasTimeTable(m.ID)
	assert.Equal(t, true, memberHasTimeTable)

	tts, err := db.SelectTimeTable(m.ID)
	assert.NoError(t, err)

	tts.Monday = int64(10000)

	_, err = db.UpdateTimeTable(tts)
	assert.NoError(t, err)

	timeTables, err := db.ListTimeTablesForDay("monday")
	assert.NoError(t, err)
	fmt.Println("All timetables: ", timeTables)
	assert.Equal(t, 1, len(timeTables))

	assert.NoError(t, db.DeleteUser(user.ID))
	assert.NoError(t, db.DeleteChannel(channel.ID))
	assert.NoError(t, db.DeleteChannelMember(user.UserID, channel.ChannelID))
	assert.NoError(t, db.DeleteTimeTable(tts.ID))
}
