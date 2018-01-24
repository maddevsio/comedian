package storage

import (
	"database/sql"
	"testing"

	"github.com/maddevsio/comedian/config"
	"github.com/maddevsio/comedian/model"
	"github.com/stretchr/testify/assert"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
	"log"
	"os"
)

var (
	cleaner = dbcleaner.New()
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestCRUDLStandup(t *testing.T) {
	c, err := config.Get()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	s, err := m.CreateStandup(model.Standup{
		Comment:  "work hard",
		Username: "user",
	})
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "work hard")
	s.Comment = "Rest"
	s, err = m.UpdateStandup(s)
	assert.NoError(t, err)
	assert.Equal(t, s.Comment, "Rest")
	items, err := m.ListStandups()
	assert.NoError(t, err)
	assert.Equal(t, items[0], s)
	selected, err := m.SelectStandup(s.ID)
	assert.NoError(t, err)
	assert.Equal(t, s, selected)
	assert.NoError(t, m.DeleteStandup(s.ID))
	s, err = m.SelectStandup(s.ID)
	assert.Equal(t, err, sql.ErrNoRows)
	assert.Equal(t, s.ID, int64(0))

}

func testAddStandup(t *testing.T) {
	cleanDb()
	c, err := config.Get()
	assert.NoError(t, err)
	m, err := NewMySQL(c)
	assert.NoError(t, err)
	comedian, err := m.AddComedian(model.Comedian{
		SlackName: "@test",
		FullName:  "Test Testtt",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test", comedian.SlackName)
	assert.Equal(t, "Test Testtt", comedian.FullName)
}

func setup() {
	config, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	db := engine.NewMySQLEngine(config.DatabaseURL)
	cleaner.SetEngine(db)
}

func cleanDb() {
	cleaner.Clean("standup", "comedians")
}
