package teammonitoring

import (
	"fmt"
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/stretchr/testify/assert"
)

func TestGetCollectorData(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	if !c.TeamMonitoringEnabled {
		fmt.Println("Warning: Team Monitoring servise is disabled")
		return
	}

	dataOnUser, err := GetCollectorData(c, "users", "UBZ6Y0P5K", "2018-09-18", "2018-09-18")
	assert.NoError(t, err)
	fmt.Printf("Report on user: Total Commits: %v, Total Worklogs: %v\n\n", dataOnUser.TotalCommits, dataOnUser.Worklogs/3600)

	dataOnProject, err := GetCollectorData(c, "projects", "comedian-testing", "2018-09-18", "2018-09-18")
	assert.NoError(t, err)
	fmt.Printf("Report on project: Total Commits: %v, Total Worklogs: %v\n\n", dataOnProject.TotalCommits, dataOnProject.Worklogs/3600)

	dataOnUserByProject, err := GetCollectorData(c, "user-in-project", "UC1JNECA3/comedian-testing", "2018-09-18", "2018-09-18")
	assert.NoError(t, err)
	fmt.Printf("Report on user in project: Total Commits: %v, Total Worklogs: %v\n\n", dataOnUserByProject.TotalCommits, dataOnUserByProject.Worklogs/3600)
}
