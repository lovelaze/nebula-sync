package config

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type Config struct {
	Primary        model.PiHole    `required:"true" envconfig:"PRIMARY"`
	Replicas       []model.PiHole  `required:"true" envconfig:"REPLICAS"`
	FullSync       bool            `required:"true" envconfig:"FULL_SYNC"`
	Cron           *string         `envconfig:"CRON"`
	ClientSettings *ClientSettings `ignored:"true"`
	SyncSettings   *SyncSettings   `ignored:"true"`
}

type ClientSettings struct {
	SkipSSLVerification bool `default:"false" envconfig:"CLIENT_SKIP_TLS_VERIFICATION"`
}

func (cs *ClientSettings) NewHttpClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cs.SkipSSLVerification},
		},
	}
}

type ManualGravity struct {
	DHCPLeases        bool `default:"false" envconfig:"SYNC_GRAVITY_DHCP_LEASES"`
	Group             bool `default:"false" envconfig:"SYNC_GRAVITY_GROUP"`
	Adlist            bool `default:"false" envconfig:"SYNC_GRAVITY_AD_LIST"`
	AdlistByGroup     bool `default:"false" envconfig:"SYNC_GRAVITY_AD_LIST_BY_GROUP"`
	Domainlist        bool `default:"false" envconfig:"SYNC_GRAVITY_DOMAIN_LIST"`
	DomainlistByGroup bool `default:"false" envconfig:"SYNC_GRAVITY_DOMAIN_LIST_BY_GROUP"`
	Client            bool `default:"false" envconfig:"SYNC_GRAVITY_CLIENT"`
	ClientByGroup     bool `default:"false" envconfig:"SYNC_GRAVITY_CLIENT_BY_GROUP"`
}

type ManualConfig struct {
	DNS       bool `default:"false" envconfig:"SYNC_CONFIG_DNS"`
	DHCP      bool `default:"false" envconfig:"SYNC_CONFIG_DHCP"`
	NTP       bool `default:"false" envconfig:"SYNC_CONFIG_NTP"`
	Resolver  bool `default:"false" envconfig:"SYNC_CONFIG_RESOLVER"`
	Database  bool `default:"false" envconfig:"SYNC_CONFIG_DATABASE"`
	Webserver bool `default:"false" ignored:"true"` // ignore for now
	Files     bool `default:"false" ignored:"true"` // ignore for now
	Misc      bool `default:"false" envconfig:"SYNC_CONFIG_MISC"`
	Debug     bool `default:"false" envconfig:"SYNC_CONFIG_DEBUG"`
}

type SyncSettings struct {
	Gravity *ManualGravity `ignored:"true"`
	Config  *ManualConfig  `ignored:"true"`
}

func (c *Config) Load() error {
	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("env vars: %w", err)
	}

	if err := c.loadClientSettings(); err != nil {
		return err
	}

	if !c.FullSync {
		if err := c.loadSyncSettings(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) loadClientSettings() error {
	clientSettings := ClientSettings{}

	if err := envconfig.Process("", &clientSettings); err != nil {
		return fmt.Errorf("client env vars: %w", err)
	}

	c.ClientSettings = &clientSettings

	return nil
}

func (c *Config) loadSyncSettings() error {
	manualGravity := ManualGravity{}
	if err := envconfig.Process("", &manualGravity); err != nil {
		return fmt.Errorf("gravity env vars: %w", err)
	}

	manualConfig := ManualConfig{}
	if err := envconfig.Process("", &manualConfig); err != nil {
		return fmt.Errorf("config env vars: %w", err)
	}

	c.SyncSettings = &SyncSettings{
		Gravity: &manualGravity,
		Config:  &manualConfig,
	}

	return nil
}

func LoadEnvFile(filename string) error {
	log.Debug().Msgf("Loading env file: %s", filename)
	return godotenv.Load(filename)
}

func (c *Config) String() string {
	replicas := make([]string, len(c.Replicas))
	for _, replica := range c.Replicas {
		replicas = append(replicas, replica.Url.String())
	}

	cron := ""
	if c.Cron != nil {
		cron = *c.Cron
	}

	syncSettings := ""
	if c.SyncSettings != nil {
		if mc := c.SyncSettings.Config; mc != nil {
			syncSettings += fmt.Sprintf("config=%+v", *mc)
		}
		if gc := c.SyncSettings.Gravity; gc != nil {
			syncSettings += fmt.Sprintf(", gravity=%+v", *gc)
		}
	}

	return fmt.Sprintf("primary=%s, replicas=%s, fullSync=%t, cron=%s, syncSettings=%s", c.Primary.Url, replicas, c.FullSync, cron, syncSettings)
}
