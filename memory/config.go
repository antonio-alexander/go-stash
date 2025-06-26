package memory

import (
	"strconv"
	"time"

	"github.com/antonio-alexander/go-stash"
)

const (
	defaultEvictionPolicy stash.EvictionPolicy = stash.LeastFrequentlyUsed
	defaultTimeToLive     time.Duration        = 30 * time.Second
	defaultMaxSize        int                  = 10
	defaultDebugEnabled   bool                 = true
)

// Configuration describes what can be configured for the
// memory stash
type Configuration struct {
	EvictionPolicy stash.EvictionPolicy `json:"eviction_policy"`
	TimeToLive     time.Duration        `json:"time_to_live"`
	MaxSize        int                  `json:"max_size"`
	Debug          bool                 `json:"debug"`
	DebugPrefix    string               `json:"debug_prefix"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		EvictionPolicy: defaultEvictionPolicy,
		TimeToLive:     defaultTimeToLive,
		MaxSize:        defaultMaxSize,
		Debug:          defaultDebugEnabled,
	}
}

func (c *Configuration) FromEnvs(envs map[string]string) {
	for key, value := range envs {
		if value == "" {
			continue
		}
		switch key {
		case "STASH_EVICTION_POLICY":
			c.EvictionPolicy = stash.EvictionPolicy(value)
		case "STASH_TIME_TO_LIVE":
			t, _ := strconv.Atoi(value)
			c.TimeToLive = time.Second * time.Duration(t)
		case "STASH_MAX_SIZE":
			c.MaxSize, _ = strconv.Atoi(value)
		case "STASH_DEBUG_ENABLED":
			c.Debug, _ = strconv.ParseBool(value)
		case "STASH_DEBUG_PREFIX":
			c.DebugPrefix = value
		}
	}
}

func (c *Configuration) Default() {
	if c == nil {
		return
	}

	c.EvictionPolicy = defaultEvictionPolicy
	c.TimeToLive = defaultTimeToLive
	c.MaxSize = defaultMaxSize
	c.Debug = defaultDebugEnabled
}
