package storage

import (
	"testing"

	// This line is must for working MySQL database
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gitlab.com/team-monitoring/comedian/config"
)

func TestNew(t *testing.T) {
	c := &config.Config{}
	_, err := New(c)
	assert.NoError(t, err)
}
