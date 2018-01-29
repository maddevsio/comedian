package storage

import (
	"database/sql"
	"testing"

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
		Comment:  "work hard",
		Username: "user",
	})
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "work hard")
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
	assert.NoError(t, m.DeleteStandup(s.ID))
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
	})
	assert.NoError(t, err)
	assert.Equal(t, "@test", su.SlackName)
	assert.Equal(t, "Test Testtt", su.FullName)
	items, err := m.ListStandupUsers()
	assert.NoError(t, err)
	assert.Equal(t, items[0].SlackName, su.SlackName)
	assert.NoError(t, m.DeleteStandupUserByUsername(su.SlackName))
	items, err = m.ListStandupUsers()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(items))
}
