package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/antonio-alexander/go-stash"

	"github.com/pkg/errors"
)

type stashMemory struct {
	sync.Mutex
	data        map[interface{}]*cacheItem
	config      Configuration
	size        int
	initialized bool
}

//New can be used to create a concrete instance of a memory
// cache/stash. If Configuration is provided, it will attempt
// to initialize the pointer (it will panic if this initialization
// fails)
func New(parameters ...interface{}) interface {
	stash.Stasher
	Memory
} {
	var config Configuration
	var configSet bool

	s := &stashMemory{}
	for _, parameter := range parameters {
		switch p := parameter.(type) {
		case *Configuration:
			config = *p
			configSet = true
		case Configuration:
			config = p
			configSet = true
		}
	}
	if configSet {
		if err := s.Initialize(config); err != nil {
			panic(err)
		}
	}
	return s
}

func (s *stashMemory) printf(format string, a ...interface{}) {
	if s.config.Debug {
		fmt.Printf(format, a...)
	}
}

func (s *stashMemory) evict() {
	evictionPolicy := s.config.EvictionPolicy
	cacheItems := toSlice(s.data)
	if evictionPolicy != "" {
		s.printf("eviction Policy: %s\n", evictionPolicy)
	}
	switch evictionPolicy {
	case stash.FirstInFirstOut:
		sort.Sort(byFirstCreated(cacheItems))
		for _, cacheItem := range cacheItems {
			s.printf("key: %v, %v\n", cacheItem.key, cacheItem.lastRead)
		}
	case stash.LeastRecentlyUsed:
		sort.Sort(byLastRead(cacheItems))
		for _, cacheItem := range cacheItems {
			s.printf("key: %v, %v\n", cacheItem.key, cacheItem.lastRead)
		}
	case stash.LeastFrequentlyUsed:
		sort.Sort(byTimesRead(cacheItems))
		for _, cacheItem := range cacheItems {
			s.printf("key: %v, %d\n", cacheItem.key, cacheItem.nTimesRead)
		}
	}
	tNow := time.Now()
	for _, cacheItem := range cacheItems {
		if s.config.MaxSize > 0 {
			s.printf("size: %d/%d\n", s.size, s.config.MaxSize)
		}
		switch {
		default:
			if s.config.TimeToLive == 0 {
				return
			}
		case s.config.MaxSize > 0 && s.size > s.config.MaxSize:
			s.size = s.size - cacheItem.size
			delete(s.data, cacheItem.key)
			s.printf("evicted key: %v, max size exceeded\n", cacheItem.key)
		case s.config.TimeToLive > 0 && tNow.Sub(cacheItem.lastUpdated) > s.config.TimeToLive:
			s.size = s.size - cacheItem.size
			delete(s.data, cacheItem.key)
			s.printf("evicted key: %v, ttl exceeded\n", cacheItem.key)
		}
	}
}

//Initialize can be used to setup internal pointers
// and ready the stash for usage
func (s *stashMemory) Initialize(config Configuration) error {
	s.Lock()
	defer s.Unlock()
	if s.initialized {
		return errors.New("already initialized")
	}
	s.size = 0
	s.config = config
	s.data = make(map[interface{}]*cacheItem)
	s.initialized = true
	return nil
}

//Shutdown can be used to tear down internal pointers
// and ready the stash for garbage collection (or reuse)
func (s *stashMemory) Shutdown() error {
	s.Lock()
	defer s.Unlock()
	if !s.initialized {
		return nil
	}
	for _, key := range s.data {
		delete(s.data, key)
	}
	s.size = 0
	return nil
}

//Write can be used to create/update a value in the cache with the given
// key. If the value exists, replaced will be true
func (s *stashMemory) Write(key interface{}, item stash.Cacheable) (bool, error) {
	s.Lock()
	defer s.evict()
	defer s.Unlock()
	cacheItem, found := s.data[key]
	if found {
		s.size -= cacheItem.size
		if err := updateCacheItem(cacheItem, item); err != nil {
			return false, err
		}
		s.size += cacheItem.size
		s.printf("updated key: %v\n", key)
		return found, nil
	}
	cacheItem, err := createCacheItem(key, item)
	if err != nil {
		return false, err
	}
	s.data[key] = cacheItem
	s.size += cacheItem.size
	s.printf("created key: %v\n", key)
	return found, nil
}

//Read can be used to read a value in the cache with the given key
// if the value exists, it will be unmarshalled into the Cacheable
// pointer; this is expected to work very much like an Unmarshal
// function. If a value isn't found with the given key, an error
// will be returned
func (s *stashMemory) Read(key interface{}, v stash.Cacheable) error {
	s.Lock()
	defer s.evict()
	defer s.Unlock()
	item, found := s.data[key]
	if !found {
		return errors.Errorf("value for %s not found", key)
	}
	item.lastRead = time.Now()
	item.nTimesRead++
	bytes := make([]byte, len(item.bytes))
	copy(bytes, item.bytes)
	err := v.UnmarshalBinary(item.bytes)
	if err != nil {
		return err
	}
	s.printf("read key: %v\n", key)
	return nil
}

//Delete can be used to remove a value from the cache with a given
// key. If the value isn't found, an error is returned.
func (s *stashMemory) Delete(key interface{}) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.data[key]; !ok {
		return errors.Errorf("value not found for key: %v", key)
	}
	delete(s.data, key)
	s.printf("deleted key: %v\n", key)
	return nil
}
