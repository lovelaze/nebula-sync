package model

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestPiHole_Decode(t *testing.T) {
	ph := PiHole{}

	err := ph.Decode("http://localhost:1337|asdfasdf")
	assert.NoError(t, err)

	expectedUrl, err := url.Parse("http://localhost:1337")
	assert.NoError(t, err)

	assert.Equal(t, expectedUrl, ph.Url)
	assert.Equal(t, "asdfasdf", ph.Password)

}
