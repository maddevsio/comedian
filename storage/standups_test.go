package storage

import (
	"testing"
	"time"

	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateStandup(t *testing.T) {

	_, err := db.CreateStandup(model.Standup{})
	assert.Error(t, err)

	st, err := db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
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

	st, err := db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)

	_, err = db.ListStandups()
	assert.NoError(t, err)

	st, err = db.SelectLatestStandupByUser("bar", "bar12")
	assert.NoError(t, err)
	assert.Equal(t, "12345", st.MessageTS)

	_, err = db.SelectLatestStandupByUser("foo", "bar12")
	assert.Error(t, err)

	_, err = db.SelectStandupByMessageTS("2345")
	assert.Error(t, err)

	_, err = db.SelectStandupByMessageTS("12345")
	assert.NoError(t, err)

	_, err = db.GetStandup(int64(0))
	assert.Error(t, err)

	_, err = db.GetStandup(st.ID)
	assert.NoError(t, err)

	res, err := db.GetStandupForPeriod("bar", "bar12", time.Now().Add(10*time.Second*(-1)), time.Now().Add(10*time.Second))
	assert.NoError(t, err)
	assert.Equal(t, "12345", res.MessageTS)

	_, err = db.GetStandupForPeriod("foo", "bar12", time.Now().Add(10*time.Second*(-1)), time.Now().Add(10*time.Second))
	assert.Error(t, err)

	_, err = db.GetStandupForPeriod("foo", "bar12", time.Now().Add(10*time.Hour*(-1)), time.Now().Add(10*time.Second*(-1)))
	assert.Error(t, err)

	assert.NoError(t, db.DeleteStandup(st.ID))
}

func TestUpdateStandup(t *testing.T) {

	st, err := db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
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
