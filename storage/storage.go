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

	// CreateChannelMember creates ChannelMember entry in database
	CreateChannelMember(model.ChannelMember) (model.ChannelMember, error)

	//FindChannelMemberByUserID finds standup user in channel by Slack member ID
	FindChannelMemberByUserID(string, string) (model.ChannelMember, error)

	//GetNonReporters returns a list of non reporters
	GetNonReporters(string, time.Time, time.Time) ([]model.ChannelMember, error)

	//IsNonReporter checks if a user is non reporter
	IsNonReporter(string, string, time.Time, time.Time) (bool, error)

	HasExistedAlready(string, string, time.Time) (bool, error)

	CheckIfUserExist(string) (bool, error)

	// DeleteChannelMember deletes standup_users entry from database
	DeleteChannelMember(string, string) error

	// DeleteAdmin deletes standup_users entry from database
	DeleteAdmin(string, string) error

	// ListChannelMembers returns array of ChannelMember entries from database
	ListChannelMembers(string) ([]model.ChannelMember, error)

	// ListAdminsByChannelID returns array of standup entries from database
	ListAdminsByChannelID(string) ([]model.ChannelMember, error)

	// ListAllChannelMembers returns array of ChannelMember entries from database
	ListAllChannelMembers() ([]model.ChannelMember, error)

	// CreateStandupTime creates standup time entry in database
	CreateStandupTime(int64, string) error

	// UpdateStandupTime creates standup time entry in database
	UpdateStandupTime(int64, string) error

	// DeleteStandupTime deletes time entry from database
	DeleteStandupTime(string) error

	// ListStandupTime returns standup time entry from database
	GetChannelStandupTime(string) (int64, error)

	// ListAllStandupTime returns standup time entry for all channels from database
	ListAllStandupTime() ([]int64, error)

	//GetAllChannels returns a list of all channels
	GetAllChannels() ([]string, error)

	//GetUserChannels returns a list of user's channels
	GetUserChannels(string) ([]string, error)

	//GetChannelName returns channel name
	GetChannelName(string) (string, error)

	//GetChannelID returns channel name
	GetChannelID(string) (string, error)

	//FindChannelMemberByUserName finds user in channel
	FindChannelMemberByUserName(string) (model.ChannelMember, error)
}
