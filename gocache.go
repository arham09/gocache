package gocache

import (
	"container/list"
	"errors"
	"sync"
)

var (
	Debug = false
)

const (
	// NoMaxSize means that the c has no maximum number of entries in the c
	// Setting Cache.maxSize to this value also means there will be no eviction
	NoMaxSize = 0

	// NoMaxMemoryUsage means that the c has no maximum number of entries in the c
	NoMaxMemoryUsage = 0

	// DefaultMaxSize is the max size set if no max size is specified
	DefaultMaxSize = 100000

	// NoExpiration is the value that must be used as TTL to specify that the given key should never expire
	NoExpiration = -1

	Kilobyte = 1024
	Megabyte = 1024 * Kilobyte
	Gigabyte = 1024 * Megabyte
)

var (
	ErrKeyDoesNotExist       = errors.New("key does not exist")         // Returned when a c key does not exist
	ErrKeyHasNoExpiration    = errors.New("key has no expiration")      // Returned when a c key has no expiration
	ErrJanitorAlreadyRunning = errors.New("janitor is already running") // Returned when the janitor has already been started
)

// Cache is the core struct of gocache which contains the data as well as all relevant configuration fields
type Cache struct {
	// maxSize is the maximum amount of entries that can be in the c at any given time
	// By default, this is set to DefaultMaxSize
	maxSize int

	// maxMemoryUsage is the maximum amount of memory that can be taken up by the c at any time
	// By default, this is set to NoMaxMemoryUsage, meaning that the default behavior is to not evict
	// based on maximum memory usage
	maxMemoryUsage int

	// evictionPolicy is the eviction policy
	evictionPolicy EvictionPolicy

	// stats is the object that contains c statistics/metrics
	stats *Statistics

	// entries is the content of the c
	entries map[string]*Entry

	// mutex is the lock for making concurrent operations on the c
	mutex sync.RWMutex

	// head is the c entry at the head of the c
	head *Entry

	// tail is the last c node and also the next entry that will be evicted
	tail *Entry

	// freqs is used to count how frequent is the entry used
	freqs *list.List

	// stopJanitor is the channel used to stop the janitor
	stopJanitor chan bool

	// memoryUsage is the approximate memory usage of the c (dataset only) in bytes
	memoryUsage int

	// forceNilInterfaceOnNilPointer determines whether all Set-like functions should set a value as nil if the
	// interface passed has a nil value but not a nil type.
	//
	// By default, interfaces are only nil when both their type and value is nil.
	// This means that when you pass a pointer to a nil value, the type of the interface
	// will still show as nil, which means that if you don't cast the interface after
	// retrieving it, a nil check will return that the value is not false.
	forceNilInterfaceOnNilPointer bool
}

// MaxSize returns the maximum amount of keys that can be present in the cache before
// new entries trigger the eviction of the tail
func (c *Cache) MaxSize() int {
	return c.maxSize
}

// MaxMemoryUsage returns the configured maxMemoryUsage of the cache
func (c *Cache) MaxMemoryUsage() int {
	return c.maxMemoryUsage
}

// EvictionPolicy returns the EvictionPolicy of the Cache
func (c *Cache) EvictionPolicy() EvictionPolicy {
	return c.evictionPolicy
}

// Stats returns statistics from the cache
func (c *Cache) Stats() Statistics {
	c.mutex.RLock()
	stats := Statistics{
		EvictedKeys: c.stats.EvictedKeys,
		ExpiredKeys: c.stats.ExpiredKeys,
		Hits:        c.stats.Hits,
		Misses:      c.stats.Misses,
	}
	c.mutex.RUnlock()
	return stats
}

// MemoryUsage returns the current memory usage of the cache's dataset in bytes
// If MaxMemoryUsage is set to NoMaxMemoryUsage, this will return 0
func (c *Cache) MemoryUsage() int {
	return c.memoryUsage
}

