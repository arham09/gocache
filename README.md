# cache
<!-- [![test](https://github.com/arham09/cache/workflows/test/badge.svg?branch=master) -->
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/TwiN/cache)](https://goreportcard.com/report/github.com/TwiN/cache) -->
<!-- [![codecov](https://codecov.io/gh/TwiN/cache/branch/master/graph/badge.svg)](https://codecov.io/gh/TwiN/cache) -->
<!-- [![Go version](https://img.shields.io/github/go-mod/go-version/arham09/cache.svg)](https://github.com/arham09/cache) -->
[![Follow arham09](https://img.shields.io/github/followers/arham09?label=Follow&style=social)](https://github.com/arham09)

cache was created to learn about cache eviction, heavily extracted from [gocache](https://github.com/TwiN/gocache) and adding LFU policy for more eviction, if you want to use it please test it carefully, cache is easy-to-use, high-performance, lightweight and thread-safe (goroutine-safe) in-memory key-value cache with support for LFU, LRU and FIFO eviction policies as well as expiration, bulk operations and even retrieval of keys by pattern.


## Table of Contents

- [Features](#features)
- [Usage](#usage)
  - [Initializing the cache](#initializing-the-cache)
  - [Functions](#functions)
  - [Examples](#examples)
    - [Creating or updating an entry](#creating-or-updating-an-entry)
    - [Getting an entry](#getting-an-entry)
    - [Deleting an entry](#deleting-an-entry)
    - [Complex example](#complex-example)
- [Eviction](#eviction)
  - [MaxSize](#maxsize)
  - [MaxMemoryUsage](#maxmemoryusage)
- [Expiration](#expiration)
- [Performance](#performance)
  - [Summary](#summary)
  - [Results](#results)


## Features
cache supports the following cache eviction policies: 
- First in first out (FIFO)
- Least recently used (LRU)
- Least frequent used (LFU)

It also supports cache entry TTL, which is both active and passive. Active expiration means that if you attempt 
to retrieve a cache key that has already expired, it will delete it on the spot and the behavior will be as if
the cache key didn't exist. As for passive expiration, there's a background task that will take care of deleting
expired keys.

It also includes what you'd expect from a cache, like GET/SET, bulk operations and get by pattern.


## Usage
```
go get -u github.com/arham09/cache
```


### Initializing the cache
```go
c := cache.NewCache(WithMaxSize(1000), WithEvictionPolicy(cache.LeastRecentlyUsed))
```

If you're planning on using expiration (`SetWithTTL` or `Expire`) and you want expired entries to be automatically deleted 
in the background, make sure to start the janitor when you instantiate the cache:

```go
cache.StartJanitor()
```

### Functions
| Function                          | Description                                                                                                                                                                                                                                                        |
|-----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| WithMaxSize                       | Sets the max size of the cache. `cache.NoMaxSize` means there is no limit. If not set, the default max size is `cache.DefaultMaxSize`.                                                                                                                         |
| WithMaxMemoryUsage                | Sets the max memory usage of the cache. `cache.NoMaxMemoryUsage` means there is no limit. The default behavior is to not evict based on memory usage.                                                                                                            |
| WithEvictionPolicy                | Sets the eviction algorithm to be used when the cache reaches the max size. If not set, the default eviction policy is `cache.FirstInFirstOut` (FIFO).                                                                                                           |
| WithForceNilInterfaceOnNilPointer | Configures whether values with a nil pointer passed to write functions should be forcefully set to nil. Defaults to true.                                                                                                                                          |
| StartJanitor                      | Starts the janitor, which is in charge of deleting expired cache entries in the background.                                                                                                                                                                        |
| StopJanitor                       | Stops the janitor.                                                                                                                                                                                                                                                 |
| Set                               | Same as `SetWithTTL`, but with no expiration (`cache.NoExpiration`)                                                                                                                                                                                              |
| SetAll                            | Same as `Set`, but in bulk                                                                                                                                                                                                                                         |
| SetWithTTL                        | Creates or updates a cache entry with the given key, value and expiration time. If the max size after the aforementioned operation is above the configured max size, the tail will be evicted. Depending on the eviction policy, the tail is defined as the oldest |
| Get                               | Gets a cache entry by its key.                                                                                                                                                                                                                                     |
| GetByKeys                         | Gets a map of entries by their keys. The resulting map will contain all keys, even if some of the keys in the slice passed as parameter were not present in the cache.                                                                                             |
| GetAll                            | Gets all cache entries.                                                                                                                                                                                                                                            |
| GetKeysByPattern                  | Retrieves a slice of keys that matches a given pattern.                                                                                                                                                                                                            |
| Delete                            | Removes a key from the cache.                                                                                                                                                                                                                                      |
| DeleteAll                         | Removes multiple keys from the cache.                                                                                                                                                                                                                              |
| DeleteKeysByPattern               | Removes all keys that that matches a given pattern.                                                                                                                                                                                                                |
| Count                             | Gets the size of the cache. This includes cache keys which may have already expired, but have not been removed yet.                                                                                                                                                |
| Clear                             | Wipes the cache.                                                                                                                                                                                                                                                   |
| TTL                               | Gets the time until a cache key expires.                                                                                                                                                                                                                           |
| Expire                            | Sets the expiration time of an existing cache key.                                                                                                                                                                                                                 |


### Examples

#### Creating or updating an entry
```go
cache.Set("key", "value") 
cache.Set("key", 1)
cache.Set("key", struct{ Text string }{Test: "value"})
cache.SetWithTTL("key", []byte("value"), 24*time.Hour)
```

#### Getting an entry
```go
value, exists := cache.Get("key")
```
You can also get multiple entries by using `cache.GetByKeys([]string{"key1", "key2"})`

#### Deleting an entry
```go
cache.Delete("key")
```
You can also delete multiple entries by using `cache.DeleteAll([]string{"key1", "key2"})`

#### Complex example
```go
package main

import (
    "fmt"
    "time"

    "github.com/arham09/cache"
)

func main() {
    c := cache.NewCache(cache.WithEvictionPolicy(cache.LeastRecentlyUsed), cache.WithMaxSize(10000))
    cache.StartJanitor() // Passively manages expired entries
    defer cache.StopJanitor()

    c.Set("key", "value")
    c.SetWithTTL("key-with-ttl", "value", 60*time.Minute)
    c.SetAll(map[string]interface{}{"k1": "v1", "k2": "v2", "k3": "v3"})

    fmt.Println("[Count] Cache size:", cache.Count())

    value, exists := c.Get("key")
    fmt.Printf("[Get] key=key; value=%s; exists=%v\n", value, exists)
    for key, value := range c.GetByKeys([]string{"k1", "k2", "k3"}) {
        fmt.Printf("[GetByKeys] key=%s; value=%s\n", key, value)
    }
    for _, key := range c.GetKeysByPattern("key*", 0) {
        fmt.Printf("[GetKeysByPattern] pattern=key*; key=%s\n", key)
    }

    c.Expire("key", time.Hour)
    time.Sleep(500*time.Millisecond)
    timeUntilExpiration, _ := c.TTL("key")
    fmt.Println("[TTL] Number of minutes before 'key' expires:", int(timeUntilExpiration.Seconds()))

    c.Delete("key")
    c.DeleteAll([]string{"k1", "k2", "k3"})
    
    c.Clear()
    fmt.Println("[Count] Cache size after clearing the cache:", c.Count())
}
```

<details>
  <summary>Output</summary>

```
[Count] Cache size: 5
[Get] key=key; value=value; exists=true
[GetByKeys] key=k1; value=v1
[GetByKeys] key=k2; value=v2
[GetByKeys] key=k3; value=v3
[GetKeysByPattern] pattern=key*; key=key-with-ttl
[GetKeysByPattern] pattern=key*; key=key
[TTL] Number of minutes before 'key' expires: 3599
[Count] Cache size after clearing the cache: 0
```
</details>


## Eviction
### MaxSize
Eviction by MaxSize is the default behavior, and is also the most efficient.

The code below will create a cache that has a maximum size of 1000:
```go
c := cache.NewCache(cache.WithMaxSize(1000))
```
This means that whenever an operation causes the total size of the cache to go above 1000, the tail will be evicted.

### MaxMemoryUsage
Eviction by MaxMemoryUsage is **disabled by default**, and is in alpha.

The code below will create a cache that has a maximum memory usage of 50MB:
```go
c := cache.NewCache(cache.WithMaxSize(0), cache.WithMaxMemoryUsage(50*cache.Megabyte))
```
This means that whenever an operation causes the total memory usage of the cache to go above 50MB, one or more tails
will be evicted.

Unlike evictions caused by reaching the MaxSize, evictions triggered by MaxMemoryUsage may lead to multiple entries
being evicted in a row. The reason for this is that if, for instance, you had 100 entries of 0.1MB each and you suddenly added 
a single entry of 10MB, 100 entries would need to be evicted to make enough space for that new big entry.

It's very important to keep in mind that eviction by MaxMemoryUsage is approximate.

**The only memory taken into consideration is the size of the cache, not the size of the entire application.**
If you pass along 100MB worth of data in a matter of seconds, even though the cache's memory usage will remain
under 50MB (or whatever you configure the MaxMemoryUsage to), the memory footprint generated by that 100MB will 
still exist until the next GC cycle.

As previously mentioned, this is a work in progress, and here's a list of the things you should keep in mind:
- The memory usage of structs are a gross estimation and may not reflect the actual memory usage.
- Native types (string, int, bool, []byte, etc.) are the most accurate for calculating the memory usage.
- Adding an entry bigger than the configured MaxMemoryUsage will work, but it will evict all other entries.


## Expiration
There are two ways that the deletion of expired keys can take place:
- Active
- Passive

**Active deletion of expired keys** happens when an attempt is made to access the value of a cache entry that expired. 
`Get`, `GetByKeys` and `GetAll` are the only functions that can trigger active deletion of expired keys.

**Passive deletion of expired keys** runs in the background and is managed by the janitor. 
If you do not start the janitor, there will be no passive deletion of expired keys.


## Performance
### Summary
- **Set**: Both map and cache have the same performance.
- **Get**: Map is faster than cache.

This is because cache keeps track of the head and the tail for eviction and expiration/TTL. 

Ultimately, the difference is negligible. 

We could add a way to disable eviction or disable expiration altogether just to match the map's performance, 
but if you're looking into using a library like cache, odds are, you want more than just a map.


### Results
| key    | value    |
|:------ |:-------- |
| goos   | windows  |
| goarch | amd64    |
| cpu    | i7-9700K |
| mem    | 32G DDR4 |

<!-- ```
// Normal map
BenchmarkMap_Get
BenchmarkMap_Get-8                                                              46087372  26.7 ns/op
BenchmarkMap_Set                                                                   
BenchmarkMap_Set/small_value-8                                                   3841911   389 ns/op
BenchmarkMap_Set/medium_value-8                                                  3887074   391 ns/op
BenchmarkMap_Set/large_value-8                                                   3921956   393 ns/op
// Cache                                                                         
BenchmarkCache_Get                                                                 
BenchmarkCache_Get/FirstInFirstOut-8                                            27273036  46.4 ns/op
BenchmarkCache_Get/LeastRecentlyUsed-8                                          26648248  46.3 ns/op
BenchmarkCache_Set                                                              
BenchmarkCache_Set/FirstInFirstOut_small_value-8                                 2919584   405 ns/op
BenchmarkCache_Set/FirstInFirstOut_medium_value-8                                2990841   391 ns/op
BenchmarkCache_Set/FirstInFirstOut_large_value-8                                 2970513   391 ns/op
BenchmarkCache_Set/LeastRecentlyUsed_small_value-8                               2962939   402 ns/op
BenchmarkCache_Set/LeastRecentlyUsed_medium_value-8                              2962963   390 ns/op
BenchmarkCache_Set/LeastRecentlyUsed_large_value-8                               2962928   394 ns/op
BenchmarkCache_SetUsingMaxMemoryUsage                                           
BenchmarkCache_SetUsingMaxMemoryUsage/small_value-8                              2683356   447 ns/op
BenchmarkCache_SetUsingMaxMemoryUsage/medium_value-8                             2637578   441 ns/op
BenchmarkCache_SetUsingMaxMemoryUsage/large_value-8                              2672434   443 ns/op
BenchmarkCache_SetWithMaxSize                                                   
BenchmarkCache_SetWithMaxSize/100_small_value-8                                  4782966   252 ns/op
BenchmarkCache_SetWithMaxSize/10000_small_value-8                                4067967   296 ns/op
BenchmarkCache_SetWithMaxSize/100000_small_value-8                               3762055   328 ns/op
BenchmarkCache_SetWithMaxSize/100_medium_value-8                                 4760479   252 ns/op
BenchmarkCache_SetWithMaxSize/10000_medium_value-8                               4081050   295 ns/op
BenchmarkCache_SetWithMaxSize/100000_medium_value-8                              3785050   330 ns/op
BenchmarkCache_SetWithMaxSize/100_large_value-8                                  4732909   254 ns/op
BenchmarkCache_SetWithMaxSize/10000_large_value-8                                4079533   297 ns/op
BenchmarkCache_SetWithMaxSize/100000_large_value-8                               3712820   331 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU                                             
BenchmarkCache_SetWithMaxSizeAndLRU/100_small_value-8                            4761732   254 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/10000_small_value-8                          4084474   296 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/100000_small_value-8                         3761402   329 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/100_medium_value-8                           4783075   254 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/10000_medium_value-8                         4103980   296 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/100000_medium_value-8                        3646023   331 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/100_large_value-8                            4779025   254 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/10000_large_value-8                          4096192   296 ns/op
BenchmarkCache_SetWithMaxSizeAndLRU/100000_large_value-8                         3726823   331 ns/op
BenchmarkCache_GetSetMultipleConcurrent                                         
BenchmarkCache_GetSetMultipleConcurrent-8                                         707142  1698 ns/op
BenchmarkCache_GetSetConcurrentWithFrequentEviction
BenchmarkCache_GetSetConcurrentWithFrequentEviction/FirstInFirstOut-8            3616256   334 ns/op
BenchmarkCache_GetSetConcurrentWithFrequentEviction/LeastRecentlyUsed-8          3636367   331 ns/op
BenchmarkCache_GetConcurrentWithLRU                                              
BenchmarkCache_GetConcurrentWithLRU/FirstInFirstOut-8                            4405557   268 ns/op
BenchmarkCache_GetConcurrentWithLRU/LeastRecentlyUsed-8                          4445475   269 ns/op
BenchmarkCache_WithForceNilInterfaceOnNilPointer
BenchmarkCache_WithForceNilInterfaceOnNilPointer/true_with_nil_struct_pointer-8  6184591   191 ns/op
BenchmarkCache_WithForceNilInterfaceOnNilPointer/true-8                          6090482   191 ns/op
BenchmarkCache_WithForceNilInterfaceOnNilPointer/false_with_nil_struct_pointer-8 6184629   187 ns/op
BenchmarkCache_WithForceNilInterfaceOnNilPointer/false-8                         6281781   186 ns/op
(Trimmed "BenchmarkCache_" for readability)
WithForceNilInterfaceOnNilPointerWithConcurrency
WithForceNilInterfaceOnNilPointerWithConcurrency/true_with_nil_struct_pointer-8  4379564   268 ns/op
WithForceNilInterfaceOnNilPointerWithConcurrency/true-8                          4379558   265 ns/op
WithForceNilInterfaceOnNilPointerWithConcurrency/false_with_nil_struct_pointer-8 4444456   261 ns/op
WithForceNilInterfaceOnNilPointerWithConcurrency/false-8                         4493896   262 ns/op
``` -->