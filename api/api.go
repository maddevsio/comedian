package api

import "github.com/maddevsio/comedian/storage"

type (
	// API struct
	API struct {
		db *storage.Storage
	}
)
