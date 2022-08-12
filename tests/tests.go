package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/antonio-alexander/go-stash"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func generateId() string {
	return uuid.Must(uuid.NewRandom()).String()
}

//TestStash validates the basic functions of being able to read, write and delete
// data in a store that is cacheable
func TestStash(t *testing.T, newFx func() stash.Stasher) func(*testing.T) {
	return func(t *testing.T) {
		s := newFx()
		assert.NotNil(t, s)

		//generate example
		example := &stash.Example{
			Int:    rand.Int(),
			Float:  rand.Float64(),
			String: generateId(),
		}
		key := generateId()

		//write
		replaced, err := s.Write(key, example)
		assert.Nil(t, err)
		assert.False(t, replaced)

		//read
		exampleRead := &stash.Example{}
		err = s.Read(key, exampleRead)
		assert.Nil(t, err)
		assert.Equal(t, example, exampleRead)

		//write
		replaced, err = s.Write(key, example)
		assert.Nil(t, err)
		assert.True(t, replaced)

		//delete
		err = s.Delete(key)
		assert.Nil(t, err)

		//read
		exampleRead = &stash.Example{}
		err = s.Read(key, exampleRead)
		assert.NotNil(t, err)

		//delete
		err = s.Delete(key)
		assert.NotNil(t, err)
	}
}

//TestEvictSize will validate that a stash, configured with a set size (in bytes) will evict
// data (like a FIFO) when that maximum size is reached
func TestEvictSize(t *testing.T, newFx func(timeToLive time.Duration, maxSize int) interface {
	stash.Stasher
}) func(*testing.T) {
	return func(t *testing.T) {
		const exampleSize = 97

		//generate common values
		key1, key2 := generateId(), generateId()
		example1 := &stash.Example{String: generateId()}
		example2 := &stash.Example{String: generateId()}

		//test eviction using size
		s := newFx(0, exampleSize)
		assert.NotNil(t, s)

		//write key/example 1
		replaced, err := s.Write(key1, example1)
		assert.Nil(t, err)
		assert.False(t, replaced)
		exampleRead := &stash.Example{}
		err = s.Read(key1, exampleRead)
		assert.Nil(t, err)
		assert.Equal(t, example1, exampleRead)

		//write key/example 2
		replaced, err = s.Write(key2, example2)
		assert.Nil(t, err)
		assert.False(t, replaced)
		exampleRead = &stash.Example{}
		err = s.Read(key2, exampleRead)
		assert.Nil(t, err)
		assert.Equal(t, example2, exampleRead)

		//KIM: the evictions occur AFTER reading
		//validate example/key 1 has been evicted
		exampleRead = &stash.Example{}
		err = s.Read(key1, exampleRead)
		assert.NotNil(t, err)
	}
}

//TestEvictLeastRecentlyUsed can be used to validate the ability to evict data that has been used
// less than other data...
func TestEvictLeastRecentlyUsed(t *testing.T, newFx func(timeToLive time.Duration, maxSize int) interface {
	stash.Stasher
}) func(*testing.T) {
	return func(t *testing.T) {
		//generate example data
		ex := &stash.Example{String: generateId()}
		bytes, _ := ex.MarshalBinary()
		exampleLength := len(bytes) //49

		//test eviction using LeastRecentlyUsed
		s := newFx(2*time.Millisecond, exampleLength+1)
		assert.NotNil(t, s)
		keys := []string{generateId(), generateId()}
		examples := []*stash.Example{{String: generateId()}, {String: generateId()}}
		_, err := s.Write(keys[0], examples[0])
		assert.Nil(t, err)
		time.Sleep(time.Millisecond)
		_, err = s.Write(keys[1], examples[1])
		assert.Nil(t, err)
		time.Sleep(time.Millisecond)
		exampleRead := &stash.Example{}
		err = s.Read(keys[1], exampleRead)
		assert.Nil(t, err)
		exampleRead = &stash.Example{}
		err = s.Read(keys[0], exampleRead)
		assert.NotNil(t, err)
	}
}

//TestEvictLeastFrequentlyUsed can be used to validate that data that isn't used
// frequently is evicted
func TestEvictLeastFrequentlyUsed(t *testing.T, newFx func(timeToLive time.Duration, maxSize int) interface {
	stash.Stasher
}) func(*testing.T) {
	return func(t *testing.T) {
		//generate example data
		keys := []string{generateId(), generateId()}
		examples := []*stash.Example{{String: generateId()}, {String: generateId()}}

		//test LeastFrequentlyUsed
		s := newFx(time.Second, 0)
		assert.NotNil(t, s)
		_, err := s.Write(keys[0], examples[0])
		assert.Nil(t, err)
		_, err = s.Write(keys[1], examples[1])
		assert.Nil(t, err)
		exampleRead := &stash.Example{}
		err = s.Read(keys[0], exampleRead)
		assert.Nil(t, err)
		time.Sleep(time.Second)
		exampleRead = &stash.Example{}
		err = s.Read(keys[0], exampleRead)
		assert.Nil(t, err)
		exampleRead = &stash.Example{}
		err = s.Read(keys[1], exampleRead)
		assert.NotNil(t, err)
	}
}

//TestEvictFirstInFirstOut can be used to validate the FIFO based eviction
func TestEvictFirstInFirstOut(t *testing.T, newFx func(timeToLive time.Duration, maxSize int) interface {
	stash.Stasher
}) func(*testing.T) {
	return func(t *testing.T) {
		const exampleSize = 97

		//generate common values
		key1, key2 := generateId(), generateId()
		example1 := &stash.Example{String: generateId()}
		example2 := &stash.Example{String: generateId()}

		//test eviction using size
		s := newFx(0, exampleSize)
		assert.NotNil(t, s)

		//
		replaced, err := s.Write(key1, example1)
		assert.Nil(t, err)
		assert.False(t, replaced)
		exampleRead := &stash.Example{}
		err = s.Read(key1, exampleRead)
		assert.Nil(t, err)
		assert.Equal(t, example1, exampleRead)

		//
		replaced, err = s.Write(key2, example2)
		assert.Nil(t, err)
		assert.False(t, replaced)
		exampleRead = &stash.Example{}
		err = s.Read(key2, exampleRead)
		assert.Nil(t, err)
		assert.Equal(t, example2, exampleRead)

		//KIM: the evictions occur AFTER reading
		exampleRead = &stash.Example{}
		err = s.Read(key1, exampleRead)
		assert.NotNil(t, err)
	}
}
