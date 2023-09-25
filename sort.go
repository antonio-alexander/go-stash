package stash

import "time"

type ByFirstCreated []*CachedItem

func (s ByFirstCreated) Len() int {
	return len(s)
}

func (s ByFirstCreated) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByFirstCreated) Less(i, j int) bool {
	return time.Unix(0, s[i].FirstCreated).Before(time.Unix(0, s[j].FirstCreated))
}

type ByTimesRead []*CachedItem

func (s ByTimesRead) Len() int {
	return len(s)
}

func (s ByTimesRead) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByTimesRead) Less(i, j int) bool {
	return s[i].NTimesRead < s[j].NTimesRead
}

type ByLastRead []*CachedItem

func (s ByLastRead) Len() int {
	return len(s)
}

func (s ByLastRead) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLastRead) Less(i, j int) bool {
	return time.Unix(0, s[i].LastRead).Before(time.Unix(0, s[j].LastRead))
}
