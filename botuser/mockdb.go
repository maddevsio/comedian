package botuser

import (
	"gitlab.com/team-monitoring/comedian/model"
	"gitlab.com/team-monitoring/comedian/storage"
)

type MockedDB struct {
	storage.Storage

	CreatedUser       model.User
	CreatedUserError  error
	SelectedUser      model.User
	SelectedUserError error
	UpdatedUser       model.User
	UpdatedUserError  error
	ListedUsers       []model.User
	ListedUsersError  error
	DeleteUserError   error

	CreatedChannel       model.Channel
	CreatedChannelError  error
	SelectedChannel      model.Channel
	SelectedChannelError error
	UpdatedChannel       model.Channel
	UpdatedChannelError  error

	CreatedStandup                model.Standup
	CreateStandupError            error
	UpdatedStandup                model.Standup
	UpdateStandupError            error
	SelectedStandupByMessageTS    model.Standup
	SelectStandupByMessageTSError error
	DeleteStandupError            error

	CreatedStanduper           model.Standuper
	CreateStanduperError       error
	UpdatedStanduper           model.Standuper
	UpdateStanduperError       error
	FoundStanduper             model.Standuper
	FoundStanduperError        error
	ChannelStandupers          []model.Standuper
	ListChannelStandupersError error
	Standupers                 []model.Standuper
	ListStandupersError        error
	DeleteStanduperError       error

	SubmittedStandupTodayResult bool
	SubmittedStandupTodayError  error
}

func (m MockedDB) CreateUser(model.User) (model.User, error) {
	return m.CreatedUser, m.CreatedUserError
}

func (m MockedDB) ListUsers() ([]model.User, error) {
	return m.ListedUsers, m.ListedUsersError
}

func (m MockedDB) UpdateUser(model.User) (model.User, error) {
	return m.UpdatedUser, m.UpdatedUserError
}

func (m MockedDB) SelectUser(string) (model.User, error) {
	return m.SelectedUser, m.SelectedUserError
}

func (m MockedDB) DeleteUser(int64) error {
	return m.DeleteUserError
}

func (m MockedDB) CreateChannel(model.Channel) (model.Channel, error) {
	return m.CreatedChannel, m.CreatedChannelError
}

func (m MockedDB) SelectChannel(string) (model.Channel, error) {
	return m.SelectedChannel, m.SelectedChannelError
}

func (m MockedDB) UpdateChannel(model.Channel) (model.Channel, error) {
	return m.UpdatedChannel, m.UpdatedChannelError
}

func (m MockedDB) CreateStandup(model.Standup) (model.Standup, error) {
	return m.CreatedStandup, m.CreateStandupError
}

func (m MockedDB) UpdateStandup(model.Standup) (model.Standup, error) {
	return m.UpdatedStandup, m.UpdateStandupError
}

func (m MockedDB) SelectStandupByMessageTS(string) (model.Standup, error) {
	return m.SelectedStandupByMessageTS, m.SelectStandupByMessageTSError
}

func (m MockedDB) DeleteStandup(int64) error {
	return m.DeleteStandupError
}

func (m MockedDB) CreateStanduper(model.Standuper) (model.Standuper, error) {
	return m.CreatedStanduper, m.CreateStanduperError
}

func (m MockedDB) UpdateStanduper(model.Standuper) (model.Standuper, error) {
	return m.UpdatedStanduper, m.UpdateStanduperError
}

func (m MockedDB) FindStansuperByUserID(string, string) (model.Standuper, error) {
	return m.FoundStanduper, m.FoundStanduperError
}

func (m MockedDB) ListStandupers() ([]model.Standuper, error) {
	return m.Standupers, m.ListStandupersError
}

func (m MockedDB) ListChannelStandupers(string) ([]model.Standuper, error) {
	return m.ChannelStandupers, m.ListChannelStandupersError
}

func (m MockedDB) DeleteStanduper(int64) error {
	return m.DeleteStanduperError
}

func (m MockedDB) UserSubmittedStandupToday(string, string) (bool, error) {
	return m.SubmittedStandupTodayResult, m.SubmittedStandupTodayError
}
