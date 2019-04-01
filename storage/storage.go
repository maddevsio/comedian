package storage

import "gitlab.com/team-monitoring/comedian/model"

type Storage interface {
	//CreateBotSettings creates bot properties for the newly created bot
	CreateBotSettings(token, userID, teamID, teamName string) (model.BotSettings, error)

	//GetAllBotSettings returns all bots
	GetAllBotSettings() ([]model.BotSettings, error)

	//GetBotSettingsByTeamName returns a particular bot
	// When dashboard is ready - DELETE!
	GetBotSettingsByTeamName(teamName string) (model.BotSettings, error)

	//GetBotSettings returns a particular bot
	GetBotSettings(id int64) (model.BotSettings, error)

	//UpdateBotSettings updates bot
	UpdateBotSettings(settings model.BotSettings) (model.BotSettings, error)

	//DeleteBotSettingsByID deletes bot
	DeleteBotSettingsByID(id int64) error

	//DeleteBotSettings deletes bot
	DeleteBotSettings(teamID string) error

	// CreateChannel creates standup entry in database
	CreateChannel(ch model.Channel) (model.Channel, error)

	// UpdateChannel updates Channel entry in database
	UpdateChannel(ch model.Channel) (model.Channel, error)

	//ListChannels returns list of channels
	ListChannels() ([]model.Channel, error)

	// SelectChannel selects Channel entry from database
	SelectChannel(channelID string) (model.Channel, error)

	// GetChannel selects Channel entry from database with specific id
	GetChannel(id int64) (model.Channel, error)

	// DeleteChannel deletes Channel entry from database
	DeleteChannel(id int64) error

	// CreateStanduper creates comedian entry in database
	CreateStanduper(s model.Standuper) (model.Standuper, error)

	// UpdateStanduper updates Standuper entry in database
	UpdateStanduper(st model.Standuper) (model.Standuper, error)

	//FindStansuperByUserID finds user in channel
	FindStansuperByUserID(userID, channelID string) (model.Standuper, error)

	// ListStandupers returns array of standup entries from database
	ListStandupers() ([]model.Standuper, error)

	//GetStanduper returns a standuper
	GetStanduper(id int64) (model.Standuper, error)

	// ListChannelStandupers returns array of standup entries from database
	ListChannelStandupers(channelID string) ([]model.Standuper, error)

	// DeleteStanduper deletes channel_members entry from database
	DeleteStanduper(id int64) error

	// CreateStandup creates standup entry in database
	CreateStandup(s model.Standup) (model.Standup, error)

	// UpdateStandup updates standup entry in database
	UpdateStandup(s model.Standup) (model.Standup, error)

	GetStandup(id int64) (model.Standup, error)

	// SelectStandupByMessageTS selects standup entry from database filtered by MessageTS parameter
	SelectStandupByMessageTS(messageTS string) (model.Standup, error)

	// ListStandups returns array of standup entries from database
	ListStandups() ([]model.Standup, error)

	// DeleteStandup deletes standup entry from database
	DeleteStandup(id int64) error

	// CreateUser creates standup entry in database
	CreateUser(u model.User) (model.User, error)
	// SelectUser selects User entry from database
	SelectUser(userID string) (model.User, error)
	// GetUser selects User entry from database
	GetUser(id int64) (model.User, error)

	// UpdateUser updates User entry in database
	UpdateUser(u model.User) (model.User, error)

	// ListUsers selects Users from database
	ListUsers() ([]model.User, error)

	// DeleteUser deletes User entry from database
	DeleteUser(id int64) error
}
