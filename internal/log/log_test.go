package log

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestInit_info(t *testing.T) {
	t.Setenv("NS_DEBUG", "false")
	Init()

	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
}

func TestInit_debug(t *testing.T) {
	t.Setenv("NS_DEBUG", "true")
	Init()

	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}
