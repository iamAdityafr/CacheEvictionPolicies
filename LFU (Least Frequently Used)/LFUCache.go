package main

import "errors"

type Node[K comparable, V any] struct {
	key        K
	value      V
	freq       int
	prev, next *Node[K, V]
}
type DLL[K comparable, V any] struct {
	head, tail *Node[K, V]
}
type LFUCache[K comparable, V any] struct {
	capacity int
	minFreq  int
	cache    map[K]*Node[K, V]
	freqMap  map[int]*DLL[K, V]
}

func NewLFUCache[K comparable, V any](capacity int) (*LFUCache[K, V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	cache := &LFUCache[K, V]{
		capacity: capacity,
		minFreq:  0,
		cache:    make(map[K]*Node[K, V]),
		freqMap:  make(map[int]*DLL[K, V]),
	}
	return cache, nil
}
func NewDLL[K comparable, V any]() *DLL[K, V] {
	dll := &DLL[K, V]{
		head: &Node[K, V]{},
		tail: &Node[K, V]{},
	}
	dll.head.next = dll.tail
	dll.tail.prev = dll.head
	return dll
}

func (dll *DLL[K, V]) addNode(node *Node[K, V]) {
	node.prev = dll.head
	node.next = dll.head.next
	dll.head.next.prev = node
	dll.head.next = node
}
func (dll *DLL[K, V]) removeNode(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}
func (dll *DLL[K, V]) removeLast() *Node[K, V] {
	if dll.tail.prev == dll.head {
		return nil
	}
	tbDeleted := dll.tail.prev
	tbDeleted.prev.next = dll.tail
	dll.tail.prev = tbDeleted.prev
	return tbDeleted
}
func (dll *DLL[K, V]) isEmpty() bool {
	return dll.tail.prev == dll.head
}
func (dll *DLL[K, V]) addFront(node *Node[K, V]) {
	node.prev = dll.head
	node.next = dll.head.next
	dll.head.next.prev = node
	dll.head.next = node
}
func (lfu *LFUCache[K, V]) addNodeToFreqList(freq int, node *Node[K, V]) {
	if dll, exists := lfu.freqMap[freq]; exists {
		dll.addFront(node)
	} else {
		newList := NewDLL[K, V]()
		newList.addFront(node)
		lfu.freqMap[freq] = newList
	}
}
func (lfu *LFUCache[K, V]) removeNodeFromFreqList(node *Node[K, V]) {
	if dll, exists := lfu.freqMap[node.freq]; exists {
		dll.removeNode(node)
		if dll.isEmpty() {
			delete(lfu.freqMap, node.freq)
		}
	}
}
func (lfu *LFUCache[K, V]) updateMinFreq() {
	for {
		if dll, exists := lfu.freqMap[lfu.minFreq]; !exists || dll.isEmpty() {
			lfu.minFreq++
		} else {
			break
		}
	}
}
func (lfu *LFUCache[K, V]) Get(key K) (V, bool) {
	if node, exists := lfu.cache[key]; exists {
		oldFreq := node.freq
		lfu.removeNodeFromFreqList(node)
		node.freq++
		lfu.addNodeToFreqList(node.freq, node)
		if oldFreq == lfu.minFreq {
			lfu.updateMinFreq()
		}
		return node.value, true
	}
	var zero V
	return zero, false
}
func (lfu *LFUCache[K, V]) Put(key K, value V) {
	if node, exists := lfu.cache[key]; exists {
		node.value = value
		oldFreq := node.freq
		node.freq++
		lfu.removeNodeFromFreqList(node)
		lfu.addNodeToFreqList(node.freq, node)
		if oldFreq == lfu.minFreq {
			lfu.updateMinFreq()
		}
		return
	}
	if len(lfu.cache) >= lfu.capacity {
		if dll, exists := lfu.freqMap[lfu.minFreq]; exists {
			tbRemoved := dll.removeLast()
			if tbRemoved != nil {
				delete(lfu.cache, tbRemoved.key)
			}
			if dll.isEmpty() {
				delete(lfu.freqMap, lfu.minFreq)
			}
		}
	}
	newNode := &Node[K, V]{key: key, value: value, freq: 1}
	lfu.cache[key] = newNode
	lfu.addNodeToFreqList(1, newNode)
	lfu.minFreq = 1
}

// just a util func
func (lfu *LFUCache[K, V]) evict() {
	if dll, exists := lfu.freqMap[lfu.minFreq]; exists {
		tbRemoved := dll.removeLast()
		if tbRemoved != nil {
			delete(lfu.cache, tbRemoved.key)
		}
		if dll.isEmpty() {
			delete(lfu.freqMap, lfu.minFreq)
		}
	}
}
