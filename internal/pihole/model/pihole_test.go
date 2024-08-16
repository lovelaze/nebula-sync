package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestPiHole_Decode(t *testing.T) {
	ph := PiHole{}
	const uri = "http://localhost:1337"
	const pw = "asdfa|sdf"

	err := ph.Decode(fmt.Sprintf("%s|%s", uri, pw))
	assert.NoError(t, err)

	expectedUrl, err := url.Parse(uri)
	assert.NoError(t, err)

	assert.Equal(t, expectedUrl, ph.Url)
	assert.Equal(t, pw, ph.Password)
}
