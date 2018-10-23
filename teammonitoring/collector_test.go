package teammonitoring

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetCollectorData(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)

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
		url := fmt.Sprintf("%s/rest/api/v1/logger/%s/%s/%s/%s/%s/", c.CollectorURL, c.TeamDomain, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, ""))
		result, err := GetCollectorData(c, tt.getDataOn, tt.data, tt.dateFrom, tt.dateTo)
		assert.Equal(t, tt.err, err)
		fmt.Printf("Report on user: Total Commits: %v, Total Worklogs: %v\n\n", result.TotalCommits, utils.SecondsToHuman(result.Worklogs))
	}
}
