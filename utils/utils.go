package utils

import (
	"fmt"
	"math"
	"strings"
)

//SplitUser divides full user object to name & id
func SplitUser(user string) (string, string) {
	userFull := strings.Split(user, "|")
	userID := strings.Replace(userFull[0], "<@", "", -1)
	userName := strings.Replace(userFull[1], ">", "", -1)
	return userID, userName
}

//SplitChannel divides full channel object to name & id
func SplitChannel(channel string) (string, string) {
	channelFull := strings.Split(channel, "|")
	channelID := strings.Replace(channelFull[0], "<#", "", -1)
	channelName := strings.Replace(channelFull[1], ">", "", -1)
	return channelID, channelName
}

//SecondsToHuman converts seconds (int) to HH:MM format
func SecondsToHuman(input int) string {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	return fmt.Sprintf("%v:%02d", int(hours), int(minutes))
}

//CommandParsing parses string into command Title and Command Body
func CommandParsing(text string) (commandTitle, commandBody string) {
	text = strings.TrimSpace(text)
	splitText := strings.Split(text, " ")
	commandTitle = splitText[0]
	commandBody = strings.Join(splitText[1:], " ")
	return commandTitle, commandBody
}
