package gocache

import (
	"reflect"
	"time"
)

// Set creates or updates a key with a given value
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, NoExpiration)
}

// SetWithTTL creates or updates a key with a given value and sets an expiration time (-1 is NoExpiration)
//
// The TTL provided must be greater than 0, or NoExpiration (-1). If a negative value that isn't -1 (NoExpiration) is
// provided, the entry will not be created if the key doesn't exist
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	// An interface is only nil if both its value and its type are nil, however, passing a nil pointer as an interface{}
	// means that the interface itself is not nil, because the interface value is nil but not the type.
	if c.forceNilInterfaceOnNilPointer {
		if value != nil && (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
			value = nil
		}
	}

	c.mutex.Lock()
	entry, ok := c.get(key)
	if !ok {
		// A negative TTL that isn't -1 (NoExpiration) or 0 is an entry that will expire instantly,
		// so might as well just not create it in the first place
		if ttl != NoExpiration && ttl < 1 {
			c.mutex.Unlock()
			return
		}
		// Cache entry doesn't exist, so we have to create a new one
		entry = &Entry{
			Key:               key,
			Value:             value,
			RelevantTimestamp: time.Now(),
			next:              c.head,
		}
		if c.head == nil {
			c.tail = entry
		} else {
			c.head.previous = entry
		}
		c.head = entry
		c.entries[key] = entry
		if c.maxMemoryUsage != NoMaxMemoryUsage {
			c.memoryUsage += entry.SizeInBytes()
		}
	} else {
		// A negative TTL that isn't -1 (NoExpiration) or 0 is an entry that will expire instantly,
		// so might as well just delete it immediately instead of updating it
		if ttl != NoExpiration && ttl < 1 {
			c.delete(key)
			c.mutex.Unlock()
			return
		}
		if c.maxMemoryUsage != NoMaxMemoryUsage {
			// Subtract the old entry from the cache's memoryUsage
			c.memoryUsage -= entry.SizeInBytes()
		}
		// Update existing entry's value
		entry.Value = value
		entry.RelevantTimestamp = time.Now()
		if c.maxMemoryUsage != NoMaxMemoryUsage {
			// Add the memory usage of the new entry to the cache's memoryUsage
			c.memoryUsage += entry.SizeInBytes()
		}
		// Because we just updated the entry, we need to move it back to HEAD
		c.moveExistingEntryToHead(entry)
	}
	if ttl != NoExpiration {
		entry.Expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.Expiration = NoExpiration
	}
	// If the cache doesn't have a maxSize/maxMemoryUsage, then there's no point
	// checking if we need to evict an entry, so we'll just return now
	if c.maxSize == NoMaxSize && c.maxMemoryUsage == NoMaxMemoryUsage {
		c.mutex.Unlock()
		return
	}
	// If there's a maxSize and the cache has more entries than the maxSize, evict
	if c.maxSize != NoMaxSize && len(c.entries) > c.maxSize {
		c.evict()
	}
	// If there's a maxMemoryUsage and the memoryUsage is above the maxMemoryUsage, evict
	if c.maxMemoryUsage != NoMaxMemoryUsage && c.memoryUsage > c.maxMemoryUsage {
		for c.memoryUsage > c.maxMemoryUsage && len(c.entries) > 0 {
			c.evict()
		}
	}

	if c.evictionPolicy == LeastFrequentUsed {
		c.incrementEntryFrequency(entry)
	}
	c.mutex.Unlock()
}

// SetAll creates or updates multiple values
func (c *Cache) SetAll(entries map[string]interface{}) {
	for key, value := range entries {
		c.SetWithTTL(key, value, NoExpiration)
	}
}
