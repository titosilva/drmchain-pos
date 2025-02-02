package clru

import "sync"

type Cache[K comparable, V any] struct {
	cache     map[K]*clruNode[K, V]
	head      *clruNode[K, V]
	tail      *clruNode[K, V]
	remaining int
	mux       *sync.Mutex
}

func New[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		cache:     make(map[K]*clruNode[K, V]),
		head:      nil,
		tail:      nil,
		remaining: capacity,
		mux:       &sync.Mutex{},
	}
}

func (clru *Cache[K, V]) Get(key K) (V, bool) {
	clru.mux.Lock()
	defer clru.mux.Unlock()
	node, exists := clru.cache[key]
	if !exists {
		return *new(V), false
	}

	clru.promote(node)
	return node.value, true
}

func (clru *Cache[K, V]) promote(node *clruNode[K, V]) {
	if node == clru.head {
		return
	}

	if node == clru.tail {
		clru.tail = node.prev
		clru.tail.next = nil
	} else {
		node.prev.next = node.next
		node.next.prev = node.prev
	}

	node.next = clru.head
	clru.head.prev = node
	clru.head = node
}

func (clru *Cache[K, V]) Put(key K, value V) {
	clru.mux.Lock()
	defer clru.mux.Unlock()
	node, exists := clru.cache[key]
	if exists {
		node.value = value
		clru.promote(node)
		return
	}

	if clru.remaining == 0 {
		clru.evict()
	}

	node = &clruNode[K, V]{
		key:   key,
		value: value,
		next:  clru.head,
		prev:  nil,
	}

	if clru.head != nil {
		clru.head.prev = node
	}

	clru.head = node
	if clru.tail == nil {
		clru.tail = node
	}

	clru.cache[key] = node
	clru.remaining--
}

func (clru *Cache[K, V]) evict() {
	if clru.tail == nil {
		return
	}

	delete(clru.cache, clru.tail.key)

	if clru.tail == clru.head {
		clru.head = nil
		clru.tail = nil
	} else {
		clru.tail = clru.tail.prev
		clru.tail.next = nil
	}

	clru.remaining++
}
