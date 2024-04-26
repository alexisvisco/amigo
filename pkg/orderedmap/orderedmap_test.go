package orderedmap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderedMap_SetGet(t *testing.T) {
	om := New[int]()

	// Set some key-value pairs
	om.Set("key1", 10)
	om.Set("key2", 20)
	om.Set("key3", 30)

	// Get values and check if they match the expected ones
	value, ok := om.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 10, value)

	value, ok = om.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 20, value)

	value, ok = om.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, 30, value)

	// Try to get a non-existent key
	_, ok = om.Get("nonexistent")
	assert.False(t, ok)
}

func TestOrderedMap_Iterate(t *testing.T) {
	om := New[string]()

	// Set some key-value pairs
	om.Set("a", "alpha")
	om.Set("b", "beta")
	om.Set("c", "gamma")

	// Iterate through the map and check the order and values
	expectedOrder := []struct {
		Key   string
		Value string
	}{
		{"a", "alpha"},
		{"b", "beta"},
		{"c", "gamma"},
	}

	index := 0
	for entry := range om.Iterate() {
		assert.Equal(t, expectedOrder[index].Key, entry.Key)
		assert.Equal(t, expectedOrder[index].Value, entry.Value)
		index++
	}
}

func TestOrderedMap_DuplicateKeys(t *testing.T) {
	om := New[string]()

	// Set a key-value pair
	om.Set("key1", "value1")

	// Set the same key with a different value
	om.Set("key1", "value2")

	// The value for the key should be updated, but the order should remain the same
	value, ok := om.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)

	// Iterate through the map and check that there is still only one key
	count := 0
	for range om.Iterate() {
		count++
	}
	assert.Equal(t, 1, count)
}

func TestOrderedMap_EmptyMap(t *testing.T) {
	om := New[string]()

	// Test empty map for Get
	_, ok := om.Get("nonexistent")
	assert.False(t, ok)

	// Test empty map for Iterate
	count := 0
	for range om.Iterate() {
		count++
	}
	assert.Equal(t, 0, count)
}
