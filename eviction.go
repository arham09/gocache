package gocache

// moveExistingEntryToHead replaces the current c head for an existing entry
func (c *Cache) moveExistingEntryToHead(entry *Entry) {
	if !(entry == c.head && entry == c.tail) {
		c.removeExistingEntryReferences(entry)
	}
	if entry != c.head {
		entry.next = c.head
		entry.previous = nil
		if c.head != nil {
			c.head.previous = entry
		}
		c.head = entry
	}
}

// removeExistingEntryReferences modifies the next and previous reference of an existing entry and re-links
// the next and previous entry accordingly, as well as the cache head or/and the cache tail if necessary.
// Note that it does not remove the entry from the cache, only the references.
func (c *Cache) removeExistingEntryReferences(entry *Entry) {
	if c.tail == entry && c.head == entry {
		c.tail = nil
		c.head = nil
	} else if c.tail == entry {
		c.tail = c.tail.previous
	} else if c.head == entry {
		c.head = c.head.next
	}
	if entry.previous != nil {
		entry.previous.next = entry.next
	}
	if entry.next != nil {
		entry.next.previous = entry.previous
	}
	entry.next = nil
	entry.previous = nil
}

// evict removes the tail from the cache
func (c *Cache) evict() {
	if c.tail == nil || len(c.entries) == 0 {
		return
	}

	if c.evictionPolicy == LeastFrequentUsed {
		if item := c.freqs.Front(); item != nil {
			for entry, _ := range item.Value.(*FrequencyItem).Entries {
				oldEntry := entry
				c.removeExistingEntryReferences(oldEntry)
				delete(c.entries, oldEntry.Key)
				c.removeEntryFromFrequencyList(item, entry)
				c.stats.EvictedKeys++
				if c.maxMemoryUsage != NoMaxMemoryUsage {
					c.memoryUsage -= oldEntry.SizeInBytes()
				}
			}
		}
		return
	}

	if c.tail != nil {
		oldTail := c.tail
		c.removeExistingEntryReferences(oldTail)
		delete(c.entries, oldTail.Key)
		if c.maxMemoryUsage != NoMaxMemoryUsage {
			c.memoryUsage -= oldTail.SizeInBytes()
		}
		c.stats.EvictedKeys++
	}
}
