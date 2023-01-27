package gocache

import (
	"fmt"
	"testing"
)

func TestCache_EvictionsRespectMaxSize(t *testing.T) {
	cache := NewCache(WithMaxSize(5))
	for n := 0; n < 10; n++ {
		cache.Set(fmt.Sprintf("test_%d", n), []byte("value"))
	}
	count := cache.Count()
	if count > 5 {
		t.Error("Max size was set to 5, but the cache size reached a size of", count)
	}
}

func TestCache_EvictionsWithFIFO(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(FirstInFirstOut))

	cache.Set("1", []byte("value"))
	cache.Set("2", []byte("value"))
	cache.Set("3", []byte("value"))
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))
	_, ok := cache.Get("1")
	if ok {
		t.Error("expected key 1 to have been removed, because FIFO")
	}
}

func TestCache_EvictionsWithLRU(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(LeastRecentlyUsed))

	cache.Set("1", []byte("value"))
	cache.Set("2", []byte("value"))
	cache.Set("3", []byte("value"))
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_EvictionsWithLFU(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(LeastFrequentUsed))

	cache.Set("1", []byte("value"))
	cache.Set("2", []byte("value"))
	cache.Set("3", []byte("value"))
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_HeadTailWorksWithFIFO(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(FirstInFirstOut))

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", []byte("value"))

	// (head) 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "1" {
		t.Error("cache head should have been entry with key 1")
	}

	cache.Set("2", []byte("value"))

	// (head) 2 - 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "2" {
		t.Error("cache head should have been the entry with key 2")
	}
	if cache.head.next.Key != "1" {
		t.Error("The entry key next to the cache head should have been 1")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.tail.previous.Key != "2" {
		t.Error("The entry key previous to the cache tail should have been 2")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}

	cache.Set("3", []byte("value"))

	// (head) 3 - 2 - 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.tail.previous.Key != "2" {
		t.Error("The entry key previous to the cache tail should have been 2")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}
	if cache.head == nil || cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.head.next.Key != "2" {
		t.Error("The entry key next to the cache head should have been 2")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.head.next.previous.Key != "3" {
		t.Error("The head's next node should have its previous node pointing to the cache head")
	}
	if cache.head.next.next.Key != "1" {
		t.Error("The head's next node should have its next node pointing to the cache tail")
	}

	// Get the first entry. This doesn't change anything for FIFO, but for LRU, it would mean that retrieved entry
	// wouldn't be evicted since it was recently accessed. Basically, we just want to make sure that FIFO works
	// as intended (i.e. not like LRU)
	_, _ = cache.Get("1")

	cache.Set("4", []byte("value"))

	// (head) 4 - 3 - 2 (tail)
	_, ok := cache.Get("1")
	if ok {
		t.Error("expected key 1 to have been removed, because FIFO")
	}
	if cache.tail == nil || cache.tail.Key != "2" {
		t.Error("cache tail should have been the entry with key 2")
	}
	if cache.tail.previous.Key != "3" {
		t.Error("The entry key previous to the cache tail should have been 3")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}
	if cache.head == nil || cache.head.Key != "4" {
		t.Error("cache head should have been the entry with key 4")
	}
	if cache.head.next.Key != "3" {
		t.Error("The entry key next to the cache head should have been 3")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.head.next.previous.Key != "4" {
		t.Error("The head's next node should have its previous node pointing to the cache head")
	}
	if cache.head.next.next.Key != "2" {
		t.Error("The head's next node should have its next node pointing to the cache tail")
	}
}

func TestCache_HeadTailWorksWithLRU(t *testing.T) {
	cache := NewCache(WithMaxSize(3), WithEvictionPolicy(LeastRecentlyUsed))

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", []byte("value"))

	// (head) 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "1" {
		t.Error("cache head should have been entry with key 1")
	}

	cache.Set("2", []byte("value"))

	// (head) 2 - 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "2" {
		t.Error("cache head should have been the entry with key 2")
	}
	if cache.head.next.Key != "1" {
		t.Error("The entry key next to the cache head should have been 1")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.tail.previous.Key != "2" {
		t.Error("The entry key previous to the cache tail should have been 2")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}

	cache.Set("3", []byte("value"))

	// (head) 3 - 2 - 1 (tail)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.tail.previous.Key != "2" {
		t.Error("The entry key previous to the cache tail should have been 2")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}
	if cache.head == nil || cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.head.next.Key != "2" {
		t.Error("The entry key next to the cache head should have been 2")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.head.next.previous.Key != "3" {
		t.Error("The head's next node should have its previous node pointing to the cache head")
	}
	if cache.head.next.next.Key != "1" {
		t.Error("The head's next node should have its next node pointing to the cache tail")
	}

	// Because we're using a LRU cache, this should cause 1 to get moved back to the head, thus
	// moving it from the tail.
	// In other words, because we retrieved the key 1 here, this is no longer the least recently used cache entry,
	// which means it will not be evicted during the next insertion.
	_, _ = cache.Get("1")

	// (head) 1 - 3 - 2 (tail) (This updated because LRU)
	cache.Set("4", []byte("value"))

	// (head) 4 - 1 - 3 (tail)
	if cache.tail == nil || cache.tail.Key != "3" {
		t.Error("cache tail should have been the entry with key 3")
	}
	if cache.tail.previous.Key != "1" {
		t.Error("The entry key previous to the cache tail should have been 1")
	}
	if cache.tail.next != nil {
		t.Error("The cache tail should not have a next node")
	}
	if cache.head == nil || cache.head.Key != "4" {
		t.Error("cache head should have been the entry with key 4")
	}
	if cache.head.next.Key != "1" {
		t.Error("The entry key next to the cache head should have been 1")
	}
	if cache.head.previous != nil {
		t.Error("The cache head should not have a previous node")
	}
	if cache.head.next.previous.Key != cache.head.Key {
		t.Error("The head's next node should have its previous node pointing to the cache head")
	}
	if cache.head.next.next.Key != cache.tail.Key {
		t.Error("Should be able to walk from head to tail")
	}
	if cache.tail.previous.previous != cache.head {
		t.Error("Should be able to walk from tail to head")
	}

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_HeadStaysTheSameIfCallRepeatedly(t *testing.T) {
	cache := NewCache(WithEvictionPolicy(LeastRecentlyUsed), WithMaxSize(10))
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Get("1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Get("1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
}

func TestCache_HeadToTailSimple(t *testing.T) {
	cache := NewCache(WithMaxSize(3))
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Set("2", "2")
	if cache.tail.Key != "1" && cache.head.Key != "2" {
		t.Error("expected tail=1 and head=2")
	}
	cache.Set("3", "3")
	if cache.tail.Key != "1" && cache.head.Key != "3" {
		t.Error("expected tail=1 and head=4")
	}
	cache.Set("4", "4")
	if cache.tail.Key != "2" && cache.head.Key != "4" {
		t.Error("expected tail=2 and head=4")
	}
	cache.Set("5", "5")
	if cache.tail.Key != "3" && cache.head.Key != "5" {
		t.Error("expected tail=3 and head=5")
	}
}
