package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/lovelaze/nebula-sync/internal/sync/filter"
)

type Config struct {
	Primary  model.PiHole   `ignored:"true" required:"true" envconfig:"PRIMARY"`
	Replicas []model.PiHole `ignored:"true" required:"true" envconfig:"REPLICAS"`
	Client   *Client        `ignored:"true"`
	Sync     *Sync          `ignored:"true"`
	API      *API           `                               envconfig:"API"`
}

type Sync struct {
	FullSync          bool     `required:"true" envconfig:"FULL_SYNC"`
	Cron              *string  `                envconfig:"CRON"`
	RunGravity        bool     `                envconfig:"RUN_GRAVITY"                     default:"false"`
	ExcludeDNSRecords []string `                envconfig:"SYNC_CONFIG_DNS_EXCLUDE_RECORDS"`
	GravitySettings   *GravitySettings
	ConfigSettings    *ConfigSettings  `                                                                            ignored:"true"`
	WebhookSettings   *WebhookSettings `                                                                            ignored:"true"`
}

type GravitySettings struct {
	DHCPLeases        bool `default:"false" envconfig:"SYNC_GRAVITY_DHCP_LEASES"`
	Group             bool `default:"false" envconfig:"SYNC_GRAVITY_GROUP"`
	Adlist            bool `default:"false" envconfig:"SYNC_GRAVITY_AD_LIST"`
	AdlistByGroup     bool `default:"false" envconfig:"SYNC_GRAVITY_AD_LIST_BY_GROUP"`
	Domainlist        bool `default:"false" envconfig:"SYNC_GRAVITY_DOMAIN_LIST"`
	DomainlistByGroup bool `default:"false" envconfig:"SYNC_GRAVITY_DOMAIN_LIST_BY_GROUP"`
	Client            bool `default:"false" envconfig:"SYNC_GRAVITY_CLIENT"`
	ClientByGroup     bool `default:"false" envconfig:"SYNC_GRAVITY_CLIENT_BY_GROUP"`
}

type ConfigSettings struct {
	DNS       *ConfigSetting
	DHCP      *ConfigSetting
	NTP       *ConfigSetting
	Resolver  *ConfigSetting
	Database  *ConfigSetting
	Webserver *ConfigSetting
	Files     *ConfigSetting
	Misc      *ConfigSetting
	Debug     *ConfigSetting
}

type RawConfigSettings struct {
	DNS             bool     `default:"false" envconfig:"SYNC_CONFIG_DNS"`
	DNSInclude      []string `                envconfig:"SYNC_CONFIG_DNS_INCLUDE"`
	DNSExclude      []string `                envconfig:"SYNC_CONFIG_DNS_EXCLUDE"`
	DHCP            bool     `default:"false" envconfig:"SYNC_CONFIG_DHCP"`
	DHCPInclude     []string `                envconfig:"SYNC_CONFIG_DHCP_INCLUDE"`
	DHCPExclude     []string `                envconfig:"SYNC_CONFIG_DHCP_EXCLUDE"`
	NTP             bool     `default:"false" envconfig:"SYNC_CONFIG_NTP"`
	NTPInclude      []string `                envconfig:"SYNC_CONFIG_NTP_INCLUDE"`
	NTPExclude      []string `                envconfig:"SYNC_CONFIG_NTP_EXCLUDE"`
	Resolver        bool     `default:"false" envconfig:"SYNC_CONFIG_RESOLVER"`
	ResolverInclude []string `                envconfig:"SYNC_CONFIG_RESOLVER_INCLUDE"`
	ResolverExclude []string `                envconfig:"SYNC_CONFIG_RESOLVER_EXCLUDE"`
	Database        bool     `default:"false" envconfig:"SYNC_CONFIG_DATABASE"`
	DatabaseInclude []string `                envconfig:"SYNC_CONFIG_DATABASE_INCLUDE"`
	DatabaseExclude []string `                envconfig:"SYNC_CONFIG_DATABASE_EXCLUDE"`
	Webserver       bool     `default:"false"                                          ignored:"true"` // ignore for now
	Files           bool     `default:"false"                                          ignored:"true"` // ignore for now
	Misc            bool     `default:"false" envconfig:"SYNC_CONFIG_MISC"`
	MiscInclude     []string `                envconfig:"SYNC_CONFIG_MISC_INCLUDE"`
	MiscExclude     []string `                envconfig:"SYNC_CONFIG_MISC_EXCLUDE"`
	Debug           bool     `default:"false" envconfig:"SYNC_CONFIG_DEBUG"`
	DebugInclude    []string `                envconfig:"SYNC_CONFIG_DEBUG_INCLUDE"`
	DebugExclude    []string `                envconfig:"SYNC_CONFIG_DEBUG_EXCLUDE"`
}

