package filter

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_ByType_Include(t *testing.T) {
	filterKeys := []string{"cache", "upstreams", "interface"}
	data := loadDNSData()
	result, err := ByType(Include, filterKeys, data)
	require.NoError(t, err)
	assert.Len(t, filterKeys, 3)
	assert.Len(t, result, 3)

	for key := range maps.Keys(data) {
		if slices.Contains(filterKeys, key) {
			assert.Contains(t, result, key)
			assert.Equal(t, data[key], result[key])
		} else {
			assert.NotContains(t, result, key)
		}
	}
}

func TestFilter_ByType_Exclude(t *testing.T) {
	filterKeys := []string{"cache", "upstreams", "interface"}
	data := loadDNSData()
	result, err := ByType(Exclude, filterKeys, data)
	require.NoError(t, err)
	assert.Equal(t, len(result), len(data)-len(filterKeys))

	for key := range maps.Keys(data) {
		if slices.Contains(filterKeys, key) {
			assert.NotContains(t, result, key)
		} else {
			assert.Contains(t, result, key)
			assert.Equal(t, data[key], result[key])
		}
	}
}

func TestFilter_ByType_MultipleNested(t *testing.T) {
	filterKeys := []string{"reply.host.force4", "reply.host.IPv4", "reply.blocking.force4"}
	data := loadDNSData()
	result, err := ByType(Include, filterKeys, data)
	require.NoError(t, err)
	assert.Len(t, result, 1)

	reply, ok := result["reply"].(map[string]any)
	assert.True(t, ok)
	host, ok := reply["host"].(map[string]any)
	assert.True(t, ok)
	blocking, ok := reply["blocking"].(map[string]any)
	assert.True(t, ok)

	assert.Len(t, reply, 2)
	assert.Len(t, host, 2)
	assert.Len(t, blocking, 1)
	assert.NotEqual(t, data["reply"].(map[string]any), reply)
}

func loadDNSData() map[string]any {
	file, err := os.ReadFile("../../../testdata/dns.json")
	if err != nil {
		panic("failed to read testdata")
	}

	var data map[string]any
	if err := json.Unmarshal(file, &data); err != nil {
		panic("failed to unmarshal testdata")
	}

	return data
}

func TestFilter_IncludeKeys(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2, "d": 3},
		"e": 4,
	}

	keys := []string{"a", "b.c", "e"}
	result := includeKeys(data, keys)

	assert.Equal(t, 1, result["a"])
	assert.Equal(t, 2, result["b"].(map[string]any)["c"])
	assert.Nil(t, result["b"].(map[string]any)["d"])
	assert.Equal(t, 4, result["e"])
	assert.Len(t, result, 3)
}

func TestFilter_IncludeKeys_MissingKey(t *testing.T) {
	data := map[string]any{"a": 1}
	keys := []string{"b"}
	result := includeKeys(data, keys)

	assert.Empty(t, result)
}

func TestFilter_ExcludeKeys(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": map[string]any{"c": 2, "d": 3},
		"e": 4,
	}

	keys := []string{"a", "b.c"}
	result := excludeKeys(data, keys)

	assert.NotContains(t, result, "a")
	assert.NotContains(t, result["b"].(map[string]any), "c")
	assert.Contains(t, result["b"].(map[string]any), "d")
	assert.Contains(t, result, "e")
}

func TestFilter_ExcludeKeys_NonExistentKey(t *testing.T) {
	data := map[string]any{"a": 1}
	keys := []string{"b"}
	result := excludeKeys(data, keys)

	assert.Equal(t, data, result)
}
