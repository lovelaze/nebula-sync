package config

import (
	"fmt"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"os"
	"strings"
)

func (c *Config) loadTargets() error {
	primary, err := loadPrimary()
	if err != nil {
		return err
	}

	replicas, err := loadReplicas()
	if err != nil {
		return err
	}

	c.Primary = *primary
	c.Replicas = replicas
	return nil
}

func loadPrimary() (*model.PiHole, error) {
	env := "PRIMARY"
	if value := os.Getenv(fmt.Sprintf("%s_FILE", env)); len(value) > 0 {
		if bytes, err := os.ReadFile(value); err != nil {
			return nil, err
		} else {
			return parse(strings.TrimSpace(string(bytes)))
		}
	} else if value := os.Getenv(env); len(value) > 0 {
		return parse(value)
	} else {
		return nil, fmt.Errorf("missing required env: %s/%s_FILE", env, env)
	}
}

func loadReplicas() ([]model.PiHole, error) {
	env := "REPLICAS"
	if value := os.Getenv(fmt.Sprintf("%s_FILE", env)); len(value) > 0 {
		if bytes, err := os.ReadFile(value); err != nil {
			return nil, err
		} else {
			return parseMultiple(strings.Split(strings.TrimSpace(string(bytes)), ","))
		}
	} else if value := os.Getenv(env); len(value) > 0 {
		return parseMultiple(strings.Split(value, ","))
	} else {
		return nil, fmt.Errorf("missing required env: %s/%s_FILE", env, env)
	}
}

func parse(value string) (*model.PiHole, error) {
	ph := model.PiHole{}
	if err := ph.Decode(value); err != nil {
		return nil, err
	}
	return &ph, nil
}

func parseMultiple(values []string) ([]model.PiHole, error) {
	replicas := []model.PiHole{}
	for _, value := range values {
		if ph, err := parse(value); err != nil {
			return nil, err
		} else {
			replicas = append(replicas, *ph)
		}

	}
	return replicas, nil
}
