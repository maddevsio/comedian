package collector

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/bot"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/utils"
)

func TestGetCollectorData(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	bot, err := bot.NewBot(c)
	assert.NoError(t, err)
	//CollectorEnabled=false
	_, err = GetCollectorData(bot, "users", "U851AU1U0", "2018-10-12", "2018-10-14")
	assert.NoError(t, err)

	bot.CP.CollectorEnabled = true

	testCases := []struct {
		getDataOn string
		data      string
		dateFrom  string
		dateTo    string
		err       error
	}{
		{"users", "U851AU1U0", "2018-10-12", "2018-10-14", nil},
		{"projects", "comedian-testing", "2018-10-11", "2018-10-11", nil},
		{"user-in-project", "UC1JNECA3/comedian-testing", "2018-10-11", "2018-10-11", nil},
		{"user-in-project", "UD6143K51/standups", "2018-10-12", "2018-10-14", nil},
		{"user-in-project", "UD6147Z4K/standups", "2018-10-12", "2018-10-14", nil},
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range testCases {
		url := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, bot.TeamDomain, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, ""))
		result, err := GetCollectorData(bot, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		assert.Equal(t, tt.err, err)
		fmt.Printf("Report on user: Total Commits: %v, Total Worklogs: %v\n\n", result.Commits, utils.SecondsToHuman(result.Worklogs))
	}
}
