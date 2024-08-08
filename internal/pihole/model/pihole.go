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
	split := strings.Split(value, "|")
	if len(split) != 2 {
		return fmt.Errorf("invalid pihole format")
	}

	res, err := url.Parse(split[0])

	if err != nil {
		return fmt.Errorf("failed to parse url: %s", err)
	}

	*piHole = PiHole{
		Url:      res,
		Password: split[1],
	}
	return nil
}
