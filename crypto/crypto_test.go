package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	hash, err := Generate("pass")
	assert.NoError(t, err)
	fmt.Println(hash)
}

func TestCompare(t *testing.T) {
	hash, err := Generate("pass")
	assert.NoError(t, err)
	err = Compare(hash, "password")
	assert.Error(t, err)
	err = Compare(hash, "pass")
	assert.NoError(t, err)
}
