package storage

import (
	"time"

	"github.com/maddevsio/comedian/model"
)

// Storage is interface for all supported storages(e.g. MySQL, Postgresql)
type Storage interface {
	// CreateStandup creates standup entry in database
	CreateStandup(model.Standup) (model.Standup, error)

	// UpdateStandup updates standup entry in database
	UpdateStandup(model.Standup) (model.Standup, error)

	// SelectStandupByMessageTS selects standup entry by messageTS from database
	SelectStandupByMessageTS(string) (model.Standup, error)

	// SelectStandupsByChannelID selects standup entry by channel ID and time period from database
	SelectStandupsByChannelIDForPeriod(string, time.Time, time.Time) ([]model.Standup, error)

	SelectStandupsFiltered(string, string, time.Time, time.Time) ([]model.Standup, error)

	// DeleteStandup deletes standup entry from database
	DeleteStandup(int64) error

	// ListStandups returns array of standup entries from database
	ListStandups() ([]model.Standup, error)

	// CreateStandupUser creates standupUser entry in database
	CreateStandupUser(model.StandupUser) (model.StandupUser, error)

	// Checks if user has admin role
	IsAdmin(string, string) bool

	//FindStandupUser finds standup user
	FindStandupUser(username string) (model.StandupUser, error)

	//FindStandupUserInChannelByUserID finds standup user in channel by Slack member ID
	FindStandupUserInChannelByUserID(string, string) (model.StandupUser, error)

	//GetNonReporters returns a list of non reporters
	GetNonReporters(string, time.Time, time.Time) ([]model.StandupUser, error)

	//IsNonReporter checks if a user is non reporter
	IsNonReporter(string, string, time.Time, time.Time) (bool, error)

	HasExistedAlready(string, string, time.Time) (bool, error)

	// DeleteStandupUser deletes standup_users entry from database
	DeleteStandupUser(string, string) error

	// DeleteAdmin deletes standup_users entry from database
	DeleteAdmin(string, string) error

	// ListStandupUsersByChannelID returns array of standupUser entries from database
	ListStandupUsersByChannelID(string) ([]model.StandupUser, error)

	// ListAdminsByChannelID returns array of standup entries from database
	ListAdminsByChannelID(string) ([]model.StandupUser, error)

	// ListAllStandupUsers returns array of standupUser entries from database
	ListAllStandupUsers() ([]model.StandupUser, error)

	// CreateStandupTime creates standup time entry in database
	CreateStandupTime(model.StandupTime) (model.StandupTime, error)

	// DeleteStandupTime deletes time entry from database
	DeleteStandupTime(string) error

	// ListStandupTime returns standup time entry from database
	GetChannelStandupTime(string) (model.StandupTime, error)

	// ListAllStandupTime returns standup time entry for all channels from database
	ListAllStandupTime() ([]model.StandupTime, error)

	//GetAllChannels returns a list of all channels
	GetAllChannels() ([]string, error)

	//GetUserChannels returns a list of user's channels
	GetUserChannels(string) ([]string, error)
}
