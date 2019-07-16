package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateUser creates standup entry in database
func (m *DB) CreateUser(u model.User) (model.User, error) {
	err := u.Validate()
	if err != nil {
		return u, err
	}
	res, err := m.db.Exec(
		"INSERT INTO `users` (team_id, user_name, user_id, real_name, tz, tz_offset, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		u.TeamID, u.UserName, u.UserID, u.RealName, u.TZ, u.TZOffset, u.Status,
	)
	if err != nil {
		return u, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return u, err
	}
	u.ID = id

	return u, nil
}

// UpdateUser updates User entry in database
func (m *DB) UpdateUser(u model.User) (model.User, error) {
	err := u.Validate()
	if err != nil {
		return u, err
	}
	_, err = m.db.Exec(
		"UPDATE `users` SET real_name=?, tz=?, tz_offset=?, status=? WHERE id=?",
		u.RealName, u.TZ, u.TZOffset, u.Status, u.ID,
	)
	if err != nil {
		return u, err
	}
	var i model.User
	err = m.db.Get(&i, "SELECT * FROM `users` WHERE id=?", u.ID)
	return i, err
}

// SelectUser selects User entry from database
func (m *DB) SelectUser(userID string) (model.User, error) {
	var c model.User
	err := m.db.Get(&c, "SELECT * FROM `users` WHERE user_id=?", userID)
	if err != nil {
		return c, err
	}
	return c, err
}

// ListUsers selects Users from database
func (m *DB) ListUsers() ([]model.User, error) {
	var u []model.User
	err := m.db.Select(&u, "SELECT * FROM `users`")
	if err != nil {
		return u, err
	}
	return u, err
}

// DeleteUser deletes User entry from database
func (m *DB) DeleteUser(id int64) error {
	_, err := m.db.Exec("DELETE FROM `users` WHERE id=?", id)
	return err
}
