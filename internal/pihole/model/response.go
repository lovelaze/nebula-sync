package model

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type AuthResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		Totp     bool   `json:"totp"`
		Sid      string `json:"sid"`
		Csrf     string `json:"csrf"`
		Validity int    `json:"validity"`
		Message  string `json:"message"`
	} `json:"session"`
}

type DHCPLease struct {
	Expires  int64  `json:"expires"`
	Name     string `json:"name"`
	HWAddr   string `json:"hwaddr"`
	IP       string `json:"ip"`
	ClientID string `json:"clientid"`
}

type DHCPLeasesResponse struct {
	Leases []DHCPLease `json:"leases"`
}

type ConfigResponse struct {
	Config map[string]any `json:"config"`
}

func (c *ConfigResponse) Get(key string) map[string]any {
	value, exists := c.Config[key]
	if !exists {
		log.Warn().Msg(fmt.Sprintf("Missing key (%s) in config response", key))
		return nil
	}

	extracted, ok := value.(map[string]any)
	if !ok {
		log.Warn().Msg(fmt.Sprintf("Received unexpected type for key (%s) in config response", key))
		return nil
	}
	return extracted
}
