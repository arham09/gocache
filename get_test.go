package gocache

import (
	"testing"
	"time"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
}

func TestCache_GetExpired(t *testing.T) {
	cache := NewCache()
	cache.SetWithTTL("key", "value", time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	_, ok := cache.Get("key")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestCache_GetEntryThatHasNotExpiredYet(t *testing.T) {
	cache := NewCache()
	cache.SetWithTTL("key", "value", time.Hour)
	_, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to not have expired")
	}
}

func TestCache_GetValue(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("key", "value")
	value := cache.GetValue("key")
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
}

func TestCache_GetByKeys(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	keyValues := cache.GetByKeys([]string{"key1", "key2", "key3"})
	if len(keyValues) != 3 {
		t.Error("expected length of map to be 3")
	}
	if keyValues["key1"] != "value1" {
		t.Errorf("expected: %s, but got: %s", "value1", keyValues["key1"])
	}
	if keyValues["key2"] != "value2" {
		t.Errorf("expected: %s, but got: %s", "value2", keyValues["key2"])
	}
	if value, ok := keyValues["key3"]; !ok || value != nil {
		t.Errorf("expected key3 to exist and be nil, but got: %s", value)
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	keyValues := cache.GetAll()
	if len(keyValues) != 2 {
		t.Error("expected length of map to be 2")
	}
	if keyValues["key1"] != "value1" {
		t.Errorf("expected: %s, but got: %s", "value1", keyValues["key1"])
	}
	if keyValues["key2"] != "value2" {
		t.Errorf("expected: %s, but got: %s", "value2", keyValues["key2"])
	}
}

func TestCache_GetAllWhenOneKeyIsExpired(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.SetWithTTL("key3", "value3", time.Nanosecond)
	time.Sleep(time.Millisecond)
	keyValues := cache.GetAll()
	if len(keyValues) != 2 {
		t.Error("expected length of map to be 2")
	}
	if keyValues["key1"] != "value1" {
		t.Errorf("expected: %s, but got: %s", "value1", keyValues["key1"])
	}
	if keyValues["key2"] != "value2" {
		t.Errorf("expected: %s, but got: %s", "value2", keyValues["key2"])
	}
}

func TestCache_GetKeysByPattern(t *testing.T) {
	// All keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "key*", 0, 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*y*", 0, 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*key*", 0, 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*", 0, 4)
	// All keys match but limit is reached
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*", 2, 2)
	// Some keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11"}, "key1*", 0, 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11"}, "*key1*", 0, 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "key1*", 0, 3)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "key11*", 0, 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "*11*", 0, 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "k*1*", 0, 3)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "*k*1", 0, 3)
	// No keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "image*", 0, 0)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "?", 0, 0)
}

func testGetKeysByPattern(t *testing.T, keys []string, pattern string, limit, expectedMatchingKeys int) {
	cache := NewCache(WithMaxSize(len(keys)))
	for _, key := range keys {
		cache.Set(key, key)
	}
	matchingKeys := cache.GetKeysByPattern(pattern, limit)
	if len(matchingKeys) != expectedMatchingKeys {
		t.Errorf("expected to have %d keys to match pattern '%s', got %d", expectedMatchingKeys, pattern, len(matchingKeys))
	}
}

func TestCache_GetKeysByPatternWithExpiredKey(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.SetWithTTL("key", "value", 10*time.Millisecond)
	// The cache entry shouldn't have expired yet, so GetKeysByPattern should return 1 key
	if matchingKeys := cache.GetKeysByPattern("*", 0); len(matchingKeys) != 1 {
		t.Errorf("expected to have %d keys to match pattern '%s', got %d", 1, "*", len(matchingKeys))
	}
	time.Sleep(30 * time.Millisecond)
	// Since the key expired, the same call should return 0 keys instead of 1
	if matchingKeys := cache.GetKeysByPattern("*", 0); len(matchingKeys) != 0 {
		t.Errorf("expected to have %d keys to match pattern '%s', got %d", 0, "*", len(matchingKeys))
	}
}
