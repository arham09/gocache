package gocache

// Get retrieves an entry using the key passed as parameter
// If there is no such entry, the value returned will be nil and the boolean will be false
// If there is an entry, the value returned will be the value cached and the boolean will be true
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	entry, ok := c.get(key)
	if !ok {
		c.stats.Misses++
		c.mutex.Unlock()
		return nil, false
	}
	if entry.Expired() {
		c.stats.ExpiredKeys++
		c.delete(key)
		c.mutex.Unlock()
		return nil, false
	}
	c.stats.Hits++
	if c.evictionPolicy == LeastRecentlyUsed {
		entry.Accessed()
		if c.head == entry {
			c.mutex.Unlock()
			return entry.Value, true
		}
		// Because the eviction policy is LRU, we need to move the entry back to HEAD
		c.moveExistingEntryToHead(entry)
	}

	if c.evictionPolicy == LeastFrequentUsed {
		c.incrementEntryFrequency(entry)
	}
	c.mutex.Unlock()
	return entry.Value, true
}

// GetValue retrieves an entry using the key passed as parameter
// Unlike Get, this function only returns the value
func (c *Cache) GetValue(key string) interface{} {
	value, _ := c.Get(key)
	return value
}

// GetByKeys retrieves multiple entries using the keys passed as parameter
// All keys are returned in the map, regardless of whether they exist or not, however, entries that do not exist in the
// cache will return nil, meaning that there is no way of determining whether a key genuinely has the value nil, or
// whether it doesn't exist in the cache using only this function.
func (c *Cache) GetByKeys(keys []string) map[string]interface{} {
	entries := make(map[string]interface{})
	for _, key := range keys {
		entries[key], _ = c.Get(key)
	}
	return entries
}

// GetAll retrieves all cache entries
//
// If the eviction policy is LeastRecentlyUsed, note that unlike Get and GetByKeys, this does not update the last access
// timestamp. The reason for this is that since all cache entries will be accessed, updating the last access timestamp
// would provide very little benefit while harming the ability to accurately determine the next key that will be evicted
//
// You should probably avoid using this if you have a lot of entries.
//
// GetKeysByPattern is a good alternative if you want to retrieve entries that you do not have the key for, as it only
// retrieves the keys and does not trigger active eviction and has a parameter for setting a limit to the number of keys
// you wish to retrieve.
func (c *Cache) GetAll() map[string]interface{} {
	entries := make(map[string]interface{})
	c.mutex.Lock()
	for key, entry := range c.entries {
		if entry.Expired() {
			c.delete(key)
			continue
		}
		entries[key] = entry.Value
	}
	c.stats.Hits += uint64(len(entries))
	c.mutex.Unlock()
	return entries
}

// GetKeysByPattern retrieves a slice of keys that match a given pattern
// If the limit is set to 0, the entire cache will be searched for matching keys.
// If the limit is above 0, the search will stop once the specified number of matching keys have been found.
//
// e.g.
//     c.GetKeysByPattern("*some*", 0) will return all keys containing "some" in them
//     c.GetKeysByPattern("*some*", 5) will return 5 keys (or less) containing "some" in them
//
// Note that GetKeysByPattern does not trigger active evictions, nor does it count as accessing the entry (if LRU).
// The reason for that behavior is that these two (active eviction and access) only applies when you access the value
// of the cache entry, and this function only returns the keys.
func (c *Cache) GetKeysByPattern(pattern string, limit int) []string {
	var matchingKeys []string
	c.mutex.Lock()
	for key, value := range c.entries {
		if value.Expired() {
			continue
		}
		if MatchPattern(pattern, key) {
			matchingKeys = append(matchingKeys, key)
			if limit > 0 && len(matchingKeys) >= limit {
				break
			}
		}
	}
	c.mutex.Unlock()
	return matchingKeys
}

// get retrieves an entry using the key passed as parameter, but unlike Get, it doesn't update the access time or
// move the position of the entry to the head
func (c *Cache) get(key string) (*Entry, bool) {
	entry, ok := c.entries[key]
	return entry, ok
}
