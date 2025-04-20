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
