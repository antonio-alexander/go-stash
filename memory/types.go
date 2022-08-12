package memory

import (
	"time"

	"github.com/antonio-alexander/go-stash"
)

type cacheItem struct {
	key          interface{}
	bytes        []byte
	firstCreated time.Time
	lastUpdated  time.Time
	lastRead     time.Time
	nTimesRead   int
	size         int
}

//Memory describes the concrete implementation of the memory stash
// that isn't described by the Stasher interface
type Memory interface {
	//Initialize can be used to setup internal pointers
	// and ready the stash for usage
	Initialize(config Configuration) error

	//Shutdown can be used to tear down internal pointers
	// and ready the stash for garbage collection (or reuse)
	Shutdown() error
}

//Configuration describes what can be configured for the
// memory stash
type Configuration struct {
	EvictionPolicy stash.EvictionPolicy `json:"eviction_policy"`
	TimeToLive     time.Duration        `json:"time_to_live"`
	MaxSize        int                  `json:"max_size"`
	Debug          bool                 `json:"debug"`
}
