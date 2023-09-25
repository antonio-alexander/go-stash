package stash

import "time"

func CreateCacheItem(key interface{}, item Cacheable) (*CachedItem, error) {
	bytes, err := item.MarshalBinary()
	if err != nil {
		return nil, err
	}
	tNow := time.Now()
	return &CachedItem{
		Key:          key,
		Bytes:        bytes,
		FirstCreated: tNow.UnixNano(),
		LastRead:     tNow.UnixNano(),
		LastUpdated:  tNow.UnixNano(),
		NTimesRead:   0,
		Size:         len(bytes),
	}, nil
}

func UpdateCacheItem(cacheItem *CachedItem, item Cacheable) error {
	bytes, err := item.MarshalBinary()
	if err != nil {
		return err
	}
	cacheItem.Bytes = bytes
	cacheItem.LastUpdated = time.Now().UnixNano()
	cacheItem.Size = len(bytes)
	return nil
}
