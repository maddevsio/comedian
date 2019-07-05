package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
)

func TestCreateUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	_, err = mysql.CreateUser(model.User{})
	assert.Error(t, err)

	u, err := mysql.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", u.TeamID)

	assert.NoError(t, mysql.DeleteUser(u.ID))
}

func TestGetUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	u, err := mysql.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)

	_, err = mysql.ListUsers()
	assert.NoError(t, err)

	_, err = mysql.SelectUser("")
	assert.Error(t, err)

	_, err = mysql.SelectUser("bar")
	assert.NoError(t, err)

	_, err = mysql.GetUser(int64(0))
	assert.Error(t, err)

	_, err = mysql.GetUser(u.ID)
	assert.NoError(t, err)

	assert.NoError(t, mysql.DeleteUser(u.ID))
}

func TestUpdateUser(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	u, err := mysql.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", u.Role)

	u.Role = "admin"

	u, err = mysql.UpdateUser(u)
	assert.NoError(t, err)
	assert.Equal(t, "admin", u.Role)

	assert.NoError(t, mysql.DeleteUser(u.ID))
}
