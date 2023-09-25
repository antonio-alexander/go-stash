package redis

import (
	"context"
	"encoding"
	"encoding/json"
	"sort"
	"sync"
	"time"

	stash "github.com/antonio-alexander/go-stash"
	internal "github.com/antonio-alexander/go-stash/internal"

	errors "github.com/pkg/errors"
	redis "github.com/redis/go-redis/v9"
)

type stashRedis struct {
	sync.RWMutex
	sync.WaitGroup
	*redis.Client
	internal.Logger
	stopper     chan struct{}
	config      *Configuration
	initialized bool
	configured  bool
}

func New(parameters ...interface{}) interface {
	stash.Stasher
	Redis
} {
	var configured bool

	s := &stashRedis{}
	s.SetParameters(parameters...)
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *Configuration:
			if err := s.Configure(p); err != nil {
				panic(err)
			}
			configured = true
		}
	}
	if configured {
		if err := s.Initialize(); err != nil {
			panic(err)
		}
	}
	return s
}

func (s *stashRedis) printf(format string, a ...interface{}) {
	if s.Logger != nil && s.config != nil && s.config.Debug {
		s.Logger.Printf(s.config.DebugPrefix+format, a...)
	}
}

func (s *stashRedis) evict() {
	var cachedItems []*stash.CachedItem

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	items, err := s.HGetAll(ctx, s.config.HashKey).Result()
	if err != nil {
		s.printf("error while evicting: %s\n", err.Error())
		return
	}
	for _, item := range items {
		var cachedItem stash.CachedItem

		if err := json.Unmarshal([]byte(item), &cachedItem); err != nil {
			s.printf("error while evicting: %s\n", err.Error())
			return
		}
		cachedItems = append(cachedItems, &cachedItem)
	}
	evictionPolicy := s.config.EvictionPolicy
	if evictionPolicy != "" {
		s.printf("eviction Policy: %s\n", evictionPolicy)
	}
	switch evictionPolicy {
	case stash.FirstInFirstOut:
		sort.Sort(stash.ByFirstCreated(cachedItems))
		for _, cacheItem := range cachedItems {
			s.printf("key: %v, %v\n", cacheItem.Key, cacheItem.LastRead)
		}
	case stash.LeastRecentlyUsed:
		sort.Sort(stash.ByLastRead(cachedItems))
		for _, cacheItem := range cachedItems {
			s.printf("key: %v, %v\n", cacheItem.Key, cacheItem.LastRead)
		}
	case stash.LeastFrequentlyUsed:
		sort.Sort(stash.ByTimesRead(cachedItems))
		for _, cacheItem := range cachedItems {
			s.printf("key: %v, %d\n", cacheItem.Key, cacheItem.NTimesRead)
		}
	}
	//ensure we don't evict the only data that's in the
	// stash even if we're above the max limit because
	// there's only a single item in the stash
	if len(cachedItems) <= 1 {
		return
	}
	tNow := time.Now()
	for _, cacheItem := range cachedItems {
		switch {
		default:
			if s.config.TimeToLive == 0 {
				return
			}
		case s.config.TimeToLive > 0 && tNow.Sub(time.Unix(0, cacheItem.LastUpdated)) > s.config.TimeToLive:
			ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
			defer cancel()
			field, err := parseKey(cacheItem.Key)
			if err != nil {
				s.printf("error while evicting: %s\n", err.Error())
				return
			}
			if err := s.HDel(ctx, s.config.HashKey, field).Err(); err != nil {
				s.printf("error while evicting: %s\n", err.Error())
				return
			}
			s.printf("evicted key: %v, ttl exceeded\n", cacheItem.Key)
		}
	}
}

func (s *stashRedis) launchEvict() {
	if s.config.EvictionRate <= 0 {
		s.printf("eviction go routine disabled\n")
		return
	}
	started := make(chan struct{})
	s.Add(1)
	go func() {
		defer s.Done()

		tEvict := time.NewTicker(s.config.EvictionRate)
		defer tEvict.Stop()
		close(started)
		s.evict()
		for {
			select {
			case <-s.stopper:
				return
			case <-tEvict.C:
				s.evict()
			}
		}
	}()
	<-started
}

