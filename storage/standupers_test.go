package storage

import (
	"testing"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateStanduper(t *testing.T) {

	_, err := db.CreateStanduper(model.Standuper{})
	assert.Error(t, err)

	s, err := db.CreateStanduper(model.Standuper{
		CreatedAt:   time.Now().Unix(),
		WorkspaceID: "foo",
		UserID:      "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.WorkspaceID)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}

func TestGetStandupers(t *testing.T) {

	s, err := db.CreateStanduper(model.Standuper{
		CreatedAt:   time.Now().Unix(),
		WorkspaceID: "foo",
		UserID:      "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)

	v, err := db.CreateStanduper(model.Standuper{
		CreatedAt:   time.Now().Unix(),
		WorkspaceID: "foo",
		UserID:      "bar",
		ChannelID:   "bar13",
	})
	assert.NoError(t, err)

	_, err = db.ListStandupers()
	assert.NoError(t, err)

	res, err := db.ListStandupersByWorkspaceID("foo")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))

	_, err = db.ListProjectStandupers("")
	assert.NoError(t, err)

	_, err = db.ListProjectStandupers("bar12")
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
		CreatedAt:   time.Now().Unix(),
		WorkspaceID: "foo",
		UserID:      "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", s.Role)

	s.Role = "developer"

	s, err = db.UpdateStanduper(s)
	assert.NoError(t, err)
	assert.Equal(t, "developer", s.Role)

	assert.NoError(t, db.DeleteStanduper(s.ID))
}
