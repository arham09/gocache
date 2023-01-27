package gocache

import (
	"testing"
	"time"
)

func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", "hey")
	cache.Set("2", []byte("sup"))
	cache.Set("3", 123456)

	// (head) 3 - 2 - 1 (tail)
	if cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}

	cache.Delete("2")

	// (head) 3 - 1 (tail)
	if cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.tail.previous.Key != "3" {
		t.Error("The entry key previous to the cache tail should have been 3")
	}
	if cache.head.next.Key != "1" {
		t.Error("The entry key next to the cache tail should have been 1")
	}

	cache.Delete("1")

	// (head) 3 (tail)
	if cache.tail.Key != "3" {
		t.Error("cache tail should have been the entry with key 3")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}

	if cache.head != cache.tail {
		t.Error("There should only be one entry in the cache")
	}
	if cache.head.next != nil || cache.tail.previous != nil {
		t.Error("Since head == tail, there should be no next/prev")
	}
}

func TestCache_DeleteAll(t *testing.T) {
	cache := NewCache()
	cache.Set("1", []byte("1"))
	cache.Set("2", []byte("2"))
	cache.Set("3", []byte("3"))
	if len(cache.GetByKeys([]string{"1", "2", "3"})) != 3 {
		t.Error("Expected keys 1, 2 and 3 to exist")
	}
	numberOfDeletedKeys := cache.DeleteAll([]string{"1", "2", "3"})
	if numberOfDeletedKeys != 3 {
		t.Errorf("Expected 3 keys to have been deleted, but only %d were deleted", numberOfDeletedKeys)
	}
}

func TestCache_DeleteKeysByPattern(t *testing.T) {
	cache := NewCache()
	cache.Set("a1", []byte("v"))
	cache.Set("a2", []byte("v"))
	cache.Set("b1", []byte("v"))
	if len(cache.GetByKeys([]string{"a1", "a2", "b1"})) != 3 {
		t.Error("Expected keys 1, 2 and 3 to exist")
	}
	numberOfDeletedKeys := cache.DeleteKeysByPattern("a*")
	if numberOfDeletedKeys != 2 {
		t.Errorf("Expected 2 keys to have been deleted, but only %d were deleted", numberOfDeletedKeys)
	}
	if _, exists := cache.Get("b1"); !exists {
		t.Error("Expected key b1 to still exist")
	}
}

func TestCache_TTL(t *testing.T) {
	cache := NewCache()
	ttl, err := cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Errorf("expected %s, got %s", ErrKeyDoesNotExist, err)
	}
	cache.Set("key", "value")
	_, err = cache.TTL("key")
	if err != ErrKeyHasNoExpiration {
		t.Error("Expected TTL on new key created using Set to have no expiration")
	}
	cache.SetWithTTL("key", "value", time.Hour)
	ttl, err = cache.TTL("key")
	if err != nil {
		t.Error("Unexpected error")
	}
	if ttl.Minutes() < 59 || ttl.Minutes() > 60 {
		t.Error("Expected the TTL to be almost an hour")
	}
	cache.SetWithTTL("key", "value", 5*time.Millisecond)
	time.Sleep(6 * time.Millisecond)
	ttl, err = cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Error("key should've expired, thus TTL should've returned ")
	}
}

func TestCache_Expire(t *testing.T) {
	cache := NewCache()
	if cache.Expire("key-that-does-not-exist", time.Minute) {
		t.Error("Expected Expire to return false, because the key used did not exist")
	}
	cache.Set("key", "value")
	_, err := cache.TTL("key")
	if err != ErrKeyHasNoExpiration {
		t.Error("Expected TTL on new key created using Set to have no expiration")
	}
	if !cache.Expire("key", time.Hour) {
		t.Error("Expected Expire to return true")
	}
	ttl, err := cache.TTL("key")
	if err != nil {
		t.Error("Unexpected error")
	}
	if ttl.Minutes() < 59 || ttl.Minutes() > 60 {
		t.Error("Expected the TTL to be almost an hour")
	}
	if !cache.Expire("key", 5*time.Millisecond) {
		t.Error("Expected Expire to return true")
	}
	time.Sleep(6 * time.Millisecond)
	_, err = cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Error("key should've expired, thus TTL should've returned ErrKeyDoesNotExist")
	}
	if cache.Expire("key", time.Hour) {
		t.Error("Expire should've returned false, because the key should've already expired, thus no longer exist")
	}
	cache.SetWithTTL("key", "value", time.Hour)
	if !cache.Expire("key", NoExpiration) {
		t.Error("Expire should've returned true")
	}
	if _, err := cache.TTL("key"); err != ErrKeyHasNoExpiration {
		t.Error("TTL should've returned ErrKeyHasNoExpiration")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(WithMaxSize(10))
	cache.Set("k1", "v1")
	cache.Set("k2", "v2")
	cache.Set("k3", "v3")
	if cache.Count() != 3 {
		t.Error("expected cache size to be 3, got", cache.Count())
	}
	cache.Clear()
	if cache.Count() != 0 {
		t.Error("expected cache to be empty")
	}
	if cache.memoryUsage != 0 {
		t.Error("expected cache.memoryUsage to be 0")
	}
}
