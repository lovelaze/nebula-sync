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

func TestPiHole_Decode_StripsLegacyAdminPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "admin", in: "https://dns.example/admin|secret", want: "https://dns.example"},
		{name: "admin login", in: "https://dns.example/admin/login|secret", want: "https://dns.example"},
		{name: "admin trailing slash", in: "https://dns.example/admin/|secret", want: "https://dns.example"},
		{name: "custom path kept", in: "https://dns.example/pihole|secret", want: "https://dns.example/pihole"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ph := PiHole{}
			err := ph.Decode(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.want, ph.URL.String())
			assert.Equal(t, "secret", ph.Password)
		})
	}
}
