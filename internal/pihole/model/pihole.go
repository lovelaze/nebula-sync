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

	*ph = PiHole{
		URL:      parsedURL,
		Password: password,
	}
	return nil
}
