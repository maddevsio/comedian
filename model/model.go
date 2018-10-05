package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/maddevsio/comedian/config"
)

type (
	// Standup model used for serialization/deserialization stored standups
	Standup struct {
		ID        int64     `db:"id" json:"id"`
		Created   time.Time `db:"created" json:"created"`
		Modified  time.Time `db:"modified" json:"modified"`
		ChannelID string    `db:"channel_id" json:"channelId"`
		UserID    string    `db:"user_id" json:"userId"`
		Comment   string    `db:"comment" json:"comment"`
		MessageTS string    `db:"message_ts" json:"message_ts"`
	}

	// User model used for serialization/deserialization stored Users
	User struct {
		ID       int64  `db:"id" json:"id"`
		UserName string `db:"user_name" json:"user_name"`
		UserID   string `db:"user_id" json:"user_id"`
		Role     string `db:"role" json:"role"`
	}

	// Channel model used for serialization/deserialization stored Channels
	Channel struct {
		ID          int64  `db:"id" json:"id"`
		ChannelName string `db:"channel_name" json:"channel_name"`
		ChannelID   string `db:"channel_id" json:"channel_id"`
		StandupTime int64  `db:"channel_standup_time" json:"time"`
	}

	// ChannelMember model used for serialization/deserialization stored ChannelMembers
	ChannelMember struct {
		ID            int64     `db:"id" json:"id"`
		UserID        string    `db:"user_id" json:"user_id"`
		ChannelID     string    `db:"channel_id" json:"channel_id"`
		RoleInChannel string    `db:"role_in_channel" json:"role_in_channel"`
		StandupTime   int64     `db:"standup_time" json:"time"`
		Created       time.Time `db:"created" json:"created"`
	}

	// TimeTable model used for serialization/deserialization stored timetables
	TimeTable struct {
		ID              int64     `db:"id" json:"id"`
		ChannelMemberID int64     `db:"channel_member_id" json:"channel_member_id"`
		Created         time.Time `db:"created" json:"created"`
		Modified        time.Time `db:"modified" json:"modified"`
		Monday          int64     `db:"monday" json:"monday"`
		Tuesday         int64     `db:"tuesday" json:"tuesday"`
		Wednesday       int64     `db:"wednesday" json:"wednesday"`
		Thursday        int64     `db:"thursday" json:"thursday"`
		Friday          int64     `db:"friday" json:"friday"`
		Saturday        int64     `db:"saturday" json:"saturday"`
		Sunday          int64     `db:"sunday" json:"sunday"`
	}

	// StandupEditHistory model used for serialization/deserialization stored standup edit history
	StandupEditHistory struct {
		ID          int64     `db:"id" json:"id"`
		Created     time.Time `db:"created" json:"created"`
		StandupID   int64     `db:"standup_id" json:"standupId"`
		StandupText string    `db:"standuptext" json:"standuptext"`
	}
)

// Validate validates Standup struct
func (c Standup) Validate() error {
	if c.UserID == "" {
		err := errors.New("User cannot be empty")
		return err
	}
	if c.ChannelID == "" {
		err := errors.New("Channel cannot be empty")
		return err
	}
	return nil
}

// Validate validates StandupUser struct
func (c ChannelMember) Validate() error {
	if c.UserID == "" && c.ChannelID == "" {
		err := errors.New("User/Channel cannot be empty")
		return err
	}
	return nil
}

// Validate validates StandupTimeHistory struct
func (c StandupEditHistory) Validate() error {
	if c.StandupText == "" {
		err := errors.New("Text cannot be empty")
		return err
	}
	return nil
}

//IsAdmin returns user status
func (u User) IsAdmin() bool {
	if u.Role == "admin" {
		return true
	}
	return false
}

//Show shows timetable
func (tt TimeTable) Show() string {
	c, _ := config.Get()
	timeTableString := ""
	if tt.Monday != 0 {
		monday := time.Unix(tt.Monday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowMonday, monday.Hour(), monday.Minute())
	}
	if tt.Tuesday != 0 {
		tuesday := time.Unix(tt.Tuesday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowTuesday, tuesday.Hour(), tuesday.Minute())
	}
	if tt.Wednesday != 0 {
		wednesday := time.Unix(tt.Wednesday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowWednesday, wednesday.Hour(), wednesday.Minute())
	}
	if tt.Thursday != 0 {
		thursday := time.Unix(tt.Thursday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowThursday, thursday.Hour(), thursday.Minute())
	}
	if tt.Friday != 0 {
		friday := time.Unix(tt.Friday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowFriday, friday.Hour(), friday.Minute())
	}
	if tt.Saturday != 0 {
		saturday := time.Unix(tt.Saturday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowSaturday, saturday.Hour(), saturday.Minute())
	}
	if tt.Sunday != 0 {
		sunday := time.Unix(tt.Sunday, 0)
		timeTableString += fmt.Sprintf(c.Translate.TimetableShowSunday, sunday.Hour(), sunday.Minute())
	}

	if timeTableString == "" {
		return c.Translate.EmptyTimetable
	} else {
		timeTableString += "|"
	}

	return timeTableString
}

func (tt TimeTable) ShowDeadlineOn(day string) int64 {
	switch day {
	case "monday":
		return tt.Monday
	case "tuesday":
		return tt.Tuesday
	case "wednesday":
		return tt.Wednesday
	case "thursday":
		return tt.Thursday
	case "friday":
		return tt.Friday
	case "saturday":
		return tt.Saturday
	case "sunday":
		return tt.Sunday
	default:
		return int64(0)
	}
}

//IsEmpty shows if timetable is empty
func (tt TimeTable) IsEmpty() bool {
	if tt.Monday == 0 && tt.Tuesday == 0 && tt.Wednesday == 0 && tt.Thursday == 0 && tt.Friday == 0 && tt.Saturday == 0 && tt.Sunday == 0 {
		return true
	}
	return false
}
