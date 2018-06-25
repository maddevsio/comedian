package storage

import (
	"database/sql"
	"testing"

	"fmt"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCRUDLStandup(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	s, err := m.CreateStandup(model.Standup{
		ChannelID: "QWERTY123",
		Comment:   "work hard",
		Username:  "user",
		MessageTS: "qweasdzxc",
	})
	s2, err := m.CreateStandup(model.Standup{
		ChannelID: "ASDF098",
		Comment:   "stubComment",
		Username:  "illidan",
		MessageTS: "you are not prepared",
	})
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "work hard")
	upd := model.StandupEditHistory{
		Created:     s.Modified,
		StandupID:   s.ID,
		StandupText: s.Comment,
	}
	upd, err = m.AddToStandupHistory(upd)
	assert.NoError(t, err)
	assert.Equal(t, s.ID, upd.StandupID)
	assert.Equal(t, s.Modified, upd.Created)
	assert.Equal(t, s.Comment, upd.StandupText)
	s.Comment = "Rest"
	s, err = m.UpdateStandup(s)
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "Rest")
	items, err := m.ListStandups()
	assert.NoError(t, err)
	assert.Equal(t, items[0], s)
	selected, err := m.SelectStandup(s.ID)
	assert.NoError(t, err)
	assert.Equal(t, s, selected)
	selectedByChannelID, err := m.SelectStandupByChannelID(s2.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, s2.Comment, selectedByChannelID[0].Comment)
	assert.Equal(t, s2.Username, selectedByChannelID[0].Username)
	selectedByChannelID, err = m.SelectStandupByChannelID(s.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, selectedByChannelID[0].Comment)
	assert.Equal(t, s.Username, selectedByChannelID[0].Username)
	selectedByMessageTS, err := m.SelectStandupByMessageTS(s2.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s2.MessageTS, selectedByMessageTS.MessageTS)
	assert.Equal(t, s2.Username, selectedByMessageTS.Username)
	selectedByMessageTS, err = m.SelectStandupByMessageTS(s.MessageTS)
	assert.NoError(t, err)
	assert.Equal(t, s.MessageTS, selectedByMessageTS.MessageTS)
	assert.Equal(t, s.Username, selectedByMessageTS.Username)
	assert.NoError(t, m.DeleteStandup(s.ID))
	assert.NoError(t, m.DeleteStandup(s2.ID))
	s, err = m.SelectStandup(s.ID)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Equal(t, s.ID, int64(0))
}

func TestCRUDStandupUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	su, err := m.CreateStandupUser(model.StandupUser{
		SlackName: "@test",
		FullName:  "Test Testtt",
		ChannelID: "chanid",
		Channel:   "chanName",
	})
	assert.NoError(t, err)
	assert.Equal(t, "@test", su.SlackName)
	assert.Equal(t, "Test Testtt", su.FullName)
	assert.Equal(t, "chanid", su.ChannelID)
	assert.Equal(t, "chanName", su.Channel)
	items, err := m.ListStandupUsersByChannelID(su.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, items[0].SlackName, su.SlackName)
	assert.NoError(t, m.DeleteStandupUserByUsername(su.SlackName, su.ChannelID))
	items, err = m.ListStandupUsersByChannelID(su.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(items))
}

func TestCRUDStandupTime(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	st, err := m.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid",
		Channel:   "chanName",
		Time:      int64(12),
	})
	assert.NoError(t, err)
	assert.Equal(t, "chanid", st.ChannelID)
	assert.Equal(t, "chanName", st.Channel)
	assert.Equal(t, int64(12), st.Time)
	time, err := m.ListStandupTime(st.ChannelID)
	assert.NoError(t, err)
	assert.Equal(t, time.Time, st.Time)
	st2, err := m.CreateStandupTime(model.StandupTime{
		ChannelID: "chanid222",
		Channel:   "chanName2",
		Time:      int64(13),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(13), st2.Time)
	_, err = m.ListAllStandupTime()
	assert.NoError(t, err)

	st.ChannelID = "'"
	time, err = m.ListStandupTime(st.ChannelID)
	fmt.Printf("DATABASE ERROR: %v", err)
	assert.Error(t, err)
	st.ChannelID = "chanid"
	assert.NoError(t, m.DeleteStandupTime(st.ChannelID))
	time, err = m.ListStandupTime(st.ChannelID)
	assert.Error(t, err)
	assert.Equal(t, int64(0), time.Time)
}
