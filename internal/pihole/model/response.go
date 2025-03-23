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

type VersionResponse struct {
	Version struct {
		Core struct {
			Local struct {
				Version string `json:"version"`
				Branch  string `json:"branch"`
				Hash    string `json:"hash"`
			} `json:"local"`
			Remote struct {
				Version interface{} `json:"version"`
				Hash    string      `json:"hash"`
			} `json:"remote"`
		} `json:"core"`
		Web struct {
			Local struct {
				Version string `json:"version"`
				Branch  string `json:"branch"`
				Hash    string `json:"hash"`
			} `json:"local"`
			Remote struct {
				Version interface{} `json:"version"`
				Hash    string      `json:"hash"`
			} `json:"remote"`
		} `json:"web"`
		Ftl struct {
			Local struct {
				Hash    string `json:"hash"`
				Branch  string `json:"branch"`
				Version string `json:"version"`
				Date    string `json:"date"`
			} `json:"local"`
			Remote struct {
				Version interface{} `json:"version"`
				Hash    string      `json:"hash"`
			} `json:"remote"`
		} `json:"ftl"`
		Docker struct {
			Local  string `json:"local"`
			Remote string `json:"remote"`
		} `json:"docker"`
	} `json:"version"`
	Took float64 `json:"took"`
}

type ConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

func (c *ConfigResponse) Get(key string) map[string]interface{} {
	value, exists := c.Config[key]
	if !exists {
		log.Warn().Msg(fmt.Sprintf("Missing key (%s) in config response", key))
		return nil
	}
	return value.(map[string]interface{})
}
