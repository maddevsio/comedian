package utils

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	return fmt.Sprintf("%v:%02d", int(hours), int(minutes))
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
func SplitTimeTalbeCommand(t, on, at string) (string, string, int64, error) {

	a := strings.Split(t, on)
	if len(a) != 2 {
		return "", "", int64(0), errors.New("Sorry, could not understand where are the standupers and where is the rest of the command. Please, check the text for mistakes and try again")
	}
	users := strings.TrimSpace(a[0])
	b := strings.Split(a[1], at)
	if len(a) != 2 {
		return "", "", int64(0), errors.New("Sorry, could not understand where are the weekdays and where is the time. Please, check the text for mistakes and try again")
	}
	weekdays := strings.ToLower(strings.TrimSpace(b[0]))
	timeText := strings.ToLower(strings.TrimSpace(b[1]))
	time, err := ParseTimeTextToInt(timeText)
	if err != nil {
		return "", "", int64(0), err
	}
	reg, err := regexp.Compile("[^а-яА-Яa-zA-Z0-9]+")
	if err != nil {
		return "", "", int64(0), err
	}
	weekdays = reg.ReplaceAllString(weekdays, " ")
	return users, weekdays, time, nil
}

func ParseTimeTextToInt(timeText string) (int64, error) {
	if timeText == "0" {
		return int64(0), nil
	}
	matchHourMinuteFormat, _ := regexp.MatchString("[0-9][0-9]:[0-9][0-9]", timeText)
	matchAMPMFormat, _ := regexp.MatchString("[0-9][0-9][a-z]", timeText)

	if matchHourMinuteFormat {
		t := strings.Split(timeText, ":")
		hours, _ := strconv.Atoi(t[0])
		munites, _ := strconv.Atoi(t[1])

		if hours > 23 || munites > 59 {
			return int64(0), errors.New("Wrong time! Please, check the time format and try again!")
		}
		return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hours, munites, 0, 0, time.Local).Unix(), nil
	}

	if matchAMPMFormat {
		return int64(0), errors.New("Seems like you used short time format, please, use 24:00 hour format instead!")
	}

	return int64(0), errors.New("Could not understand how you mention time. Please, use 24:00 hour format and try again!")

}
