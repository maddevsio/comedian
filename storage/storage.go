package storage

import (
	"time"

	"github.com/maddevsio/comedian/model"
)

type (
	// Storage is interface for all supported storages(e.g. MySQL, Postgresql)
	Storage interface {
		// CreateStandup creates standup entry in database
		CreateStandup(model.Standup) (model.Standup, error)
		// UpdateStandup updates standup entry in database
		UpdateStandup(model.Standup) (model.Standup, error)
		// SelectStandup selects standup entry from database
		SelectStandup(int64) (model.Standup, error)
		// SelectStandupByMessageTS selects standup entry by messageTS from database
		SelectStandupByMessageTS(string) (model.Standup, error)
		// SelectStandupsByChannelID selects standup entry by channel ID from database
		SelectStandupsByChannelID(string) ([]model.Standup, error)
		// SelectStandupsByChannelID selects standup entry by channel name and time period from database
		SelectStandupByChannelNameForPeriod(string, time.Time, time.Time) ([]model.Standup, error)
		// SelectStandupsByChannelID selects standup entry by channel ID and time period from database
		SelectStandupsByChannelIDForPeriod(string, time.Time, time.Time) ([]model.Standup, error)
		// SelectStandupsForPeriod selects standup entrys for time period from database
		SelectStandupsForPeriod(dateStart, dateEnd time.Time) ([]model.Standup, error)
		// DeleteStandup deletes standup entry from database
		DeleteStandup(int64) error
		// ListStandups returns array of standup entries from database
		ListStandups() ([]model.Standup, error)
		// CreateStandupUser creates standupUser entry in database
		CreateStandupUser(model.StandupUser) (model.StandupUser, error)
		//FindStandupUser finds standup user
		FindStandupUser(username string) (model.StandupUser, error)
		//FindStandupUserInChannel finds standup user in channel
		FindStandupUserInChannel(string, string) (model.StandupUser, error)
		//FindStandupUserInChannelName finds standup user in channel
		FindStandupUserInChannelName(string, string) (model.StandupUser, error)
		// DeleteStandupUser deletes standup_users entry from database
		DeleteStandupUserByUsername(string, string) error
		// ListStandupUsersByChannelID returns array of standupUser entries from database
		ListStandupUsersByChannelID(string) ([]model.StandupUser, error)
		// ListAllStandupUsers returns array of standupUser entries from database
		ListAllStandupUsers() ([]model.StandupUser, error)
		// ListStandupUsers returns array of standupUser entries from database filtered by channel name
		ListStandupUsersByChannelName(string) ([]model.StandupUser, error)
		// CreateStandupTime creates standup time entry in database
		CreateStandupTime(model.StandupTime) (model.StandupTime, error)
		// DeleteStandupTime deletes time entry from database
		DeleteStandupTime(string) error
		// ListStandupTime returns standup time entry from database
		ListStandupTime(string) (model.StandupTime, error)
		// ListAllStandupTime returns standup time entry for all channels from database
		ListAllStandupTime() ([]model.StandupTime, error)
	}
)