// WithMaxMemoryUsage sets the maximum amount of memory that can be used by the cache at any given time
//
// NOTE: This is approximate.
//
// // Setting this to NoMaxMemoryUsage will disable eviction by memory usage
func WithMaxMemoryUsage(maxMemoryUsageInBytes int) func(c *Cache) {
	return func(c *Cache) {
		if maxMemoryUsageInBytes < 0 {
			maxMemoryUsageInBytes = NoMaxMemoryUsage
		}
		c.maxMemoryUsage = maxMemoryUsageInBytes
	}
}

// WithMaxSize sets the maximum amount of entries that can be in the cache at any given time
// A maxSize of 0 or less means infinite
func WithMaxSize(maxSize int) func(c *Cache) {
	return func(c *Cache) {
		if maxSize < 0 {
			maxSize = NoMaxSize
		}
		if maxSize != NoMaxSize && c.Count() == 0 {
			c.entries = make(map[string]*Entry, maxSize)
		}
		c.maxSize = maxSize
	}
}

// WithEvictionPolicy sets eviction algorithm.
// Defaults to FirstInFirstOut (FIFO)
func WithEvictionPolicy(policy EvictionPolicy) func(c *Cache) {
	return func(c *Cache) {
		if policy == LeastFrequentUsed {
			c.freqs = list.New()
		}
		c.evictionPolicy = policy
	}
}

// WithForceNilInterfaceOnNilPointer sets whether all Set-like functions should set a value as nil if the
// interface passed has a nil value but not a nil type.
//
// In Go, an interface is only nil if both its type and value are nil, which means that a nil pointer
// (e.g. (*Struct)(nil)) will retain its attribution to the type, and the unmodified value returned from
// Cache.Get, for instance, would return false when compared with nil if this option is set to false.
//
// We can bypass this by detecting if the interface's value is nil and setting it to nil rather than
// a nil pointer, which will make the value returned from Cache.Get return true when compared with nil.
// This is exactly what passing true to WithForceNilInterfaceOnNilPointer does, and it's also the default behavior.
//
// Alternatively, you may pass false to WithForceNilInterfaceOnNilPointer, which will mean that you'll have
// to cast the value returned from Cache.Get to its original type to check for whether the pointer returned
// is nil or not.
//
// If set to true (default):
//     c := gocache.NewCache(WithForceNilInterfaceOnNilPointer(true))
//     c.Set("key", (*Struct)(nil))
//     value, _ := c.Get("key")
//     // the following returns true, because the interface{} was forcefully set to nil
//     if value == nil {}
//     // the following will panic, because the value has been casted to its type (which is nil)
//     if value.(*Struct) == nil {}
//
// If set to false:
//     c := gocache.NewCache(WithForceNilInterfaceOnNilPointer(false))
//     c.Set("key", (*Struct)(nil))
//     value, _ := c.Get("key")
//     // the following returns false, because the interface{} returned has a non-nil type (*Struct)
//     if value == nil {}
//     // the following returns true, because the value has been casted to its type
//     if value.(*Struct) == nil {}
//
// In other words, if set to true, you do not need to cast the value returned from the cache to
// to check if the value is nil.
//
// Defaults to true
func WithForceNilInterfaceOnNilPointer(forceNilInterfaceOnNilPointer bool) func(c *Cache) {
	return func(c *Cache) {
		c.forceNilInterfaceOnNilPointer = forceNilInterfaceOnNilPointer
	}
}

// NewCache creates a new Cache
func NewCache(opts ...func(*Cache)) *Cache {
	c := &Cache{
		maxSize:                       DefaultMaxSize,
		evictionPolicy:                FirstInFirstOut,
		stats:                         &Statistics{},
		entries:                       make(map[string]*Entry),
		mutex:                         sync.RWMutex{},
		stopJanitor:                   nil,
		forceNilInterfaceOnNilPointer: true,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}
