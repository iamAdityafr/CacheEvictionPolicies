package main

import (
	"errors"
)

type Node[K comparable, V any] struct {
	key   K
	value V
	prev  *Node[K, V]
	next  *Node[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*Node[K, V]
	head     *Node[K, V]
	tail     *Node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) (*LRUCache[K, V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	cache := &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*Node[K, V]),
	}
	// Simplifying list operations by eliminating edge cases -
	// - empty list or single node by Initializing with dummy head and tail nodes btw
	cache.head = &Node[K, V]{}
	cache.tail = &Node[K, V]{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	return cache, nil
}

func (c *LRUCache[K, V]) addNode(node *Node[K, V]) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

func (c *LRUCache[K, V]) removeNode(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (c *LRUCache[K, V]) moveToHead(node *Node[K, V]) {
	c.removeNode(node)
	c.addNode(node)
}

func (c *LRUCache[K, V]) removeTail() *Node[K, V] {
	if c.head.next == c.tail {
		return nil
	}
	node := c.tail.prev
	c.removeNode(node)
	return node
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	if node, exists := c.cache[key]; exists {
		node.value = value
		c.moveToHead(node)
		return
	}

	newNode := &Node[K, V]{key: key, value: value}
	c.cache[key] = newNode
	c.addNode(newNode)

	if len(c.cache) > c.capacity {
		tail := c.removeTail()
		delete(c.cache, tail.key)
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if node, exists := c.cache[key]; exists {
		c.moveToHead(node)
		return node.value, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Remove(key K) bool {
	if node, exists := c.cache[key]; exists {
		c.removeNode(node)
		delete(c.cache, key)
		return true
	}
	return false
}

func (c *LRUCache[K, V]) Len() int {
	return len(c.cache)
}
