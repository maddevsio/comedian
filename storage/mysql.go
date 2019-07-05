package storage

import (
	"time"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/maddevsio/comedian/config"
)

// DB provides api for work with DB database
type DB struct {
	db *sqlx.DB
}

// New creates a new instance of database API
func New(c *config.Config) (*DB, error) {
	con, err := sqlx.Connect("mysql", c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	con.SetConnMaxLifetime(time.Second)
	db := &DB{con}

	return db, nil
}
