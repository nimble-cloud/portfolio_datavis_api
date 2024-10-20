package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFakeit(t *testing.T) {
	InitEnv()
	InitDB()

	assert.NoError(t, FakeNames())
}
