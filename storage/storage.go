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
		// CreateStandupUser creates standupUser entry in database
		CreateStandupUser(model.StandupUser) (model.StandupUser, error)
		// DeleteStandupUser deletes standup_users entry from database
		DeleteStandupUserByUsername(string, string) error
		// ListStandupUsers returns array of standupUser entries from database
		ListStandupUsers(string) ([]model.StandupUser, error)
		// CreateStandupTime creates time entry in database
		CreateStandupTime(model.StandupTime) (model.StandupTime, error)
	}
)
