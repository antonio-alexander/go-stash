package memory

import (
	"github.com/antonio-alexander/go-stash"
)

func toSlice(items map[any]*stash.CachedItem) []*stash.CachedItem {
	cachedItems := make([]*stash.CachedItem, 0, len(items))
	for _, cachedItem := range items {
		cachedItems = append(cachedItems, cachedItem)
	}
	return cachedItems
}
