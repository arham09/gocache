package gocache

import (
	"container/list"
)

type FrequencyItem struct {
	Entries map[*Entry]byte // Set of entries
	Freq    int             // Access frequency
}

func (c *Cache) incrementEntryFrequency(entry *Entry) {
	var (
		currentFrequency    = entry.frequencyParent
		nextFrequencyAmount int
		nextFrequency       *list.Element
	)

	// if current frequency is nil, we will create with frequency 1
	if currentFrequency == nil {
		nextFrequencyAmount = 1
		nextFrequency = c.freqs.Front()
	} else {
		// set the next frequency amount to current + 1, since we need to increment the current entry
		// frequency and move that to the +1 key in the list
		nextFrequencyAmount = currentFrequency.Value.(*FrequencyItem).Freq + 1
		nextFrequency = currentFrequency.Next()
	}

	// if nextFrequency doesnt exist or the key isnt same as the nextFrequencyAmount
	// we will create a new key for the entry
	if nextFrequency == nil || nextFrequency.Value.(*FrequencyItem).Freq != nextFrequencyAmount {
		newFrequencyItem := new(FrequencyItem)
		newFrequencyItem.Freq = nextFrequencyAmount
		newFrequencyItem.Entries = make(map[*Entry]byte)
		if currentFrequency == nil {
			nextFrequency = c.freqs.PushFront(newFrequencyItem)
		} else {
			nextFrequency = c.freqs.InsertAfter(newFrequencyItem, currentFrequency)
		}
	}

	entry.frequencyParent = nextFrequency
	nextFrequency.Value.(*FrequencyItem).Entries[entry] = 1

	if currentFrequency != nil {
		c.removeEntryFromFrequencyList(currentFrequency, entry)
	}
}

func (c *Cache) removeEntryFromFrequencyList(listItem *list.Element, item *Entry) {
	frequencyItem := listItem.Value.(*FrequencyItem)

	// delete entry in the frequency list
	delete(frequencyItem.Entries, item)

	// if no other cache in the frequency list, remove the frequency
	if len(frequencyItem.Entries) == 0 {
		c.freqs.Remove(listItem)
	}
}
