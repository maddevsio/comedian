package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/bouk/monkey"

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
		{"0:03", 180},
		{"0:04", 240},
		{"1:00", 3600},
		{"1:03", 3780},
		{"1:03", 3782},
		{"4:10", 15000},
		{"12:30", 45000},
	}
	for _, tt := range testCases {
		text := SecondsToHuman(tt.seconds)
		assert.Equal(t, tt.output, text)
	}

}

func TestSplitTimeTalbeCommand(t *testing.T) {
	d := time.Date(2018, 1, 2, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy on friday at 01:00", "@anatoliy", "friday", int64(1514833200), ""},
		{"@anatoliy n friday ft 01:00", "", "", int64(0), "Sorry, could not understand where are the standupers and where is the rest of the command. Please, check the text for mistakes and try again"},
		{"@anatoliy on Friday at 01:00", "@anatoliy", "friday", int64(1514833200), ""},
		{"<@UB9AE7CL9|fedorenko.tolik> on monday at 01:00", "<@UB9AE7CL9|fedorenko.tolik>", "monday", int64(1514833200), ""},
		{"@anatoliy @erik @alex on friday tuesday monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1514833200), ""},
		{"@anatoliy @erik @alex on friday, tuesday, monday wednesday at 01:00", "@anatoliy @erik @alex", "friday tuesday monday wednesday", int64(1514833200), ""},
	}
	for _, tt := range testCases {
		users, weekdays, _, err := SplitTimeTalbeCommand(tt.command, " on ", " at ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		//assert.Equal(t, tt.time, deadline)
		if err != nil {
			assert.Equal(t, errors.New(tt.err), err)
		}
	}

	testCasesRus := []struct {
		command  string
		users    string
		weekdays string
		time     int64
		err      string
	}{
		{"@anatoliy по пятницам в 02:04", "@anatoliy", "пятницам", int64(1514837040), ""},
		{"@anatoliy @erik @alex по понедельникам пятницам вторникам в 23:04", "@anatoliy @erik @alex", "понедельникам пятницам вторникам", int64(1514912640), ""},
	}
	for _, tt := range testCasesRus {
		users, weekdays, _, err := SplitTimeTalbeCommand(tt.command, " по ", " в ")
		assert.Equal(t, tt.users, users)
		assert.Equal(t, tt.weekdays, weekdays)
		//assert.Equal(t, tt.time, deadline)
		if err != nil {
			assert.Equal(t, errors.New(tt.err), err)
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

	d := time.Date(2018, 10, 4, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

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

func TestParseTimeTextToInt(t *testing.T) {

	d := time.Date(2018, 10, 4, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	testCases := []struct {
		timeText string
		time     int64
		err      error
	}{
		{"0", 0, nil},
		{"10:00", 1538625600, nil},
		{"xx:00", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
		{"00:xx", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
		{"00:62", 0, errors.New("Wrong time! Please, check the time format and try again!")},
		{"10am", 0, errors.New("Seems like you used short time format, please, use 24:00 hour format instead!")},
		{"20", 0, errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")},
	}
	for _, tt := range testCases {
		_, err := ParseTimeTextToInt(tt.timeText)
		assert.Equal(t, tt.err, err)
		//assert.Equal(t, tt.time, time)
	}
}
