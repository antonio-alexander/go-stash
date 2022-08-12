# github.com/antonio-alexander/go-stash

go-stash is a proof of concept/implementation of a caching library that can be used to store, retrieve and delete a pointer that is cache-able. It's expected use case is to be placed between the API making the call to retrieve data and the actual process to read that data.

Caches attempt to reduce the time to read objects that have already been read before with the assumption that reading something that's already in-memory is almost always faster than making the actual call to read. This "time savings" can be extended to centralized caching (e.g., Redis) or extended to more complex calls where you can perform a repeat-able search and return the same items. Another interesting use case for caches is when you need to query a lot of data, but only need to display a fixed amount of data at a time (e.g., a page); if you cache all the data, once you've loaded the first page, it can appear to the user that other pages load significantly faster/instant because you can cache the results.

## Eviction

The hard part about caching is knowing when you can trust the data: am I serving valid data? This problem is specifically solved using eviction rules:

- Least Recently Used: the cache will record when data is used and will periodically evict data that hasn't been used recently
- Least Frequently Used: the cache will record how often data is used and will periodically evict data that's not used often
- First In First Out: the cache will remember the order in which data is placed in the cache, and when the cache is "full", it will evict the data that was placed first.

```go
//EvictionPolicy is a typed string used to describe the configured eviction
// policy for a given Stasher
type EvictionPolicy string

const (
 LeastRecentlyUsed   EvictionPolicy = "least_recently_used"
 LeastFrequentlyUsed EvictionPolicy = "least_frequently_used"
 FirstInFirstOut     EvictionPolicy = "first_in_first_out"
)
```

## Creating your own concrete implementation

## Memory

Within the memory folder, a concrete implementation of the stasher (with eviction logic) is provided for research purposes and for situations where there's some efficiency benefit to having a memory layer/stash between a "slower" concrete implementation of your own making.

### Getting Started

go-stash has a defined "implementation" of a cache with matching tests to verify behavior as well as a concrete implementation of a memory cache. This is a basic example of how to use the Stasher interface as well as how to read and write data using the concrete memory implementation.

```go
package main

import (
    "fmt"
    "math/rand"
    "reflect"
    "time"

    "github.com/antonio-alexander/go-stash"
    "github.com/antonio-alexander/go-stash/memory"

    "github.com/google/uuid"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func main() {
    //create stash pointer/interface
    s := memory.New()

    //initialize the stash
    if err := s.Initialize(memory.Configuration{
        EvictionPolicy: stash.FirstInFirstOut,
        TimeToLive:     time.Minute,
        MaxSize:        -1,
        Debug:          true,
    }); err != nil {
        fmt.Printf("error while initializing: %s\n", err)
    }

    //defer shutdown (so it happens even if there's a panic)
    defer func() {
        if err := s.Shutdown(); err != nil {
            fmt.Printf("error while shutting down: %s\n", err)
        }
    }()

    //create example data
    key := uuid.Must(uuid.NewRandom()).String()
    example := &stash.Example{Int: rand.Int()}

    //write data
    if _, err := s.Write(key, example); err != nil {
        fmt.Printf("error while writing: %s\n", err)
    }

    //read data
    exampleRead := &stash.Example{}
    if err := s.Read(key, exampleRead); err != nil {
        fmt.Printf("error while reading: %s\n", err)
    }
    if !reflect.DeepEqual(example, exampleRead) {
        fmt.Println("read value isn't equal to write value")
    }

    //delete data
    if err := s.Delete(key); err != nil {
        fmt.Printf("error while deleting: %s\n", err)
    }
}
```

### Configuration

The memory stash can be lightly configured to control how eviction happens as well as whether or not to enable debugging:

- Eviction Policy: This determines which logic to use when evicting
- Time To Live: This determines the general lifetime of any data within the stash
- Max Size: This provides the maximum size of the stash (this is generally what signals eviction)

```go
//Configuration describes what can be configured for the
// memory stash
type Configuration struct {
 EvictionPolicy stash.EvictionPolicy `json:"eviction_policy"`
 TimeToLive     time.Duration        `json:"time_to_live"`
 MaxSize        int                  `json:"max_size"`
 Debug          bool                 `json:"debug"`
}
```
