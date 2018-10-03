package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

//SplitUser divides full user object to name & id
func SplitUser(user string) (string, string) {
	userFull := strings.Split(user, "|")
	userID := strings.Replace(userFull[0], "<@", "", -1)
	userName := strings.Replace(userFull[1], ">", "", -1)
	return userID, userName
}

//SecondsToHuman converts seconds (int) to HH:MM format
func SecondsToHuman(input int) string {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	return fmt.Sprintf("%02d:%02d", int(hours), int(minutes))
}

// FormatTime returns hour and minutes from string
func FormatTime(t string) (hour, min int, err error) {
	newErr := errors.New("time format error")
	ts := strings.Split(t, ":")
	if len(ts) != 2 {
		return 0, 0, newErr
	}

	hour, err = strconv.Atoi(ts[0])
	if err != nil {
		return 0, 0, newErr
	}
	min, err = strconv.Atoi(ts[1])
	if err != nil {
		return 0, 0, newErr
	}

	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return 0, 0, newErr
	}
	return hour, min, nil
}

//SplitTimeTalbeCommand returns set of substrings
func SplitTimeTalbeCommand(t, on, at string) (string, string, string) {
	a := strings.Split(t, on)
	b := strings.Split(a[1], at)
	return strings.TrimSpace(a[0]), strings.TrimSpace(b[0]), strings.TrimSpace(b[1])
}
