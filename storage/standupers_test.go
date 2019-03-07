package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestCreateStanduper(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	_, err = mysql.CreateStanduper(model.Standuper{})
	assert.Error(t, err)

	s, err := mysql.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.TeamID)

	assert.NoError(t, mysql.DeleteStanduper(s.ID))
}

func TestGetStandupers(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	s, err := mysql.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)

	_, err = mysql.ListStandupers()
	assert.NoError(t, err)

	_, err = mysql.ListChannelStandupers("")
	assert.NoError(t, err)

	_, err = mysql.ListChannelStandupers("bar12")
	assert.NoError(t, err)

	_, err = mysql.GetStanduper(int64(0))
	assert.Error(t, err)

	_, err = mysql.GetStanduper(s.ID)
	assert.NoError(t, err)

	_, err = mysql.FindStansuperByUserID("noUser", "bar12")
	assert.Error(t, err)

	_, err = mysql.FindStansuperByUserID("bar", "bar12")
	assert.NoError(t, err)

	assert.NoError(t, mysql.DeleteStanduper(s.ID))
}

func TestUpdateStanduper(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := NewMySQL(c)
	assert.NoError(t, err)

	s, err := mysql.CreateStanduper(model.Standuper{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", s.RoleInChannel)
	assert.Equal(t, false, s.SubmittedStandupToday)

	s.RoleInChannel = "developer"
	s.SubmittedStandupToday = true

	s, err = mysql.UpdateStanduper(s)
	assert.NoError(t, err)
	assert.Equal(t, "developer", s.RoleInChannel)
	assert.Equal(t, true, s.SubmittedStandupToday)

	assert.NoError(t, mysql.DeleteStanduper(s.ID))
}
