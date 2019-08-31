package storage

import (
	"testing"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateProject(t *testing.T) {
	_, err := db.CreateProject(model.Project{})
	assert.Error(t, err)

	ch, err := db.CreateProject(model.Project{
		CreatedAt:   time.Now().Unix(),
		WorkspaceID: "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
		Deadline:    "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", ch.WorkspaceID)

	assert.NoError(t, db.DeleteProject(ch.ID))
}

func TestGetProjects(t *testing.T) {
	ch, err := db.CreateProject(model.Project{
		WorkspaceID: "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
		Deadline:    "",
	})
	assert.NoError(t, err)

	_, err = db.ListProjects()
	assert.NoError(t, err)

	_, err = db.SelectProject("")
	assert.Error(t, err)

	_, err = db.SelectProject("bar12")
	assert.NoError(t, err)

	_, err = db.GetProject(int64(0))
	assert.Error(t, err)

	_, err = db.GetProject(ch.ID)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteProject(ch.ID))
}

func TestUpdateProjects(t *testing.T) {
	ch, err := db.CreateProject(model.Project{
		WorkspaceID: "foo",
		ChannelName: "bar",
		ChannelID:   "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", ch.Deadline)

	ch.Deadline = "10:00"
	ch, err = db.UpdateProject(ch)
	assert.NoError(t, err)
	assert.Equal(t, "10:00", ch.Deadline)

	assert.NoError(t, db.DeleteProject(ch.ID))
}
