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
	if fileValue := os.Getenv(fmt.Sprintf("%s_FILE", env)); len(fileValue) > 0 {
		bytes, err := os.ReadFile(fileValue)
		if err != nil {
			return nil, err
		}

		return parse(strings.TrimSpace(string(bytes)))
	} else if envValue := os.Getenv(env); len(envValue) > 0 {
		return parse(envValue)
	}

	return nil, fmt.Errorf("missing required env: %s/%s_FILE", env, env)
}

func loadReplicas() ([]model.PiHole, error) {
	env := "REPLICAS"
	if fileValue := os.Getenv(fmt.Sprintf("%s_FILE", env)); len(fileValue) > 0 {
		bytes, err := os.ReadFile(fileValue)
		if err != nil {
			return nil, err
		}

		return parseMultiple(strings.Split(strings.TrimSpace(string(bytes)), ","))
	} else if envValue := os.Getenv(env); len(envValue) > 0 {
		return parseMultiple(strings.Split(envValue, ","))
	}

	return nil, fmt.Errorf("missing required env: %s/%s_FILE", env, env)
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
		ph, err := parse(value)
		if err != nil {
			return nil, err
		}
		
		replicas = append(replicas, *ph)
	}
	return replicas, nil
}
