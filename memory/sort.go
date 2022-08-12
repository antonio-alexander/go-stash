package memory

func toSlice(items map[interface{}]*cacheItem) (cacheItems []*cacheItem) {
	for _, cacheItem := range items {
		cacheItems = append(cacheItems, cacheItem)
	}
	return
}

type byFirstCreated []*cacheItem

func (s byFirstCreated) Len() int { return len(s) }

func (s byFirstCreated) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byFirstCreated) Less(i, j int) bool { return s[i].firstCreated.Before(s[j].firstCreated) }

type byTimesRead []*cacheItem

func (s byTimesRead) Len() int { return len(s) }

func (s byTimesRead) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byTimesRead) Less(i, j int) bool { return s[i].nTimesRead < s[j].nTimesRead }

type byLastRead []*cacheItem

func (s byLastRead) Len() int { return len(s) }

func (s byLastRead) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s byLastRead) Less(i, j int) bool { return s[i].lastRead.Before(s[j].lastRead) }
