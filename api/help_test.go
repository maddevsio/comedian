package api

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/chat"
	"gitlab.com/team-monitoring/comedian/config"
)

func SetUp() *REST {
	c, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	c.ManagerSlackUserID = "SuperAdminID"
	slack, err := chat.NewSlack(c)
	if err != nil {
		log.Fatal(err)
	}
	r, err := NewRESTAPI(slack)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func TestHelpText(t *testing.T) {
	r := SetUp()
	text := r.displayHelpText()
	assert.Equal(t, "Unite Slash Command is working!!!", text)
}
