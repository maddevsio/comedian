package storage

import (
	"github.com/maddevsio/comedian/model"
)

// CreateNotificationThread create notifications
func (m *DB) CreateNotificationThread(s model.NotificationThread) (model.NotificationThread, error) {
	res, err := m.db.Exec(
		"INSERT INTO `notification_threads` (channel_id,user_ids, notification_time, reminder_counter) VALUES (?, ?, ?, ?)",
		s.ChannelID, s.UserIDs, s.NotificationTime, s.ReminderCounter,
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
	_, err := m.db.Exec("DELETE FROM `notification_threads` WHERE id=?", id)
	return err
}

// SelectNotificationsThread returns array of notifications entries from database
func (m *DB) SelectNotificationsThread(channelID string) (model.NotificationThread, error) {
	var items model.NotificationThread
	err := m.db.Get(&items, "SELECT * FROM `notification_threads` WHERE channel_id=?", channelID)
	return items, err
}

// UpdateNotificationThread update field reminder counter
func (m *DB) UpdateNotificationThread(id int64, notificationTime int64, nonReporters string) error {
	_, err := m.db.Exec("UPDATE `notification_threads` SET user_ids=?, reminder_counter=reminder_counter+1, notification_time=? WHERE id=?", nonReporters, notificationTime, id)
	return err
}
