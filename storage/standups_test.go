package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateStandup(t *testing.T) {
	db := setupDB()

	_, err := db.CreateStandup(model.Standup{})
	assert.Error(t, err)

	st, err := db.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", st.TeamID)

	assert.NoError(t, db.DeleteStandup(st.ID))
}

func TestGetStandups(t *testing.T) {
	db := setupDB()

	st, err := db.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)

	_, err = db.ListStandups()
	assert.NoError(t, err)

	_, err = db.SelectStandupByMessageTS("2345")
	assert.Error(t, err)

	_, err = db.SelectStandupByMessageTS("12345")
	assert.NoError(t, err)

	_, err = db.GetStandup(int64(0))
	assert.Error(t, err)

	_, err = db.GetStandup(st.ID)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteStandup(st.ID))
}

func TestUpdateStandup(t *testing.T) {
	db := setupDB()

	st, err := db.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", st.Comment)
	assert.Equal(t, "12345", st.MessageTS)

	st.Comment = "yesterday, today, problems"
	st.MessageTS = "123456"

	st, err = db.UpdateStandup(st)
	assert.NoError(t, err)
	assert.Equal(t, "yesterday, today, problems", st.Comment)
	assert.Equal(t, "123456", st.MessageTS)

	assert.NoError(t, db.DeleteStandup(st.ID))
}
