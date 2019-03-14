package storage

import (
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
	"gitlab.com/team-monitoring/comedian/model"
)

func TestUserSubmittedStandupToday(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	mysql, err := New(c)
	assert.NoError(t, err)

	st, err := mysql.CreateStandup(model.Standup{
		TeamID:    "foo",
		UserID:    "bar",
		ChannelID: "bar12",
		MessageTS: strconv.FormatInt((time.Now().Unix() - int64(1)), 10),
	})
	assert.NoError(t, err)
	assert.Equal(t, "foo", st.TeamID)

	ok, err := mysql.UserSubmittedStandupToday("bar12", "bar")
	if err != nil {
		log.Error(err)
	}
	assert.Equal(t, true, ok)

	assert.NoError(t, mysql.DeleteStandup(st.ID))

}
