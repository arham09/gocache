package gocache

import "time"

// Delete removes a key from the cache
//
// Returns false if the key did not exist.
func (c *Cache) Delete(key string) bool {
	c.mutex.Lock()
	ok := c.delete(key)
	c.mutex.Unlock()
	return ok
}

// DeleteAll deletes multiple entries based on the keys passed as parameter
//
// Returns the number of keys deleted
func (c *Cache) DeleteAll(keys []string) int {
	numberOfKeysDeleted := 0
	c.mutex.Lock()
	for _, key := range keys {
		if c.delete(key) {
			numberOfKeysDeleted++
		}
	}
	c.mutex.Unlock()
	return numberOfKeysDeleted
}

// DeleteKeysByPattern deletes all entries matching a given key pattern and returns the number of entries deleted.
//
// Note that DeleteKeysByPattern does not trigger active evictions, nor does it count as accessing the entry (if LRU).
func (c *Cache) DeleteKeysByPattern(pattern string) int {
	return c.DeleteAll(c.GetKeysByPattern(pattern, 0))
}

// Count returns the total amount of entries in the cache, regardless of whether they're expired or not
func (c *Cache) Count() int {
	c.mutex.RLock()
	count := len(c.entries)
	c.mutex.RUnlock()
	return count
}

// Clear deletes all entries from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	c.entries = make(map[string]*Entry)
	c.memoryUsage = 0
	c.head = nil
	c.tail = nil
	c.mutex.Unlock()
}

// TTL returns the time until the cache entry specified by the key passed as parameter
// will be deleted.
func (c *Cache) TTL(key string) (time.Duration, error) {
	c.mutex.RLock()
	entry, ok := c.get(key)
	c.mutex.RUnlock()
	if !ok {
		return 0, ErrKeyDoesNotExist
	}
	if entry.Expiration == NoExpiration {
		return 0, ErrKeyHasNoExpiration
	}
	timeUntilExpiration := time.Until(time.Unix(0, entry.Expiration))
	if timeUntilExpiration < 0 {
		// The key has already expired but hasn't been deleted yet.
		// From the client's perspective, this means that the c entry doesn't exist
		return 0, ErrKeyDoesNotExist
	}
	return timeUntilExpiration, nil
}

// Expire sets a key's expiration time
//
// A TTL of -1 means that the key will never expire
// A TTL of 0 means that the key will expire immediately
// If using LRU, note that this does not reset the position of the key
//
// Returns true if the cache key exists and has had its expiration time altered
func (c *Cache) Expire(key string, ttl time.Duration) bool {
	entry, ok := c.get(key)
	if !ok || entry.Expired() {
		return false
	}
	if ttl != NoExpiration {
		entry.Expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.Expiration = NoExpiration
	}
	return true
}

func (c *Cache) delete(key string) bool {
	entry, ok := c.entries[key]
	if ok {
		if c.maxMemoryUsage != NoMaxMemoryUsage {
			c.memoryUsage -= entry.SizeInBytes()
		}

		if c.evictionPolicy == LeastFrequentUsed {
			c.removeEntryFromFrequencyList(entry.frequencyParent, entry)
		}

		c.removeExistingEntryReferences(entry)
		delete(c.entries, key)

	}
	return ok
}
