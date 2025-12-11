package main

import (
	"sync"
	"time"
)

type Node[K comparable, V any] struct {
	key        K
	value      V
	expiryTime int64
	slot       *slot[K, V]
	prev, next *Node[K, V]
}

type slot[K comparable, V any] struct {
	head, tail *Node[K, V]
}

type TTLCache[K comparable, V any] struct {
	cache    map[K]*Node[K, V]
	wheel    [3][]slot[K, V] // 3 level timing wheel
	tick     uint32          // global tick
	mu       sync.RWMutex
	stopCh   chan struct{}
	onceStop sync.Once
}

func NewTTLCache[K comparable, V any]() (*TTLCache[K, V], error) {
	cache := &TTLCache[K, V]{
		cache:  make(map[K]*Node[K, V]),
		stopCh: make(chan struct{}),
	}

	cache.wheel[0] = make([]slot[K, V], 512) // wheel 0 -> 1 ms slots, 512 ms total
	cache.wheel[1] = make([]slot[K, V], 256) // wheel 1 -> 512 ms slots, 131 s total
	cache.wheel[2] = make([]slot[K, V], 256) // wheel 2 -> 131 s slots, ~9.4 h total

	for i := 0; i < 3; i++ {
		for j := range cache.wheel[i] {
			headDummy := &Node[K, V]{}
			tailDummy := &Node[K, V]{}
			headDummy.next = tailDummy
			tailDummy.prev = headDummy

			cache.wheel[i][j].head = headDummy
			cache.wheel[i][j].tail = tailDummy
		}
	}
	go cache.startTicker()
	return cache, nil
}

// this is the global ticker
func (c *TTLCache[K, V]) startTicker() {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.tick++
			c.expireWheel(0)
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

func (c *TTLCache[K, V]) expireWheel(level int) {
	w := &c.wheel[level][c.tick%uint32(len(c.wheel[level]))]

	for e := w.head.next; e != w.tail; {
		next := e.next
		if time.Now().UnixNano() >= e.expiryTime {
			delete(c.cache, e.key)
			w.remove(e)
		}
		e = next
	}

	if c.tick%uint32(len(c.wheel[level])) == 0 && level+1 < 3 {
		c.expireWheel(level + 1)
	}
}

func (s *slot[K, V]) add(e *Node[K, V]) {
	previous := s.tail.prev
	previous.next = e
	e.prev = previous
	e.next = s.tail
	s.tail.prev = e
}

func (s *slot[K, V]) remove(e *Node[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	}
	e.prev = nil
	e.next = nil
}
func (c *TTLCache[K, V]) Set(key K, value V, ttl time.Duration) {
	expiry := time.Now().Add(ttl).UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	if old, ok := c.cache[key]; ok {
		if old.slot != nil {
			old.slot.remove(old)
		}
		delete(c.cache, key)
	}
	e := &Node[K, V]{
		key:        key,
		value:      value,
		expiryTime: expiry,
	}

	c.insertEntry(e, ttl)
	c.cache[key] = e
}
func (c *TTLCache[K, V]) insertEntry(e *Node[K, V], ttl time.Duration) {
	ms := ttl.Milliseconds()
	var wheelIdx, slotIdx int

	if ms < 512 {
		wheelIdx = 0
		slotIdx = int((c.tick + uint32(ms)) % 512)
	} else if ms < 512*256 {
		wheelIdx = 1
		slotIdx = int((c.tick/512 + uint32(ms)/512) % 256)
	} else {
		if ms > 256*131072 {
			ms = 256 * 131072
			e.expiryTime = time.Now().Add(time.Duration(ms) * time.Millisecond).UnixNano()
		}
		wheelIdx = 2
		slotIdx = int((c.tick/131072 + uint32(ms)/131072) % 256)
	}

	slot := &c.wheel[wheelIdx][slotIdx]
	slot.add(e)
	e.slot = slot
}
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	if time.Now().UnixNano() > e.expiryTime {
		var zero V
		return zero, false
	}

	return e.value, true
}
func (c *TTLCache[K, V]) Stop() {
	c.onceStop.Do(func() {
		close(c.stopCh)
	})
}
