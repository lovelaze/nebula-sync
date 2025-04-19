package config

type API struct {
	Enabled bool `default:"false" envconfig:"ENABLED"` // internal use only
}
