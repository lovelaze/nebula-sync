package filter

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"maps"
	"os"
	"slices"
	"testing"
)

func TestFilter_ByType_Include(t *testing.T) {
	filterKeys := []string{"cache", "upstreams", "interface"}
	data := loadDnsData()
	result, err := ByType(Include, filterKeys, data)
	assert.NoError(t, err)
	assert.Equal(t, len(result), len(filterKeys))

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
	data := loadDnsData()
	result, err := ByType(Exclude, filterKeys, data)
	assert.NoError(t, err)
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
	data := loadDnsData()
	result, err := ByType(Include, filterKeys, data)
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1)

	reply := result["reply"].(map[string]interface{})
	host := reply["host"].(map[string]interface{})
	blocking := reply["blocking"].(map[string]interface{})

	assert.Equal(t, len(reply), 2)
	assert.Equal(t, len(host), 2)
	assert.Equal(t, len(blocking), 1)
	assert.NotEqual(t, data["reply"].(map[string]interface{}), reply)
}

func loadDnsData() map[string]interface{} {
	file, err := os.ReadFile("../../../testdata/dns.json")
	if err != nil {
		panic("failed to read testdata")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		panic("failed to unmarshal testdata")
	}

	return data
}

func TestFilter_IncludeKeys(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{"c": 2, "d": 3},
		"e": 4,
	}

	keys := []string{"a", "b.c", "e"}
	result := includeKeys(data, keys)

	assert.Equal(t, 1, result["a"])
	assert.Equal(t, 2, result["b"].(map[string]interface{})["c"])
	assert.Equal(t, nil, result["b"].(map[string]interface{})["d"])
	assert.Equal(t, 4, result["e"])
	assert.Len(t, result, 3)
}

func TestFilter_IncludeKeys_MissingKey(t *testing.T) {
	data := map[string]interface{}{"a": 1}
	keys := []string{"b"}
	result := includeKeys(data, keys)

	assert.Empty(t, result)
}

func TestFilter_ExcludeKeys(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{"c": 2, "d": 3},
		"e": 4,
	}

	keys := []string{"a", "b.c"}
	result := excludeKeys(data, keys)

	assert.NotContains(t, result, "a")
	assert.NotContains(t, result["b"].(map[string]interface{}), "c")
	assert.Contains(t, result["b"].(map[string]interface{}), "d")
	assert.Contains(t, result, "e")
}

func TestFilter_ExcludeKeys_NonExistentKey(t *testing.T) {
	data := map[string]interface{}{"a": 1}
	keys := []string{"b"}
	result := excludeKeys(data, keys)

	assert.Equal(t, data, result)
}
