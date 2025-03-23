package config

import (
	"crypto/tls"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"net/http"
	"time"
)

type Client struct {
	SkipSSLVerification bool  `default:"false" envconfig:"CLIENT_SKIP_TLS_VERIFICATION"`
	RetryDelay          int64 `default:"1" envconfig:"CLIENT_RETRY_DELAY_SECONDS"`
	Timeout             int64 `default:"20" envconfig:"CLIENT_TIMEOUT_SECONDS"`
}

func (c *Config) loadClient() error {
	client := Client{}

	if err := envconfig.Process("", &client); err != nil {
		return fmt.Errorf("client env vars: %w", err)
	}

	c.Client = &client

	return nil
}

func (settings *Client) NewHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Duration(settings.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: settings.SkipSSLVerification},
		},
	}
}
