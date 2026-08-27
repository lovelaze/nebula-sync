package model

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

type PiHole struct {
	URL      *url.URL
	Password string
}

func NewPiHole(host, password string) PiHole {
	u, err := url.Parse(host)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing host %s", host)
	}

	return PiHole{
		URL:      u,
		Password: password,
	}
}

func (ph *PiHole) String() string {
	return fmt.Sprintf("{URL:%s}", ph.URL)
}

func (ph *PiHole) Decode(value string) error {
	uri, password, found := strings.Cut(value, "|")

	if !found {
		return errors.New("invalid pihole format")
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	stripLegacyAdminPath(parsedURL)

	*ph = PiHole{
		URL:      parsedURL,
		Password: password,
	}
	return nil
}

// stripLegacyAdminPath removes /admin and /admin/login suffixes that users often
// copy from the Pi-hole web UI. The v6 API lives at /api on the same origin, so
// joining those paths produces /admin/login/api/auth and auth fails.
func stripLegacyAdminPath(u *url.URL) {
	path := strings.TrimSuffix(u.Path, "/")
	if strings.EqualFold(path, "/admin") || strings.EqualFold(path, "/admin/login") {
		u.Path = ""
		u.RawPath = ""
	}
}
