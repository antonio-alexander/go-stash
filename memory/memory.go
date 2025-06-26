package memory

import (
	"sort"
	"sync"
	"time"

	"github.com/antonio-alexander/go-stash"

	"github.com/pkg/errors"
)

type stashMemory struct {
	sync.Mutex
	logger      stash.Logger
	data        map[any]*stash.CachedItem
	config      *Configuration
	size        int
	initialized bool
	configured  bool
}

// New can be used to create a concrete instance of a memory
// cache/stash. If Configuration is provided, it will attempt
// to initialize the pointer (it will panic if this initialization
// fails)
func New(parameters ...any) interface {
	stash.Stasher
	stash.Configurer
	stash.Initializer
	stash.Shutdowner
	stash.Parameterizer
} {
	s := &stashMemory{
		data: make(map[any]*stash.CachedItem),
	}
	s.SetParameters(parameters...)
	return s
}

func (s *stashMemory) printf(format string, a ...any) {
	if s.logger != nil && s.config != nil && s.config.Debug {
		s.logger.Printf(s.config.DebugPrefix+format, a...)
	}
}

func (s *stashMemory) evict() {
	evictionPolicy := s.config.EvictionPolicy
	cacheItems := toSlice(s.data)
	switch evictionPolicy {
	case stash.FirstInFirstOut:
		sort.Sort(stash.ByFirstCreated(cacheItems))
	case stash.LeastRecentlyUsed:
		sort.Sort(stash.ByLastRead(cacheItems))
	case stash.LeastFrequentlyUsed:
		sort.Sort(stash.ByTimesRead(cacheItems))
	}
	//ensure we don't evict the only data that's in the
	// stash even if we're above the max limit because
	// there's only a single item in the stash

	//ensure that we don't evict items unnecessarily, we shuoldn't if:
	// - we haven't exceeded the max size and it's greater than 0
	// - the time to live is greater than 0
	if (s.config.MaxSize > 0 && (s.config.MaxSize-s.size) > 0) ||
		s.config.TimeToLive > 0 {
		return
	}
	tNow := time.Now()
	for _, cacheItem := range cacheItems {
		switch {
		case s.config.MaxSize > 0 && s.size > s.config.MaxSize:
			s.size = s.size - cacheItem.Size
			delete(s.data, cacheItem.Key)
			s.printf("evicted key: %v, max size exceeded\n", cacheItem.Key)
		case s.config.TimeToLive > 0 && tNow.Sub(time.Unix(0, cacheItem.LastUpdated)) > s.config.TimeToLive:
			s.size = s.size - cacheItem.Size
			delete(s.data, cacheItem.Key)
			s.printf("evicted key: %v, ttl exceeded\n", cacheItem.Key)
		}
	}
}

// Configure
func (s *stashMemory) Configure(items ...any) error {
	s.Lock()
	defer s.Unlock()

	var config *Configuration

	for _, item := range items {
		switch item := item.(type) {
		case *Configuration:
			config = item
		case Configuration:
			config = &item
		case map[string]string:
			config = &Configuration{}
			config.Default()
			config.FromEnvs(item)
		}
	}
	if config != nil {
		s.config = config
		s.configured = true
	}

	return nil
}

// SetParameters
func (s *stashMemory) SetParameters(items ...any) {
	for _, item := range items {
		switch item := item.(type) {
		case stash.Logger:
			s.logger = item
		}
	}
}

// Initialize can be used to setup internal pointers
// and ready the stash for usage
func (s *stashMemory) Initialize() error {
	s.Lock()
	defer s.Unlock()

	if !s.configured {
		return errors.New("not configured")
	}
	if s.initialized {
		return errors.New("already initialized")
	}
	if s.config.MaxSize > 0 {
		maxSize := float64(s.config.MaxSize) / 1024 / 1024
		s.printf("configured max size: %dMB\n", s.size, maxSize)
	}
	if s.config.TimeToLive > 0 {
		s.printf("configured time to live: %#v", s.config.TimeToLive)
	}
	s.printf("configured eviction policy: %s", s.config.EvictionPolicy)
	s.size = 0
	s.initialized = true

	return nil
}

// Shutdown can be used to tear down internal pointers
// and ready the stash for garbage collection (or reuse)
func (s *stashMemory) Shutdown() error {
	s.Lock()
	defer s.Unlock()

	if !s.initialized {
		return nil
	}
	s.size = 0
	s.data = make(map[any]*stash.CachedItem)
	s.initialized, s.configured = false, false

	return nil
}

// Write can be used to create/update a value in the cache with the given
// key. If the value exists, replaced will be true
func (s *stashMemory) Write(key any, item stash.Cacheable) (bool, error) {
	s.Lock()
	defer s.evict()
	defer s.Unlock()

	cacheItem, found := s.data[key]
	if found {
		s.size -= cacheItem.Size
		if err := stash.UpdateCacheItem(cacheItem, item); err != nil {
			return false, err
		}
		s.size += cacheItem.Size
		s.printf("updated key: %v\n", key)
		return found, nil
	}
	cacheItem, err := stash.CreateCacheItem(key, item)
	if err != nil {
		return false, err
	}
	s.data[key] = cacheItem
	s.size += cacheItem.Size
	s.printf("created key: %v\n", key)

	return found, nil
}

// Read can be used to read a value in the cache with the given key
// if the value exists, it will be unmarshalled into the Cacheable
// pointer; this is expected to work very much like an Unmarshal
// function. If a value isn't found with the given key, an error
// will be returned
func (s *stashMemory) Read(key any, v stash.Cacheable) error {
	s.Lock()
	defer s.evict()
	defer s.Unlock()

	item, found := s.data[key]
	if !found {
		return errors.Errorf("value for %v not found", key)
	}
	item.LastRead = time.Now().UnixNano()
	item.NTimesRead++
	bytes := make([]byte, len(item.Bytes))
	copy(bytes, item.Bytes)
	if err := v.UnmarshalBinary(item.Bytes); err != nil {
		return err
	}
	s.printf("read key: %v\n", key)

	return nil
}

// Delete can be used to remove a value from the cache with a given
// key. If the value isn't found, an error is returned.
func (s *stashMemory) Delete(key any) error {
	s.Lock()
	defer s.evict()
	defer s.Unlock()

	if _, ok := s.data[key]; !ok {
		return errors.Errorf("value not found for key: %v", key)
	}
	delete(s.data, key)
	s.printf("deleted key: %v\n", key)

	return nil
}

func (s *stashMemory) Clear() error {
	s.Lock()
	defer s.Unlock()

	//KIM: although we could keep the existing map
	// it makes sense to re-create the pointer to
	// trigger garbage collection instead; this is
	// also...probably...slightly faster
	s.data = nil
	s.data = make(map[any]*stash.CachedItem)
	s.printf("cleared cache")
	return nil
}
