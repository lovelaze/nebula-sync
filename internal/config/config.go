package config

import (
	"crypto/tls"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Primary  model.PiHole   `required:"true" envconfig:"PRIMARY"`
	Replicas []model.PiHole `required:"true" envconfig:"REPLICAS"`
	Client   *Client        `ignored:"true"`
	Sync     *Sync          `ignored:"true"`
}

type Client struct {
	SkipSSLVerification bool  `default:"false" envconfig:"CLIENT_SKIP_TLS_VERIFICATION"`
	RetryDelay          int64 `default:"1" envconfig:"CLIENT_RETRY_DELAY_SECONDS"`
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

type Sync struct {
	FullSync        bool    `required:"true" envconfig:"FULL_SYNC"`
	Cron            *string `envconfig:"CRON"`
	RunGravity      bool    `default:"false" envconfig:"RUN_GRAVITY"`
	GravitySettings *GravitySettings
	ConfigSettings  *ConfigSettings
}

func (c *Config) Load() error {
	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("env vars: %w", err)
	}

	if err := c.loadClient(); err != nil {
		return err
	}

	if err := c.loadSync(); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadClient() error {
	client := Client{}

	if err := envconfig.Process("", &client); err != nil {
		return fmt.Errorf("client env vars: %w", err)
	}

	c.Client = &client

	return nil
}

func (c *Config) loadSync() error {
	sync := Sync{}
	if err := envconfig.Process("", &sync); err != nil {
		return fmt.Errorf("sync env vars: %w", err)
	}

	c.Sync = &sync
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
	if c.Sync.Cron != nil {
		cron = *c.Sync.Cron
	}

	sync := ""
	if c.Sync != nil {
		if mc := c.Sync.ConfigSettings; mc != nil {
			sync += fmt.Sprintf("config=%+v", *mc)
		}
		if gc := c.Sync.GravitySettings; gc != nil {
			sync += fmt.Sprintf(", gravity=%+v", *gc)
		}
	}

	return fmt.Sprintf("primary=%s, replicas=%s, fullSync=%t, cron=%s, sync=%s", c.Primary.Url, replicas, c.Sync.FullSync, cron, sync)
}

func (cs *Client) NewHttpClient() *http.Client {
	defaultTimeout := 20 * time.Second

	timeoutEnv := os.Getenv("HTTP_CLIENT_TIMEOUT")
	if timeoutEnv != "" {
		if timeout, err := strconv.Atoi(timeoutEnv); err == nil {
			defaultTimeout = time.Duration(timeout) * time.Second
		}
	}

	return &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cs.SkipSSLVerification},
		},
	}
}
