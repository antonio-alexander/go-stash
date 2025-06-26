package redis

import (
	"strconv"
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

func (c *Configuration) Default() {
	if c == nil {
		return
	}
	c.Address = defaultAddress
	c.Port = defaultPort
	c.Database = defaultDatabase
	c.HashKey = defaultHashKey
	c.Timeout = defaultTimeout
	c.EvictionRate = defaultEvictionRate
}

func (c *Configuration) FromEnvs(envs map[string]string) {
	for key, value := range envs {
		if value == "" {
			continue
		}
		switch key {
		case "REDIS_ADDRESS":
			c.Address = value
		case "REDIS_PORT":
			c.Port = value
		case "REDIS_PASSWORD":
			c.Password = value
		case "REDIS_DATABASE":
			c.Database, _ = strconv.Atoi(value)
		case "REDIS_HASH_KEY":
			c.HashKey = value
		case "REDIS_TIMEOUT":
			t, _ := strconv.Atoi(value)
			c.EvictionRate = time.Second * time.Duration(t)
		case "STASH_EVICTION_RATE":
			t, _ := strconv.Atoi(value)
			c.EvictionRate = time.Second * time.Duration(t)
		case "STASH_EVICTION_POLICY":
			c.EvictionPolicy = stash.EvictionPolicy(value)
		case "STASH_TIME_TO_LIVE":
			t, _ := strconv.Atoi(value)
			c.TimeToLive = time.Second * time.Duration(t)
		case "STASH_DEBUG_ENABLED":
			c.Debug, _ = strconv.ParseBool(value)
		case "STASH_DEBUG_PREFIX":
			c.DebugPrefix = value
		}
	}
}
