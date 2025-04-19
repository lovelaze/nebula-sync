package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfig_Load_Target(t *testing.T) {
	conf := Config{}

	t.Setenv("PRIMARY", "http://localhost:1337|asdf")
	t.Setenv("REPLICAS", "http://localhost:1338|qwerty,http://localhost:1339|foobar")
	require.Empty(t, os.Getenv("PRIMARY_FILE"))
	require.Empty(t, os.Getenv("REPLICAS_FILE"))

	err := conf.loadTargets()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:1337", conf.Primary.URL.String())
	assert.Equal(t, "asdf", conf.Primary.Password)
	assert.Len(t, conf.Replicas, 2)
	assert.Equal(t, "http://localhost:1338", conf.Replicas[0].URL.String())
	assert.Equal(t, "qwerty", conf.Replicas[0].Password)
	assert.Equal(t, "http://localhost:1339", conf.Replicas[1].URL.String())
	assert.Equal(t, "foobar", conf.Replicas[1].Password)
}

func TestConfig_Load_TargetFiles(t *testing.T) {
	conf := Config{}

	t.Setenv("PRIMARY_FILE", "../../testdata/primary")
	t.Setenv("REPLICAS_FILE", "../../testdata/replicas")
	require.Empty(t, os.Getenv("PRIMARY"))
	require.Empty(t, os.Getenv("REPLICAS"))

	err := conf.loadTargets()
	require.NoError(t, err)

	assert.Equal(t, "https://ph1.example.com", conf.Primary.URL.String())
	assert.Equal(t, "password1", conf.Primary.Password)
	assert.Len(t, conf.Replicas, 2)
	assert.Equal(t, "https://ph2.example.com", conf.Replicas[0].URL.String())
	assert.Equal(t, "password2", conf.Replicas[0].Password)
	assert.Equal(t, "https://ph3.example.com", conf.Replicas[1].URL.String())
	assert.Equal(t, "password3", conf.Replicas[1].Password)
}

func TestConfig_Load_NoPrimary(t *testing.T) {
	conf := Config{}

	t.Setenv("REPLICAS_FILE", "../../testdata/replicas")
	require.Empty(t, os.Getenv("PRIMARY"))
	require.Empty(t, os.Getenv("PRIMARY_FILE"))
	require.Empty(t, os.Getenv("REPLICAS"))

	err := conf.loadTargets()
	assert.Error(t, err)
}

func TestConfig_Load_NoReplicas(t *testing.T) {
	conf := Config{}

	t.Setenv("PRIMARY_FILE", "../../testdata/primary")
	require.Empty(t, os.Getenv("PRIMARY"))
	require.Empty(t, os.Getenv("REPLICAS"))
	require.Empty(t, os.Getenv("REPLICAS_FILE"))

	err := conf.loadTargets()
	assert.Error(t, err)
}
