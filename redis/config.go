package redis

import (
	"time"

	stash "github.com/antonio-alexander/go-stash"
	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultAddress      string        = "localhost"
	defaultPort         string        = "6379"
	defaultDatabase     int           = 0
	defaultHashKey      string        = "gostash_redis"
	defaultTimeout      time.Duration = 10 * time.Second
	defaultEvictionRate time.Duration = time.Minute
)

type Configuration struct {
	Address        string               `json:"address"`
	Port           string               `json:"port"`
	Password       string               `json:"password"`
	Database       int                  `json:"database"`
	HashKey        string               `json:"hash_key"`
	Timeout        time.Duration        `json:"timeout"`
	EvictionPolicy stash.EvictionPolicy `json:"eviction_policy"`
	TimeToLive     time.Duration        `json:"time_to_live"`
	Debug          bool                 `json:"debug"`
	DebugPrefix    string               `json:"debug_prefix"`
	EvictionRate   time.Duration        `json:"eviction_rate"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Address:      defaultAddress,
		Port:         defaultPort,
		Database:     defaultDatabase,
		HashKey:      defaultHashKey,
		Timeout:      defaultTimeout,
		EvictionRate: defaultEvictionRate,
	}
}

func (c *Configuration) ToRedisOptions() *goredis.Options {
	address := c.Address
	if c.Port != "" {
		address = address + ":" + c.Port
	}
	return &goredis.Options{
		Addr:     address,
		Password: c.Password,
		DB:       c.Database,
	}
}
