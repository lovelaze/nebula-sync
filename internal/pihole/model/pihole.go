package model

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
)

type PiHole struct {
	Url      *url.URL
	Password string
}

func (ph PiHole) String() string {
	return fmt.Sprintf("{Url:%s}", ph.Url)
}

func NewPiHole(host, password string) PiHole {
	u, err := url.Parse(host)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing host %s", host)
	}

	return PiHole{
		Url:      u,
		Password: password,
	}
}

func (piHole *PiHole) Decode(value string) error {
	uri, password, found := strings.Cut(value, "|")

	if !found {
		return fmt.Errorf("invalid pihole format")
	}

	parsedUrl, err := url.Parse(uri)

	if err != nil {
		return fmt.Errorf("parse url: %s", err)
	}

	*piHole = PiHole{
		Url:      parsedUrl,
		Password: password,
	}
	return nil
}
