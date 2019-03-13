package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestCreateStandup(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	_, err = mysql.CreateStandup(model.Standup{})
	assert.Error(t, err)

	st, err := mysql.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", st.TeamID)

	assert.NoError(t, mysql.DeleteStandup(st.ID))
}

func TestGetStandups(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	st, err := mysql.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: "12345",
	})
	assert.NoError(t, err)

	_, err = mysql.ListStandups()
	assert.NoError(t, err)

	_, err = mysql.SelectStandupByMessageTS("2345")
	assert.Error(t, err)

	_, err = mysql.SelectStandupByMessageTS("12345")
	assert.NoError(t, err)

	_, err = mysql.GetStandup(int64(0))
	assert.Error(t, err)

	_, err = mysql.GetStandup(st.ID)
	assert.NoError(t, err)

	assert.NoError(t, mysql.DeleteStandup(st.ID))
}

func TestUpdateStandup(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	st, err := mysql.CreateStandup(model.Standup{
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

	st, err = mysql.UpdateStandup(st)
	assert.NoError(t, err)
	assert.Equal(t, "yesterday, today, problems", st.Comment)
	assert.Equal(t, "123456", st.MessageTS)

	assert.NoError(t, mysql.DeleteStandup(st.ID))
}
