package memory_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/antonio-alexander/go-stash"
	"github.com/antonio-alexander/go-stash/memory"
	"github.com/antonio-alexander/go-stash/tests"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestStashMemory(t *testing.T) {
	const debug = true

	newStash := func(config memory.Configuration) stash.Stasher {
		return memory.New(config)
	}
	t.Run("Stash", tests.TestStash(t, func() stash.Stasher {
		return newStash(memory.Configuration{})
	}))
	t.Run("Evict Size", tests.TestEvictSize(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				TimeToLive: timeToLive,
				MaxSize:    maxSize,
				Debug:      debug,
			})
		}))
	t.Run("Evict Least Recently Used", tests.TestEvictLeastRecentlyUsed(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				EvictionPolicy: stash.LeastRecentlyUsed,
				TimeToLive:     timeToLive,
				MaxSize:        maxSize,
				Debug:          debug,
			})
		}))
	t.Run("Evict Least Frequently Used", tests.TestEvictLeastFrequentlyUsed(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				EvictionPolicy: stash.LeastFrequentlyUsed,
				TimeToLive:     timeToLive,
				MaxSize:        maxSize,
				Debug:          debug,
			})
		}))
	t.Run("Evict First In First Out", tests.TestEvictFirstInFirstOut(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				EvictionPolicy: stash.FirstInFirstOut,
				TimeToLive:     timeToLive,
				MaxSize:        maxSize,
				Debug:          debug,
			})
		}))
}
