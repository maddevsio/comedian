package storage

import "github.com/maddevsio/comedian/model"

type (
	// Storage is interface for all supported storages(e.g. MySQL, Postgresql)
	Storage interface {
		// CreateStandup creates standup entry in database
		CreateStandup(model.Standup) (model.Standup, error)
		// UpdateStandup updates standup entry in database
		UpdateStandup(model.Standup) (model.Standup, error)
		// SelectStandup selects standup entry from database
		SelectStandup(int64) (model.Standup, error)
		// DeleteStandup deletes standup entry from database
		DeleteStandup(int64) error
		// ListStandups returns array of standup entries from database
		ListStandups() ([]model.Standup, error)
	}
)
