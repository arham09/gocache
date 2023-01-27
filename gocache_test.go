package gocache

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewCache(t *testing.T) {
	cache := NewCache(WithMaxSize(1234), WithEvictionPolicy(LeastRecentlyUsed))
	if cache.MaxMemoryUsage() != NoMaxMemoryUsage {
		t.Error("shouldn't have a max memory usage configured")
	}
	if cache.EvictionPolicy() != LeastRecentlyUsed {
		t.Error("should've had a LeastRecentlyUsed eviction policy")
	}
	if cache.MaxSize() != 1234 {
		t.Error("should've had a max cache size of 1234")
	}
	if cache.MemoryUsage() != 0 {
		t.Error("should've had a memory usage of 0")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache(WithMaxSize(1234), WithEvictionPolicy(LeastRecentlyUsed))
	cache.Set("key", "value")
	if cache.Stats().Hits != 0 {
		t.Error("should have 0 hits")
	}
	if cache.Stats().Misses != 0 {
		t.Error("should have 0 misses")
	}
	cache.Get("key")
	if cache.Stats().Hits != 1 {
		t.Error("should have 1 hit")
	}
	if cache.Stats().Misses != 0 {
		t.Error("should have 0 misses")
	}
	cache.Get("key-that-does-not-exist")
	if cache.Stats().Hits != 1 {
		t.Error("should have 1 hit")
	}
	if cache.Stats().Misses != 1 {
		t.Error("should have 1 miss")
	}
}

func TestCache_WithMaxSize(t *testing.T) {
	cache := NewCache(WithMaxSize(1234))
	if cache.MaxSize() != 1234 {
		t.Error("expected cache to have a maximum size of 1234")
	}
}

func TestCache_WithMaxSizeAndNegativeValue(t *testing.T) {
	cache := NewCache(WithMaxSize(-10))
	if cache.MaxSize() != NoMaxSize {
		t.Error("expected cache to have no maximum size")
	}
}

func TestCache_WithMaxMemoryUsage(t *testing.T) {
	const ValueSize = Kilobyte
	cache := NewCache(WithMaxSize(0), WithMaxMemoryUsage(Kilobyte*64))
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("%d", i), strings.Repeat("0", ValueSize))
	}
	if cache.MemoryUsage()/1024 < 63 || cache.MemoryUsage()/1024 > 65 {
		t.Error("expected memoryUsage to be between 63KB and 64KB")
	}
}

func TestCache_WithMaxMemoryUsageWhenAddingAnEntryThatCausesMoreThanOneEviction(t *testing.T) {
	const ValueSize = Kilobyte
	cache := NewCache(WithMaxSize(0), WithMaxMemoryUsage(64*Kilobyte))
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("%d", i), strings.Repeat("0", ValueSize))
	}
	if cache.MemoryUsage()/1024 < 63 || cache.MemoryUsage()/1024 > 65 {
		t.Error("expected memoryUsage to be between 63KB and 64KB")
	}
}

func TestCache_WithMaxMemoryUsageAndNegativeValue(t *testing.T) {
	cache := NewCache(WithMaxSize(0), WithMaxMemoryUsage(-1234))
	if cache.MaxMemoryUsage() != NoMaxMemoryUsage {
		t.Error("attempting to set a negative max memory usage should force MaxMemoryUsage to NoMaxMemoryUsage")
	}
}

func TestCache_MemoryUsageAfterSet10000AndDelete5000(t *testing.T) {
	const ValueSize = 64
	cache := NewCache(WithMaxSize(10000), WithMaxMemoryUsage(Gigabyte))
	for i := 0; i < cache.maxSize; i++ {
		cache.Set(fmt.Sprintf("%05d", i), strings.Repeat("0", ValueSize))
	}
	memoryUsageBeforeDeleting := cache.MemoryUsage()
	for i := 0; i < cache.maxSize/2; i++ {
		key := fmt.Sprintf("%05d", i)
		cache.Delete(key)
	}
	memoryUsageRatio := float32(cache.MemoryUsage()) / float32(memoryUsageBeforeDeleting)
	if memoryUsageRatio != 0.5 {
		t.Error("Since half of the keys were deleted, the memoryUsage should've been half of what the memory usage was before beginning deletion")
	}
}

func TestCache_MemoryUsageIsReliable(t *testing.T) {
	cache := NewCache(WithMaxMemoryUsage(Megabyte))
	previousCacheMemoryUsage := cache.MemoryUsage()
	if previousCacheMemoryUsage != 0 {
		t.Error("cache.MemoryUsage() should've been 0")
	}
	cache.Set("1", 1)
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.SetAll(map[string]interface{}{"2": "2", "3": "3", "4": "4"})
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("2")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", 1)
	if cache.MemoryUsage() != previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() shouldn't have changed, because the entry didn't change")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("3")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("4")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("1")
	if cache.MemoryUsage() >= previousCacheMemoryUsage || cache.memoryUsage != 0 {
		t.Error("cache.MemoryUsage() should've been 0")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", "v4lu3")
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", "value")
	if cache.MemoryUsage() != previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() shouldn't have changed")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", true)
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased, because a bool uses less memory than a string")
	}
}

func TestCache_WithForceNilInterfaceOnNilPointer(t *testing.T) {
	type Struct struct{}
	cache := NewCache(WithForceNilInterfaceOnNilPointer(true))
	cache.Set("key", (*Struct)(nil))
	if value, exists := cache.Get("key"); !exists {
		t.Error("expected key to exist")
	} else {
		if value != nil {
			// the value is not nil, because cache.Get returns an interface{}, and the type of that interface is not nil
			t.Error("value should be nil")
		}
	}

	cache.Clear()

	cache.forceNilInterfaceOnNilPointer = false
	cache.Set("key", (*Struct)(nil))
	if value, exists := cache.Get("key"); !exists {
		t.Error("expected key to exist")
	} else {
		if value == nil {
			t.Error("value should be not be nil, because the type of the interface is not nil")
		}
		if value.(*Struct) != nil {
			t.Error("casted value should be nil")
		}
	}
}

func TestEvictionWhenThereIsNothingToEvict(t *testing.T) {
	cache := NewCache()
	cache.evict()
	cache.evict()
	cache.evict()
}

func TestCache(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(LeastRecentlyUsed))
	cache.Set("1", 1)
	cache.Set("2", 2)
	cache.Set("3", 3)
	cache.Set("4", 4)
	if _, ok := cache.Get("4"); !ok {
		t.Error("expected 4 to exist")
	}
	if _, ok := cache.Get("3"); !ok {
		t.Error("expected 3 to exist")
	}
	if _, ok := cache.Get("2"); !ok {
		t.Error("expected 2 to exist")
	}
	if _, ok := cache.Get("1"); ok {
		t.Error("expected 1 to have been evicted")
	}
	cache.Set("5", 5)
	if _, ok := cache.Get("1"); ok {
		t.Error("expected 1 to have been evicted")
	}
	if _, ok := cache.Get("2"); !ok {
		t.Error("expected 2 to exist")
	}
	if _, ok := cache.Get("3"); !ok {
		t.Error("expected 3 to exist")
	}
	if _, ok := cache.Get("4"); ok {
		t.Error("expected 4 to have been evicted")
	}
	if _, ok := cache.Get("5"); !ok {
		t.Error("expected 5 to exist")
	}
}
