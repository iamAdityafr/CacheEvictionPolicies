package main

import "errors"

type Node[K comparable, V any] struct {
	key        K
	value      V
	prev, next *Node[K, V]
}

type MRUCache[K comparable, V any] struct {
	capacity   int
	cache      map[K]*Node[K, V]
	head, tail *Node[K, V]
}

func NewMRUCache[K comparable, V any](capacity int) (*MRUCache[K, V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	cache := &MRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*Node[K, V]),
	}
	cache.head = &Node[K, V]{}
	cache.tail = &Node[K, V]{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	return cache, nil
}

func (c *MRUCache[K, V]) addNode(node *Node[K, V]) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

func (c *MRUCache[K, V]) removeNode(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}
func (c *MRUCache[K, V]) movetoHead(node *Node[K, V]) {
	c.removeNode(node)
	c.addNode(node)
}

func (c *MRUCache[K, V]) Put(key K, value V) {
	if node, exists := c.cache[key]; exists {
		node.value = value
		c.movetoHead(node)
		return
	}
	var tbRemoved *Node[K, V]
	if len(c.cache) >= c.capacity {
		tbRemoved = c.head.next
	}
	newNode := &Node[K, V]{key: key, value: value}
	c.cache[key] = newNode
	c.addNode(newNode)

	if tbRemoved != nil {
		c.removeNode(tbRemoved)
		delete(c.cache, tbRemoved.key)
	}
}

func (c *MRUCache[K, V]) Get(key K) (V, bool) {
	if node, exists := c.cache[key]; exists {
		c.movetoHead(node)
		return node.value, true
	}
	var zero V
	return zero, false
}
func (c *MRUCache[K, V]) Remove(key K) bool {
	if node, exists := c.cache[key]; exists {
		c.removeNode(node)
		delete(c.cache, key)
		return true
	}
	return false
}
func (c *MRUCache[K, V]) Len() int {
	return len(c.cache)
}
