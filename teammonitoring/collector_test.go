package teammonitoring

import (
	"fmt"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/comedian/chat"
	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/maddevsio/comedian/utils"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestTeamMonitoringIsOFF(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	c.TeamMonitoringEnabled = false
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	_, err = NewTeamMonitoring(c, slack)
	assert.Error(t, err)
	assert.Equal(t, "team monitoring is disabled", err.Error())
}

func TestTeamMonitoringOnWeekEnd(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	if !c.TeamMonitoringEnabled {
		fmt.Println("Warning: Team Monitoring servise is disabled")
		return
	}
	tm, err := NewTeamMonitoring(c, slack)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 16, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	_, err = tm.RevealRooks()
	assert.Error(t, err)
	assert.Equal(t, "Day off today! Next report on Monday!", err.Error())
}

func TestTeamMonitoringOnMonday(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	s, err := chat.NewSlack(c)
	assert.NoError(t, err)
	if !c.TeamMonitoringEnabled {
		fmt.Println("Warning: Team Monitoring servise is disabled")
		return
	}
	tm, err := NewTeamMonitoring(c, s)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	attachments, err := tm.RevealRooks()
	assert.NoError(t, err)
	assert.Equal(t, []slack.Attachment{}, attachments)
}

func TestTeamMonitoringOnWeekDay(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	slack, err := chat.NewSlack(c)
	assert.NoError(t, err)
	if !c.TeamMonitoringEnabled {
		fmt.Println("Warning: Team Monitoring servise is disabled")
		return
	}
	tm, err := NewTeamMonitoring(c, slack)
	assert.NoError(t, err)

	d := time.Date(2018, 9, 17, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	ch, err := tm.db.CreateChannel(model.Channel{
		ChannelID:   "QWERTY123",
		ChannelName: "chanName1",
		StandupTime: int64(0),
	})

	su, err := tm.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID1",
		ChannelID: ch.ChannelID,
	})
	assert.NoError(t, err)
	su2, err := tm.db.CreateChannelMember(model.ChannelMember{
		UserID:    "userID2",
		ChannelID: ch.ChannelID,
	})

	s, err := tm.db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		ChannelID: ch.ChannelID,
		Comment:   "work hard",
		UserID:    "userID1",
		MessageTS: "qweasdzxc",
	})
	assert.NoError(t, err)

	s1, err := tm.db.CreateStandup(model.Standup{
		Created:   time.Now(),
		Modified:  time.Now(),
		ChannelID: ch.ChannelID,
		Comment:   "",
		UserID:    "userID2",
		MessageTS: "djklsfklfjsdl",
	})
	assert.NoError(t, err)

	d = time.Date(2018, 9, 18, 10, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })

	text, err := tm.RevealRooks()
	assert.NoError(t, err)
	assert.Equal(t, "<@userID1> twiddled in #chanName1 yesterday! (Not enough worklogs: 00:00 hours,no commits,submitted a standup!)\n\n\n<@userID2> twiddled in #chanName1 yesterday! (Not enough worklogs: 00:00 hours,no commits,did not write standup!)\n\n\n", text)

	assert.NoError(t, tm.db.DeleteChannelMember(su.UserID, su.ChannelID))
	assert.NoError(t, tm.db.DeleteChannelMember(su2.UserID, su2.ChannelID))

	assert.NoError(t, tm.db.DeleteStandup(s.ID))
	assert.NoError(t, tm.db.DeleteStandup(s1.ID))
	assert.NoError(t, tm.db.DeleteChannel(ch.ID))
}

func TestGetCollectorData(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	if !c.TeamMonitoringEnabled {
		fmt.Println("Warning: Team Monitoring servise is disabled")
		return
	}

	dataOnUser, err := GetCollectorData(c, "users", "UC1JNECA3", "2018-10-11", "2018-10-11")
	assert.NoError(t, err)
	fmt.Printf("Report on user: Total Commits: %v, Total Worklogs: %v\n\n", dataOnUser.TotalCommits, utils.SecondsToHuman(dataOnUser.Worklogs))

	dataOnProject, err := GetCollectorData(c, "projects", "comedian-testing", "2018-10-11", "2018-10-11")
	assert.NoError(t, err)
	fmt.Printf("Report on project: Total Commits: %v, Total Worklogs: %v\n\n", dataOnProject.TotalCommits, utils.SecondsToHuman(dataOnProject.Worklogs))

	dataOnUserByProject, err := GetCollectorData(c, "user-in-project", "UC1JNECA3/comedian-testing", "2018-10-11", "2018-10-11")
	assert.NoError(t, err)
	fmt.Printf("Report on user in project: Total Commits: %v, Total Worklogs: %v\n\n", dataOnUserByProject.TotalCommits, utils.SecondsToHuman(dataOnUserByProject.Worklogs))

}
