package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type RandomCache[K comparable, V any] struct {
	mu   sync.Mutex
	data map[K]Entry[V]
	keys []K
	cap  int
	rnd  *rand.Rand
}

type Entry[V any] struct {
	value    V
	expireAt time.Time
}

func NewRandomCache[K comparable, V any](capacity int) (*RandomCache[K, V], error) {
	if capacity < 0 {
		return nil, errors.New("capacity must be positive")
	}
	cache := &RandomCache[K, V]{
		data: make(map[K]Entry[V]),
		keys: make([]K, 0, capacity),
		cap:  capacity,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return cache, nil
}

func (c *RandomCache[K, V]) evictRandom() {
	if len(c.keys) == 0 {
		return
	}
	// prioritizing removing the expired ones for only the first 5 since we are also setting up with TTL btw
	for i := 0; i < 5; i++ {
		idx := c.rnd.Intn(len(c.keys))
		key := c.keys[idx]
		entry := c.data[key]
		if c.isExpired(entry) {
			delete(c.data, key)
			c.Remove(key)
			return
		}
	}
	idx := c.rnd.Intn(len(c.keys))
	tbDeleted := c.keys[idx]
	delete(c.data, tbDeleted)
	lastIndex := len(c.keys) - 1
	c.keys[idx] = c.keys[lastIndex]
	c.keys = c.keys[:lastIndex]
}
func (c *RandomCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, exists := c.data[key]
	if !exists {
		var zero V
		return zero, false
	}
	if c.isExpired(v) {
		delete(c.data, key)
		c.Remove(key)
		var zero V
		return zero, false
	}
	return v.value, true

}

func (c *RandomCache[K, V]) Put(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, exists := c.data[key]; exists {
		v.value = val
		c.data[key] = v
		return
	}
	if len(c.data) >= c.cap {
		c.evictRandom()
	}

	c.data[key] = Entry[V]{value: val, expireAt: time.Time{}}
	c.keys = append(c.keys, key)
}

func (c *RandomCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.Remove(key)
		return true
	}
	return false
}

// only to remove that expired ones
func (c *RandomCache[K, V]) SetWithTTL(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, exists := c.data[key]; exists {
		v.value = val
		v.expireAt = time.Now().Add(ttl)
		c.data[key] = v
		return
	}
	if c.cap <= len(c.data) {
		c.evictRandom()
	}
	c.data[key] = Entry[V]{value: val, expireAt: time.Now().Add(ttl)}
	c.keys = append(c.keys, key)
}

// these are just util funcs
func (c *RandomCache[K, V]) Len() int {
	return len(c.data)
}
func (c *RandomCache[K, V]) Remove(key K) {
	for i, k := range c.keys {
		if k == key {
			last := len(c.keys) - 1
			c.keys[i] = c.keys[last]
			c.keys = c.keys[:last]
			break
		}
	}
}
func (c *RandomCache[K, V]) Capacity() int {
	return c.cap
}

func (c *RandomCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]Entry[V], c.cap)
	c.keys = c.keys[:0]
}

func (c *RandomCache[K, V]) isExpired(e Entry[V]) bool {
	return time.Now().After(e.expireAt) && !e.expireAt.IsZero()
}