func (s *stashRedis) write(key, item interface{}) error {
	var bytes []byte

	field, err := parseKey(key)
	if err != nil {
		return err
	}
	switch v := item.(type) {
	default:
		//return error
		return errors.Errorf("unsupported item to write: %T", v)
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	case encoding.BinaryMarshaler:
		byts, err := v.MarshalBinary()
		if err != nil {
			return err
		}
		bytes = byts
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	return s.HSet(ctx, s.config.HashKey, field, string(bytes)).Err()
}

func (s *stashRedis) read(key interface{}) (*stash.CachedItem, error) {
	var cachedItem stash.CachedItem

	field, err := parseKey(key)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	value, err := s.HGet(ctx, s.config.HashKey, field).Result()
	if err != nil {
		return nil, err
	}
	if err := cachedItem.UnmarshalBinary([]byte(value)); err != nil {
		return nil, err
	}
	return &cachedItem, nil
}

func (s *stashRedis) Configure(items ...interface{}) error {
	s.Lock()
	defer s.Unlock()

	var config *Configuration

	for _, item := range items {
		switch item := item.(type) {
		case *Configuration:
			config = item
		case Configuration:
			config = &item
		}
	}
	if config != nil {
		s.config = config
		s.configured = true
	}
	return nil
}

func (s *stashRedis) SetParameters(items ...interface{}) {
	s.Lock()
	defer s.Unlock()

	for _, item := range items {
		switch item := item.(type) {
		case internal.Logger:
			s.Logger = item
		}
	}
}

func (s *stashRedis) Initialize() error {
	s.Lock()
	defer s.Unlock()

	if !s.configured {
		return errors.New("not configured")
	}
	if s.initialized {
		return errors.New("already initialized")
	}
	s.Client = redis.NewClient(s.config.ToRedisOptions())
	s.stopper = make(chan struct{})
	s.launchEvict()
	s.initialized = true
	return nil
}

func (s *stashRedis) Shutdown() error {
	s.Lock()
	defer s.Unlock()

	if !s.initialized {
		return nil
	}
	close(s.stopper)
	s.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	if err := s.Client.Shutdown(ctx).Err(); err != nil {
		return err
	}
	s.initialized, s.configured = false, false
	return nil
}

func (s *stashRedis) Write(key interface{}, itemToCache stash.Cacheable) (bool, error) {
	s.RLock()
	defer s.evict()
	defer s.RUnlock()

	cachedItem, err := s.read(key)
	if err != nil && err != redis.Nil {
		return false, err
	}
	switch {
	default: //found
		if err := stash.UpdateCacheItem(cachedItem, itemToCache); err != nil {
			return false, err
		}
		if err := s.write(key, cachedItem); err != nil {
			return false, err
		}
		s.printf("updated key: %v\n", key)
		return true, nil
	case err == redis.Nil: //not found
		cachedItem, err := stash.CreateCacheItem(key, itemToCache)
		if err != nil {
			return false, err
		}
		if err := s.write(key, cachedItem); err != nil {
			return false, err
		}
		s.printf("created key: %v\n", key)
		return false, nil
	}
}

func (s *stashRedis) Read(key interface{}, v stash.Cacheable) error {
	s.RLock()
	defer s.evict()
	defer s.RUnlock()

	cachedItem, err := s.read(key)
	if err != nil {
		switch err {
		default:
			return err
		case redis.Nil:
			return errors.Errorf("value for %s not found", key)
		}
	}
	cachedItem.LastRead = time.Now().UnixNano()
	cachedItem.NTimesRead++
	if err := v.UnmarshalBinary(cachedItem.Bytes); err != nil {
		return err
	}
	if err := s.write(key, cachedItem); err != nil {
		return err
	}
	s.printf("read key: %v\n", key)
	return nil
}

func (s *stashRedis) Delete(key interface{}) error {
	s.RLock()
	defer s.evict()
	defer s.RUnlock()

	field, err := parseKey(key)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	result, err := s.HDel(ctx, s.config.HashKey, field).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return errors.Errorf("value for %s not found", key)
	}
	s.printf("deleted key: %v\n", key)
	return nil
}
