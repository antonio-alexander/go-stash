package memory

import (
	"time"

	"github.com/antonio-alexander/go-stash"
)

func createCacheItem(key interface{}, item stash.Cacheable) (*cacheItem, error) {
	bytes, err := item.MarshalBinary()
	if err != nil {
		return nil, err
	}
	tNow := time.Now()
	return &cacheItem{
		key:          key,
		bytes:        bytes,
		firstCreated: tNow,
		lastRead:     tNow,
		lastUpdated:  tNow,
		nTimesRead:   0,
		size:         len(bytes),
	}, nil
}

func updateCacheItem(cacheItem *cacheItem, item stash.Cacheable) error {
	bytes, err := item.MarshalBinary()
	if err != nil {
		return err
	}
	cacheItem.bytes = bytes
	cacheItem.lastUpdated = time.Now()
	cacheItem.size = len(bytes)
	return nil
}
