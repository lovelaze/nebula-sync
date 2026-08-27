package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lovelaze/nebula-sync/internal/pihole/model"
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
		bytes, err := os.ReadFile(filepath.Clean(fileValue))
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
		bytes, err := os.ReadFile(filepath.Clean(fileValue))
		if err != nil {
			return nil, err
		}

		return parseMultiple(splitReplicaList(strings.TrimSpace(string(bytes))))
	} else if envValue := os.Getenv(env); len(envValue) > 0 {
		return parseMultiple(splitReplicaList(envValue))
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
		ph, err := parse(strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}

		replicas = append(replicas, *ph)
	}
	return replicas, nil
}

// splitReplicaList splits REPLICAS on commas that start a new URL, so passwords
// may contain commas without breaking the list (http://host|a,b,http://host2|c).
func splitReplicaList(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	replicas := make([]string, 0, len(parts))
	var current strings.Builder

	for _, part := range parts {
		if current.Len() == 0 {
			current.WriteString(part)
			continue
		}

		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			replicas = append(replicas, current.String())
			current.Reset()
			current.WriteString(part)
			continue
		}

		current.WriteByte(',')
		current.WriteString(part)
	}

	if current.Len() > 0 {
		replicas = append(replicas, current.String())
	}

	return replicas
}
