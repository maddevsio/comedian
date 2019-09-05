package storage

import (
	"log"
	"time"

	"github.com/maddevsio/comedian/config"
)

var db = setupDB()

func setupDB() *DB {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	db, err := New(c.DatabaseURL, "../migrations")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	return db
}
