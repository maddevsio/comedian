package api

import (
	"github.com/maddevsio/comedian/storage"
	//"github.com/sirupsen/logrus"
)

type (
	// API struct
	API struct {
		db *storage.Storage
	}
)
