package storage

import (
	// This line is must for working MySQL database
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/team-monitoring/comedian/model"
)

func (m *DB) UserSubmittedStandupToday(channel, user string) (bool, error) {
	var s model.Standup
	err := m.DB.Get(&s, "SELECT * FROM `standups` WHERE user_id=? AND channel_id=? AND message_ts BETWEEN ? and ?", user, channel, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC).Unix(), time.Now().Unix())
	if err != nil {
		return false, err
	}
	return true, nil
}
