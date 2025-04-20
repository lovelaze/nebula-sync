package model

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPiHole_Decode(t *testing.T) {
	ph := PiHole{}
	const uri = "http://localhost:1337"
	const pw = "asdfa|sdf"

	err := ph.Decode(fmt.Sprintf("%s|%s", uri, pw))
	require.NoError(t, err)

	expectedURL, err := url.Parse(uri)
	require.NoError(t, err)

	assert.Equal(t, expectedURL, ph.URL)
	assert.Equal(t, pw, ph.Password)
}
