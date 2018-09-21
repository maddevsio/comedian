package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitUser(t *testing.T) {
	user := "<@USERID|userName"
	id, name := SplitUser(user)
	assert.Equal(t, "USERID", id)
	assert.Equal(t, "userName", name)
}

func TestSecondsToHuman(t *testing.T) {
	testCases := []struct {
		output  string
		seconds int
	}{
		{"00:03", 180},
		{"00:04", 240},
		{"01:00", 3600},
		{"01:03", 3780},
		{"01:03", 3782},
	}
	for _, tt := range testCases {
		text := SecondsToHuman(tt.seconds)
		assert.Equal(t, tt.output, text)
	}

}
