package redis_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/antonio-alexander/go-stash"
	"github.com/antonio-alexander/go-stash/internal"
	"github.com/antonio-alexander/go-stash/redis"
	"github.com/antonio-alexander/go-stash/tests"

	"github.com/stretchr/testify/assert"
)

var (
	configuration = redis.NewConfiguration()
	envs          = make(map[string]string)
)

func init() {
	for _, env := range os.Environ() {
		if s := strings.Split(env, "="); len(s) > 1 {
			envs[s[0]] = strings.Join(s[1:], "=")
		}
	}
	configuration.Debug = true
}

func TestStashRedis(t *testing.T) {
	newStash := func(config *redis.Configuration) stash.Stasher {
		logger := internal.NewLogger()
		r := redis.New()
		r.SetParameters(logger)
		err := r.Configure(config)
		assert.Nil(t, err)
		err = r.Initialize()
		assert.Nil(t, err)
		return r
	}
	t.Run("Stash", tests.TestStash(t, func() stash.Stasher {
		return newStash(configuration)
	}))
	t.Run("Evict Least Recently Used", tests.TestEvictLeastRecentlyUsed(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			config := redis.NewConfiguration()
			config.EvictionPolicy = stash.LeastRecentlyUsed
			config.TimeToLive = timeToLive
			config.DebugPrefix = "[stash] "
			return newStash(config)
		}))
	t.Run("Evict Least Frequently Used", tests.TestEvictLeastFrequentlyUsed(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			config := redis.NewConfiguration()
			config.EvictionPolicy = stash.LeastFrequentlyUsed
			config.TimeToLive = timeToLive
			config.DebugPrefix = "[stash] "
			return newStash(config)
		}))
	// t.Run("Evict First In First Out", tests.TestEvictFirstInFirstOut(t,
	// 	func(timeToLive time.Duration, maxSize int) interface {
	// 		stash.Stasher
	// 	} {
	// 		config := redis.NewConfiguration()
	// 		config.EvictionPolicy = stash.FirstInFirstOut
	// 		config.TimeToLive = timeToLive
	// 		config.DebugPrefix = "[stash] "
	// 		return newStash(config)
	// 	}))
}
