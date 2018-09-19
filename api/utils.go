package api

import (
	"strings"
)

//SplitUser divides full user object to name & id
func SplitUser(user string) (string, string) {
	userFull := strings.Split(user, "|")
	userID := strings.Replace(userFull[0], "<@", "", -1)
	userName := strings.Replace(userFull[1], ">", "", -1)
	return userID, userName
}
