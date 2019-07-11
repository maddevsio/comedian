package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateStanduper(t *testing.T) {

	_, err := db.CreateStanduper(model.Standuper{})
	assert.Error(t, err)

	s, err := db.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.TeamID)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}

func TestGetStandupers(t *testing.T) {

	s, err := db.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)

	_, err = db.ListStandupers()
	assert.NoError(t, err)

	_, err = db.ListChannelStandupers("")
	assert.NoError(t, err)

	_, err = db.ListChannelStandupers("bar12")
	assert.NoError(t, err)

	_, err = db.GetStanduper(int64(0))
	assert.Error(t, err)

	_, err = db.GetStanduper(s.ID)
	assert.NoError(t, err)

	_, err = db.FindStansuperByUserID("noUser", "bar12")
	assert.Error(t, err)

	_, err = db.FindStansuperByUserID("bar", "bar12")
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}

func TestUpdateStanduper(t *testing.T) {

	s, err := db.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", s.RoleInChannel)
	assert.Equal(t, false, s.SubmittedStandupToday)

	s.RoleInChannel = "developer"
	s.SubmittedStandupToday = true

	s, err = db.UpdateStanduper(s)
	assert.NoError(t, err)
	assert.Equal(t, "developer", s.RoleInChannel)
	assert.Equal(t, true, s.SubmittedStandupToday)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}
