package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Client struct {
	SkipTLSVerification bool  `default:"false" envconfig:"CLIENT_SKIP_TLS_VERIFICATION"`
	RetryDelay          int64 `default:"1"     envconfig:"CLIENT_RETRY_DELAY_SECONDS"`
	Timeout             int64 `default:"20"    envconfig:"CLIENT_TIMEOUT_SECONDS"`
}

func (c *Config) loadClient() error {
	client := Client{}

	if err := envconfig.Process("", &client); err != nil {
		return fmt.Errorf("client env vars: %w", err)
	}

	c.Client = &client

	return nil
}

func (c *Client) NewHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.SkipTLSVerification, //nolint:gosec // G402: opt-in for self-signed Pi-hole certs
	}

	return &http.Client{
		Timeout:   time.Duration(c.Timeout) * time.Second,
		Transport: transport,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("%+v", *c)
}
