package gocache

import (
	"bytes"
	"testing"
)

func TestCache_Set(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
	cache.Set("key", "newvalue")
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "newvalue" {
		t.Errorf("expected: %s, but got: %s", "newvalue", value)
	}
}

func TestCache_SetDifferentTypesOfData(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", 1)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 1 {
		t.Errorf("expected: %v, but got: %v", 1, value)
	}
	cache.Set("key", struct{ Test string }{Test: "test"})
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value.(struct{ Test string }) != struct{ Test string }{Test: "test"} {
		t.Errorf("expected: %s, but got: %s", "newvalue", value)
	}
}

func TestCache_SetGetInt(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", 1)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 1 {
		t.Errorf("expected: %v, but got: %v", 1, value)
	}
	cache.Set("key", 2.1)
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 2.1 {
		t.Errorf("expected: %v, but got: %v", 2.1, value)
	}
}

func TestCache_SetGetBool(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", true)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != true {
		t.Errorf("expected: %v, but got: %v", true, value)
	}
}

func TestCache_SetGetByteSlice(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", []byte("hey"))
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if bytes.Compare(value.([]byte), []byte("hey")) != 0 {
		t.Errorf("expected: %v, but got: %v", []byte("hey"), value)
	}
}

func TestCache_SetGetStringSlice(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.Set("key", []string{"john", "doe"})
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value.([]string)[0] != "john" {
		t.Errorf("expected: %v, but got: %v", "john", value)
	}
	if value.([]string)[1] != "doe" {
		t.Errorf("expected: %v, but got: %v", "doe", value)
	}
}

func TestCache_SetGetStruct(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	type Custom struct {
		Int     int
		Uint    uint
		Float32 float32
		String  string
		Strings []string
		Nested  struct {
			String string
		}
	}
	cache.Set("key", Custom{
		Int:     111,
		Uint:    222,
		Float32: 123.456,
		String:  "hello",
		Strings: []string{"s1", "s2"},
		Nested:  struct{ String string }{String: "nested field"},
	})
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if ExpectedValue := 111; value.(Custom).Int != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := uint(222); value.(Custom).Uint != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := float32(123.456); value.(Custom).Float32 != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "hello"; value.(Custom).String != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "s1"; value.(Custom).Strings[0] != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "s2"; value.(Custom).Strings[1] != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "nested field"; value.(Custom).Nested.String != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
}

func TestCache_SetAll(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.SetAll(map[string]interface{}{"k1": "v1", "k2": "v2"})
	value, ok := cache.Get("k1")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "v1" {
		t.Errorf("expected: %s, but got: %s", "v1", value)
	}
	value, ok = cache.Get("k2")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "v2" {
		t.Errorf("expected: %s, but got: %s", "v2", value)
	}
	cache.SetAll(map[string]interface{}{"k1": "updated"})
	value, ok = cache.Get("k1")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "updated" {
		t.Errorf("expected: %s, but got: %s", "updated", value)
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.SetWithTTL("key", "value", NoExpiration)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
}

func TestCache_SetWithTTLWhenTTLIsNegative(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.SetWithTTL("key", "value", -12345)
	_, ok := cache.Get("key")
	if ok {
		t.Error("expected key to not exist, because there's no point in creating a cache entry that has a negative TTL")
	}
}

func TestCache_SetWithTTLWhenTTLIsZero(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.SetWithTTL("key", "value", 0)
	_, ok := cache.Get("key")
	if ok {
		t.Error("expected key to not exist, because there's no point in creating a cache entry that has a TTL of 0")
	}
}

func TestCache_SetWithTTLWhenTTLIsZeroAndEntryAlreadyExists(t *testing.T) {
	cache := NewCache(WithMaxSize(NoMaxSize))
	cache.SetWithTTL("key", "value", NoExpiration)
	cache.SetWithTTL("key", "value", 0)
	_, ok := cache.Get("key")
	if ok {
		t.Error("expected key to not exist, because there's the entry was created with a TTL of 0, so it should have been deleted immediately")
	}
}
