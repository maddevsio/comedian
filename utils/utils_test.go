package utils

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSplitUser(t *testing.T) {
	user := "<@USERID|userName"
	id, name := SplitUser(user)
	assert.Equal(t, "USERID", id)
	assert.Equal(t, "userName", name)
}

func TestSecondsToHuman(t *testing.T) {
	testCases := []struct {
		output  string
		seconds int
	}{
		{"00:03", 180},
		{"00:04", 240},
		{"01:00", 3600},
		{"01:03", 3780},
		{"01:03", 3782},
	}
	for _, tt := range testCases {
		text := SecondsToHuman(tt.seconds)
		assert.Equal(t, tt.output, text)
	}

}

func TestSplitTimeTalbeCommand(t *testing.T) {
	testCases := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy on friday at 01:00", "@anatoliy", "friday", int64(1538593200), ""},
		{"@anatoliy n friday ft 01:00", "", "", int64(0), "Sorry, could not understand where are the standupers and where is the rest of the command. Please, check the text for mistakes and try again"},

		{"@anatoliy on Friday at 01:00", "@anatoliy", "friday", int64(1538593200), ""},
		{"<@UB9AE7CL9|fedorenko.tolik> on monday at 01:00", "<@UB9AE7CL9|fedorenko.tolik>", "monday", int64(1538593200), ""},
		{"@anatoliy @erik @alex on friday tuesday monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1538593200), ""},
		{"@anatoliy @erik @alex on friday, tuesday, monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1538593200), ""},
	}
	for i, tt := range testCases {
		users, weekdays, deadline, err := SplitTimeTalbeCommand(tt.command, " on ", " at ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		assert.Equal(t, tt.time, deadline)
		if err != nil {
			fmt.Println(i, err)
		}
	}

	testCasesRus := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy по пятницам в 02:04", "@anatoliy", "пятницам", int64(1538597040), ""},
		{"@anatoliy @erik @alex по понедельникам пятницам вторникам в 23:04", "@anatoliy @erik @alex", "понедельникам пятницам вторникам", int64(1538672640), ""},
	}
	for i, tt := range testCasesRus {
		users, weekdays, deadline, err := SplitTimeTalbeCommand(tt.command, " по ", " в ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		assert.Equal(t, tt.time, deadline)
		if err != nil {
			fmt.Println(i, err)
		}
	}
}

func TestFormatTime(t *testing.T) {
	testCases := []struct {
		timeString string
		hour       int
		minute     int
		err        error
	}{
		{"10:00", 10, 0, nil},
		{"11:20", 11, 20, nil},
		{"25:20", 0, 0, errors.New("time format error")},
		{"25:20:30", 0, 0, errors.New("time format error")},
		{"shit:fuck", 0, 0, errors.New("time format error")},
		{"10:fuck", 0, 0, errors.New("time format error")},
	}
	for _, tt := range testCases {
		h, m, err := FormatTime(tt.timeString)
		assert.Equal(t, tt.hour, h)
		assert.Equal(t, tt.minute, m)
		assert.Equal(t, tt.err, err)
	}
}

func TestPeriodToWeekdays(t *testing.T) {
	testCases := []struct {
		dateFrom time.Time
		dateTo   time.Time
		days     []string
		err      error
	}{
		{time.Date(2018, time.Now().Month(), 1, 1, 0, 0, 0, time.Local),
			time.Date(2018, time.Now().Month(), 3, 1, 0, 0, 0, time.Local),
			[]string{"monday", "tuesday", "wednesday"}, nil,
		},
		{time.Date(2018, time.Now().Month(), 4, 1, 0, 0, 0, time.Local),
			time.Date(2018, time.Now().Month(), 3, 1, 0, 0, 0, time.Local),
			[]string{}, errors.New("DateTo is before DateFrom"),
		},
		{time.Date(2018, time.Now().Month(), 1, 1, 0, 0, 0, time.Local),
			time.Date(2018, time.Now().Month(), 5, 1, 0, 0, 0, time.Local),
			[]string{}, errors.New("DateTo is in the future"),
		},
	}
	for _, tt := range testCases {
		days, err := PeriodToWeekdays(tt.dateFrom, tt.dateTo)
		assert.Equal(t, tt.err, err)
		for i, day := range tt.days {
			assert.Equal(t, day, days[i])
		}
	}
}
