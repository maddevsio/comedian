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

	// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
	SelectStandupByMessageTS(string) (model.Standup, error)

	// SelectStandupsByChannelIDForPeriod selects standup entrys by channel ID and time period from database
	SelectStandupsByChannelIDForPeriod(string, time.Time, time.Time) ([]model.Standup, error)

	// SelectStandupsFiltered selects standup entrys by channel ID and time period from database
	SelectStandupsFiltered(string, string, time.Time, time.Time) (model.Standup, error)

	// DeleteStandup deletes standup entry from database
	DeleteStandup(int64) error

	// CreateChannelMember creates comedian entry in database
	CreateChannelMember(model.ChannelMember) (model.ChannelMember, error)

	//FindChannelMemberByUserID finds user in channel
	FindChannelMemberByUserID(string, string) (model.ChannelMember, error)

	//SelectChannelMember finds user in channel
	SelectChannelMember(int64) (model.ChannelMember, error)

	//FindChannelMemberByUserName finds user in channel
	FindChannelMemberByUserName(string, string) (model.ChannelMember, error)

	// ListAllChannelMembers returns array of standup entries from database
	ListAllChannelMembers() ([]model.ChannelMember, error)

	//GetNonReporters returns a list of non reporters in selected time period
	GetNonReporters(string, time.Time, time.Time) ([]model.ChannelMember, error)

	// IsNonReporter returns true if user did not submit standup in time period, false othervise
	IsNonReporter(string, string, time.Time, time.Time) (bool, error)

	ListChannelMembers(string) ([]model.ChannelMember, error)

	// DeleteChannelMember deletes channel_members entry from database
	DeleteChannelMember(string, string) error

	// CreateStandupTime creates time entry in database
	CreateStandupTime(int64, string) error

	// UpdateChannelStandupTime updates time entry in database
	UpdateChannelStandupTime(int64, string) error

	// GetChannelStandupTime returns standup time entry from database
	GetChannelStandupTime(string) (int64, error)

	// ListAllStandupTime returns standup time entry for all channels from database
	ListAllStandupTime() ([]int64, error)

	// DeleteStandupTime deletes channels entry for channel from database
	DeleteStandupTime(string) error

	// AddToStandupHistory creates backup standup entry in standup_edit_history database
	AddToStandupHistory(model.StandupEditHistory) (model.StandupEditHistory, error)

	//GetAllChannels returns list of unique channels
	GetAllChannels() ([]string, error)

	//GetUserChannels returns list of user's channels
	GetUserChannels(string) ([]string, error)

	//GetChannelName returns channel name
	GetChannelName(string) (string, error)

	//GetChannelID returns channel name
	GetChannelID(string) (string, error)

	// ListStandups returns array of standup entries from database
	// Helper function for testing
	ListStandups() ([]model.Standup, error)

	// CreateChannel creates standup entry in database
	CreateChannel(model.Channel) (model.Channel, error)

	// SelectChannel selects Channel entry from database
	SelectChannel(string) (model.Channel, error)

	// GetChannels selects Channel entry from database
	GetChannels() ([]model.Channel, error)

	// DeleteChannel deletes Channel entry from database
	DeleteChannel(int64) error

	// CreateUser creates standup entry in database
	CreateUser(model.User) (model.User, error)

	// UpdateUser updates User entry in database
	UpdateUser(model.User) (model.User, error)

	// SelectUser selects User entry from database
	SelectUser(string) (model.User, error)

	// SelectUserByUserName selects User entry from database
	SelectUserByUserName(string) (model.User, error)

	// DeleteUser deletes User entry from database
	DeleteUser(int64) error

	// ListAdmins selects User entry from database
	ListAdmins() ([]model.User, error)

	//SubmittedStandupToday shows if a user submitted standup today
	SubmittedStandupToday(string, string) bool

	// CreatePM creates comedian entry in database
	CreatePM(model.ChannelMember) (model.ChannelMember, error)

	// UserIsPMForProject returns true if user is a project's PM.
	UserIsPMForProject(string, string) bool

	// CreateUser creates standup entry in database
	CreateTimeTable(model.TimeTable) (model.TimeTable, error)

	// UpdateTimeTable updates TimeTable entry in database
	UpdateTimeTable(model.TimeTable) (model.TimeTable, error)

	// SelectTimeTable selects TimeTable entry from database
	SelectTimeTable(int64) (model.TimeTable, error)

	// DeleteTimeTable deletes TimeTable entry from database
	DeleteTimeTable(int64) error

	//ListStandupersWithTimeTablesForToday returns list of chan members who has timetables
	ListTimeTablesForDay(string) ([]model.TimeTable, error)

	//MemberHasTimeTable returns true if member has timetable
	MemberHasTimeTable(int64) bool

	//MemberShouldBeTracked returns true if member has timetable
	MemberShouldBeTracked(int64, time.Time, time.Time) bool
}
