package storage

import "github.com/maddevsio/comedian/model"

type (
	// Storage is interface for all supported storages(e.g. MySQL, Postgresql)
	Storage interface {
		CreateStandup(model.Standup) (model.Standup, error)
		UpdateStandup(model.Standup) (model.Standup, error)
		SelectStandup(int64) (model.Standup, error)
		DeleteStandup(int64) error
		ListStandups() ([]model.Standup, error)
	}
)
