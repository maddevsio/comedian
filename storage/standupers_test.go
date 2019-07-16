package storage

import (
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateStanduper(t *testing.T) {

	_, err := db.CreateStanduper(model.Standuper{})
	assert.Error(t, err)

	s, err := db.CreateStanduper(model.Standuper{
		Created:   time.Now(),
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
		Created:   time.Now(),
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)

	v, err := db.CreateStanduper(model.Standuper{
		Created:   time.Now(),
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar13",
	})
	assert.NoError(t, err)

	_, err = db.ListStandupers()
	assert.NoError(t, err)

	res, err := db.ListStandupersByTeamID("foo")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))

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

	res, err = db.FindStansupersByUserID("bar")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))

	_, err = db.FindStansuperByUserID("bar", "bar12")
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStanduper(s.ID))
	assert.NoError(t, db.DeleteStanduper(v.ID))
}

func TestUpdateStanduper(t *testing.T) {

	s, err := db.CreateStanduper(model.Standuper{
		Created:   time.Now(),
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", s.RoleInChannel)

	s.RoleInChannel = "developer"

	s, err = db.UpdateStanduper(s)
	assert.NoError(t, err)
	assert.Equal(t, "developer", s.RoleInChannel)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}
