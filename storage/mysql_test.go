package storage

import (
	"github.com/maddevsio/comedian/config"
	"log"
	"time"
)

var db = setupDB()

func setupDB() *DB {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	db, err := New(c)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	return db
}