func (raw *RawConfigSettings) Validate() error {
	exclusive := func(name string, include, exclude []string) error {
		if include != nil && exclude != nil {
			return fmt.Errorf("%s: INCLUDE/EXCLUDE must be mutually exclusive", name)
		}
		return nil
	}

	if err := exclusive("dns", raw.DNSInclude, raw.DNSExclude); err != nil {
		return err
	}
	if err := exclusive("dhcp", raw.DHCPInclude, raw.DHCPExclude); err != nil {
		return err
	}
	if err := exclusive("ntp", raw.NTPInclude, raw.NTPExclude); err != nil {
		return err
	}
	if err := exclusive("resolver", raw.ResolverInclude, raw.ResolverExclude); err != nil {
		return err
	}
	if err := exclusive("database", raw.DatabaseInclude, raw.DatabaseExclude); err != nil {
		return err
	}
	if err := exclusive("misc", raw.MiscInclude, raw.MiscExclude); err != nil {
		return err
	}

	return exclusive("debug", raw.DebugInclude, raw.DebugExclude)
}

func (raw *RawConfigSettings) Parse() (*ConfigSettings, error) {
	if err := raw.Validate(); err != nil {
		return nil, err
	}

	return &ConfigSettings{
		DNS:       NewConfigSetting(raw.DNS, raw.DNSInclude, raw.DNSExclude),
		DHCP:      NewConfigSetting(raw.DHCP, raw.DHCPInclude, raw.DHCPExclude),
		NTP:       NewConfigSetting(raw.NTP, raw.NTPInclude, raw.NTPExclude),
		Resolver:  NewConfigSetting(raw.Resolver, raw.ResolverInclude, raw.ResolverExclude),
		Database:  NewConfigSetting(raw.Database, raw.DatabaseInclude, raw.DatabaseExclude),
		Webserver: NewConfigSetting(raw.Webserver, nil, nil),
		Files:     NewConfigSetting(raw.Files, nil, nil),
		Misc:      NewConfigSetting(raw.Misc, raw.MiscInclude, raw.MiscExclude),
		Debug:     NewConfigSetting(raw.Debug, raw.DebugInclude, raw.DebugExclude),
	}, nil
}

type ConfigSetting struct {
	Enabled bool
	Filter  *ConfigFilter
}

type ConfigFilter struct {
	Type filter.Type
	Keys []string
}

func newConfigFilter(filterType filter.Type, keys []string) *ConfigFilter {
	return &ConfigFilter{
		Type: filterType,
		Keys: keys,
	}
}

func NewConfigSetting(enabled bool, included, excluded []string) *ConfigSetting {
	var configFilter *ConfigFilter

	switch {
	case included != nil:
		configFilter = newConfigFilter(filter.Include, included)
	case excluded != nil:
		configFilter = newConfigFilter(filter.Exclude, excluded)
	default:
		configFilter = nil
	}

	return &ConfigSetting{
		Enabled: enabled,
		Filter:  configFilter,
	}
}

func (c *Config) Load() error {
	if err := envconfig.Process("", c); err != nil {
		return err
	}

	if err := c.loadTargets(); err != nil {
		return err
	}

	if err := c.loadClient(); err != nil {
		return err
	}

	if err := c.loadSync(); err != nil {
		return err
	}

	return c.loadWebhookSettings()
}

func (c *Config) loadSync() error {
	sync := Sync{}
	if err := envconfig.Process("", &sync); err != nil {
		return fmt.Errorf("sync env vars: %w", err)
	}

	if err := sync.loadConfigSettings(); err != nil {
		return fmt.Errorf("load config settings: %w", err)
	}

	c.Sync = &sync
	return nil
}

func (s *Sync) loadConfigSettings() error {
	raw := RawConfigSettings{}

	if err := envconfig.Process("", &raw); err != nil {
		return fmt.Errorf("config settings env vars: %w", err)
	}

	configSettings, err := raw.Parse()
	if err != nil {
		return err
	}

	s.ConfigSettings = configSettings
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%+v", *c)
}

func (s *Sync) String() string {
	return fmt.Sprintf("%+v", *s)
}

func (a *API) String() string {
	return fmt.Sprintf("%+v", *a)
}

func (gs *GravitySettings) String() string {
	return fmt.Sprintf("%+v", *gs)
}

func (cs *ConfigSettings) String() string {
	return fmt.Sprintf("%+v", *cs)
}

func (cs *ConfigSetting) String() string {
	return fmt.Sprintf("%+v", *cs)
}

func (cs *ConfigFilter) String() string {
	return fmt.Sprintf("%+v", *cs)
}

func (wes *WebhookRequest) String() string {
	return fmt.Sprintf("%+v", *wes)
}

func (wes *WebhookSettings) String() string {
	return fmt.Sprintf("%+v", *wes)
}
