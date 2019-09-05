package storage

import (
	"fmt"
	"io/ioutil"

	"github.com/pressly/goose"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// DB provides api for work with DB database
type DB struct {
	db *sqlx.DB
}

// New creates a new instance of database API
func New(dbConn, migrationsPathh string) (*DB, error) {
	conn, err := sqlx.Connect("mysql", dbConn)
	if err != nil {
		conn, err = sqlx.Connect("mysql", "comedian:comedian@tcp(localhost:3306)/comedian?parseTime=true")
		if err != nil {
			return nil, err
		}
	}
	db := &DB{conn}

	goose.SetDialect("mysql")

	current, err := goose.EnsureDBVersion(conn.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to EnsureDBVersion: %v", err)
	}

	files, err := ioutil.ReadDir(migrationsPathh)
	if err != nil {
		return nil, err
	}

	migrations, err := goose.CollectMigrations(migrationsPathh, current, int64(len(files)))
	if err != nil {
		return nil, err
	}

	for _, m := range migrations {
		err := m.Up(conn.DB)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
