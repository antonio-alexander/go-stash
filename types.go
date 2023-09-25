package stash

import (
	"encoding"
	"encoding/json"
)

// EvictionPolicy is a typed string used to describe the configured eviction
// policy for a given Stasher
type EvictionPolicy string

const (
	LeastRecentlyUsed   EvictionPolicy = "least_recently_used"
	LeastFrequentlyUsed EvictionPolicy = "least_frequently_used"
	FirstInFirstOut     EvictionPolicy = "first_in_first_out"
)

// Stasher is an interface used to read and write data to a cache/stash
// KIM: Although key is an interface, if that interface doesn't contain
// something that is serializable, other concrete implementations won't
// work.
type Stasher interface {
	//Write can be used to create/update a value in the cache with the given
	// key. If the value exists, replaced will be true
	Write(key interface{}, value Cacheable) (replaced bool, err error)

	//Read can be used to read a value in the cache with the given key
	// if the value exists, it will be unmarshalled into the Cacheable
	// pointer; this is expected to work very much like an Unmarshal
	// function. If a value isn't found with the given key, an error
	// will be returned
	Read(key interface{}, v Cacheable) (err error)

	//Delete can be used to remove a value from the cache with a given
	// key. If the value isn't found, an error is returned.
	//KIM: this function doesn't return the value by design; why would
	// you need to read the value if you're deleting it?
	Delete(key interface{}) (err error)
}

// Cacheable is an interface used to describe values (and keys) that can
// be stored within a cache/stash; this generally means that any value
// provided to the cache/stash MUST be serializable
type Cacheable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type CachedItem struct {
	Key          interface{} `json:"key"`
	Bytes        []byte      `json:"bytes"`
	FirstCreated int64       `json:"first_created,string"`
	LastUpdated  int64       `json:"last_updated,string"`
	LastRead     int64       `json:"last_read,string"`
	NTimesRead   int         `json:"n_times_read"`
	Size         int         `json:"size"`
}

func (c *CachedItem) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CachedItem) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, c)
}
