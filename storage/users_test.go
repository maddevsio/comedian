package storage

import (

	// This line is must for working MySQL database
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {

	_, err := db.CreateUser(model.User{})
	assert.Error(t, err)

	u, err := db.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", u.TeamID)

	assert.NoError(t, db.DeleteUser(u.ID))
}

func TestGetUser(t *testing.T) {

	u, err := db.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)

	_, err = db.ListUsers()
	assert.NoError(t, err)

	_, err = db.SelectUser("")
	assert.Error(t, err)

	_, err = db.SelectUser("bar")
	assert.NoError(t, err)

	_, err = db.GetUser(int64(0))
	assert.Error(t, err)

	_, err = db.GetUser(u.ID)
	assert.NoError(t, err)

	assert.NoError(t, db.DeleteUser(u.ID))
}

func TestUpdateUser(t *testing.T) {

	u, err := db.CreateUser(model.User{
		TeamID:   "foo",
		UserID:   "bar",
		UserName: "fooUser",
	})
	assert.NoError(t, err)
	assert.Equal(t, "", u.Role)

	u.Role = "admin"

	u, err = db.UpdateUser(u)
	assert.NoError(t, err)
	assert.Equal(t, "admin", u.Role)

	assert.NoError(t, db.DeleteUser(u.ID))
}
