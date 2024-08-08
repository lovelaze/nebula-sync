package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"log"
)

type Config struct {
	Primary      model.PiHole   `required:"true" envconfig:"PRIMARY"`
	Replicas     []model.PiHole `required:"true" envconfig:"REPLICAS"`
	FullSync     bool           `required:"true" envconfig:"FULL_SYNC"`
	Cron         *string        `envconfig:"CRON"`
	SyncSettings *SyncSettings  `ignored:"true"`
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

func (c *Config) Load() {
	if err := envconfig.Process("", c); err != nil {
		log.Fatal(err)
	}

	if !c.FullSync {
		c.loadSyncSettings()
	}
}

func (c *Config) loadSyncSettings() {
	manualGravity := ManualGravity{}
	if err := envconfig.Process("", &manualGravity); err != nil {
		log.Fatal(err)
	}

	manualConfig := ManualConfig{}
	if err := envconfig.Process("", &manualConfig); err != nil {
		log.Fatal(err)
	}

	c.SyncSettings = &SyncSettings{
		Gravity: &manualGravity,
		Config:  &manualConfig,
	}
}
