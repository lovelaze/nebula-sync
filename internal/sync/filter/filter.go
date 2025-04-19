package filter

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
)

type Type int

const (
	Include Type = iota
	Exclude
)

func (ft Type) String() string {
	var s string
	switch ft {
	case Include:
		s = "Include"
	case Exclude:
		s = "Exclude"
	}
	return s
}

func ByType(filter Type, keys []string, json map[string]any) (map[string]any, error) {
	switch filter {
	case Include:
		return includeKeys(json, keys), nil
	case Exclude:
		return excludeKeys(json, keys), nil
	default:
		return nil, fmt.Errorf("unknown filter type: %v", filter)
	}
}

func includeKeys(jsonData map[string]any, keys []string) map[string]any {
	result := make(map[string]any)

	for _, key := range keys {
		value := getNestedValue(jsonData, key)
		if value != nil {
			setNestedValue(result, key, value)
		} else {
			log.Warn().Str("key", key).Msg("Attempted to include missing config")
		}
	}

	return result
}

func excludeKeys(jsonData map[string]any, keys []string) map[string]any {
	result := deepCopy(jsonData)

	for _, key := range keys {
		removeNestedKey(result, strings.Split(key, "."))
	}

	return result
}

func getNestedValue(data map[string]any, key string) any {
	keys := strings.Split(key, ".")
	current := data
	for i, k := range keys {
		if next, ok := current[k].(map[string]any); ok {
			current = next
			if i == len(keys)-1 {
				return next
			}
		} else if value, ok := current[k]; ok {
			return value
		} else {
			return nil
		}
	}
	return current
}

func setNestedValue(target map[string]any, key string, value any) {
	keys := strings.Split(key, ".")
	current := target

	for _, k := range keys[:len(keys)-1] {
		if _, exists := current[k]; !exists {
			current[k] = make(map[string]any)
		}
		if next, ok := current[k].(map[string]any); ok {
			current = next
		}
	}

	lastKey := keys[len(keys)-1]
	current[lastKey] = value
}

func removeNestedKey(target map[string]any, keys []string) {
	if len(keys) == 0 {
		return
	}

	currentKey := keys[0]
	remainingKeys := keys[1:]

	_, exists := target[currentKey]
	if !exists {
		log.Warn().Str("key", strings.Join(keys, ".")).Msg("Attempted to exclude missing config")
		return
	}

	if len(remainingKeys) == 0 {
		delete(target, currentKey)
		return
	}

	if nested, exists := target[currentKey].(map[string]any); exists {
		removeNestedKey(nested, remainingKeys)
		if len(nested) == 0 {
			delete(target, currentKey)
		}
	}
}

func deepCopy(original map[string]any) map[string]any {
	copied := make(map[string]any)
	for key, value := range original {
		switch v := value.(type) {
		case map[string]any:
			copied[key] = deepCopy(v)
		default:
			copied[key] = v
		}
	}
	return copied
}
