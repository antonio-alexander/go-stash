package memory

import (
	"github.com/antonio-alexander/go-stash"
)

func toSlice(items map[interface{}]*stash.CachedItem) (cachedItems []*stash.CachedItem) {
	for _, cachedItem := range items {
		cachedItems = append(cachedItems, cachedItem)
	}
	return
}
