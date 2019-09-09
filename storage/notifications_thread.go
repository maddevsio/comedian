package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateNotificationThread create notifications
func (m *DB) CreateNotificationThread(s model.NotificationThread) (model.NotificationThread, error) {
	res, err := m.db.Exec(
		"INSERT INTO `notifications_thread` (channel_id,user_id, notification_time, reminder_counter) VALUES (?, ?, ?, ?)",
		s.ChannelID, s.UserID, s.NotificationTime, s.ReminderCounter,
	)
	if err != nil {
		return s, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return s, err
	}
	s.ID = id
	return s, nil
}

// DeleteNotificationThread deletes notification entry from database
func (m *DB) DeleteNotificationThread(id int64) error {
	_, err := m.db.Exec("DELETE FROM `notifications_thread` WHERE id=?", id)
	return err
}

// SelectNotificationsThread returns array of notifications entries from database
func (m *DB) SelectNotificationsThread(channelID string) (model.NotificationThread, error) {
	var items model.NotificationThread
	err := m.db.Get(&items, "SELECT * FROM `notifications_thread` WHERE channel_id= ?", channelID)
	return items, err
}

// UpdateNotificationThread update field reminder counter
func (m *DB) UpdateNotificationThread(id int64, channelID string, t int64) error {
	_, err := m.db.Exec("UPDATE `notifications_thread` SET reminder_counter=reminder_counter+1, notification_time=? WHERE id=? AND channel_id=?", t, id, channelID)
	return err
}
