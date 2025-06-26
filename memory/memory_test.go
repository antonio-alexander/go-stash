package memory_test

import (
	"testing"
	"time"

	"github.com/antonio-alexander/go-stash"
	"github.com/antonio-alexander/go-stash/internal"
	"github.com/antonio-alexander/go-stash/memory"
	"github.com/antonio-alexander/go-stash/tests"

	"github.com/stretchr/testify/assert"
)

func TestStashMemory(t *testing.T) {
	const debug = true

	newStash := func(config memory.Configuration) stash.Stasher {
		logger := internal.NewLogger()
		m := memory.New()
		m.SetParameters(logger)
		err := m.Configure(config)
		assert.Nil(t, err)
		return m
	}
	t.Run("Stash", tests.TestStash(t, func() stash.Stasher {
		return newStash(memory.Configuration{})
	}))
	t.Run("Evict Size", tests.TestEvictSize(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				TimeToLive:  timeToLive,
				MaxSize:     maxSize,
				Debug:       debug,
				DebugPrefix: "[stash] ",
			})
		}))
	// t.Run("Evict Least Recently Used", tests.TestEvictLeastRecentlyUsed(t,
	// 	func(timeToLive time.Duration, maxSize int) interface {
	// 		stash.Stasher
	// 	} {
	// 		return newStash(memory.Configuration{
	// 			EvictionPolicy: stash.LeastRecentlyUsed,
	// 			TimeToLive:     timeToLive,
	// 			MaxSize:        maxSize,
	// 			Debug:          debug,
	// 			DebugPrefix:    "[stash] ",
	// 		})
	// 	}))
	// t.Run("Evict Least Frequently Used", tests.TestEvictLeastFrequentlyUsed(t,
	// 	func(timeToLive time.Duration, maxSize int) interface {
	// 		stash.Stasher
	// 	} {
	// 		return newStash(memory.Configuration{
	// 			EvictionPolicy: stash.LeastFrequentlyUsed,
	// 			TimeToLive:     timeToLive,
	// 			MaxSize:        maxSize,
	// 			Debug:          debug,
	// 			DebugPrefix:    "[stash] ",
	// 		})
	// 	}))
	t.Run("Evict First In First Out", tests.TestEvictFirstInFirstOut(t,
		func(timeToLive time.Duration, maxSize int) interface {
			stash.Stasher
		} {
			return newStash(memory.Configuration{
				EvictionPolicy: stash.FirstInFirstOut,
				TimeToLive:     timeToLive,
				MaxSize:        maxSize,
				Debug:          debug,
				DebugPrefix:    "[stash] ",
			})
		}))
}
