package tenure

import (
	"container/list"
	"errors"
	"sync"
)

type Callback func(key interface{}, value interface{})

type LRUController interface {
	Get(key interface{}) (value interface{}, ok bool)
	Put(key, value interface{}) (wasEvicted bool)
	Del(key interface{}) (wasDeleted bool)
	Keys() []interface{}
	Peek(key interface{}) (value interface{})
	Has(key interface{}) (ok bool)
	Purge()
	Size() int
	AdjustCapacity(bufCap int) (numEvicted int)
}

type LRUCache struct {
	capacity      int
	links         *list.List
	cache         map[interface{}]*list.Element
	onItemEvicted Callback
	lock          sync.RWMutex
}

type pair struct {
	key   interface{}
	value interface{}
}

// New initializes a new LRU cache with a buffer capacity of `bufCap`
// It accepts as a second parameter a callback to be invoked upon successful invocation
// of the Least Recently-Used cache policy i.e. when a key/value pair is removed
// All transactions utilize locks and are therefore thread-safe
func New(bufCap int, onItemEvicted Callback) (*LRUCache, error) {
	if bufCap <= 0 {
		return nil, errors.New("an LRU Cache must be initialized with a whole number greater than zero")
	}

	c := &LRUCache{
		capacity:      bufCap,
		links:         list.New(),
		cache:         make(map[interface{}]*list.Element, bufCap),
		onItemEvicted: onItemEvicted,
	}
	return c, nil
}

// Get attempts to retrieve the value for the given key from the cache
// Returns the corresponding value and true if extant; else, returns nil, false
// Get transactions will move the item to the head of the cache, designating it as most recently-used
func (lc *LRUCache) Get(key interface{}) (value interface{}, ok bool) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if kv, ok := lc.cache[key]; ok {
		lc.links.MoveToFront(kv)

		if kv.Value.(*pair) == nil {
			return nil, false
		}

		return kv.Value.(*pair).value, true
	}

	return nil, false
}

// Put adds or inserts a given key / value pair into the cache
// Put transactions will move the key to the head of the cache, designating it as 'most recently-used'
// If the cache has reached the specified capacity, Put transactions will also enact the eviction policy
// thereby removing the least recently-used item
// Returns a boolean flag indicating whether an eviction occurred
func (lc *LRUCache) Put(key, value interface{}) (wasEvicted bool) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if kv, ok := lc.cache[key]; ok {
		lc.links.MoveToFront(kv)

		kv.Value.(*pair).value = value

		return false
	}

	kv := &pair{key, value}

	k := lc.links.PushFront(kv)
	lc.cache[key] = k

	if lc.links.Len() > lc.capacity {
		if kv := lc.links.Back(); kv != nil {
			lc.purgeLRUItem(kv)
			lc.tryEvict(kv)

			return true
		}
	}

	return false
}

// Del deletes an item corresponding to a given key from the cache, if extant
// A boolean flag is returned, indicating whether of not the transaction occurred
func (lc *LRUCache) Del(key interface{}) (wasDeleted bool) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	if kv, ok := lc.cache[key]; ok {
		lc.purgeLRUItem(kv)

		return true
	}

	return false
}

// Keys returns a slice of the keys currently extant in the cache
func (lc *LRUCache) Keys() []interface{} {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	keys := make([]interface{}, lc.links.Len())

	for i, k := 0, lc.links.Back(); k != nil; k = k.Prev() {
		keys[i] = k.Value.(*pair).key
		i++
	}

	return keys
}

// Has returns a boolean flag verifying the existence (or lack thereof)
// of a given key in the cache without enacting the eviction policy
func (lc *LRUCache) Has(key interface{}) (ok bool) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	_, ok = lc.cache[key]
	return
}

// Drop drops all items from the cache
func (lc *LRUCache) Drop() {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	for _, v := range lc.cache {
		if lc.onItemEvicted != nil {
			lc.purgeLRUItem(v)
			lc.tryEvict(v)
		}
	}

	lc.links.Init()
}

// Size returns the current size of the cache
func (lc *LRUCache) Size() int {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	return lc.links.Len()
}

// Capacity returns the current maximum buffer capacity of the cache
func (lc *LRUCache) Capacity() int {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	return lc.capacity
}

// AdjustCapacity resizes the cache capacity
// Invoking this transaction will evict all least recently-used items
// to adjust the cache, where necessary
func (lc *LRUCache) AdjustCapacity(bufCap int) (numEvicted int) {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	diff := lc.links.Len() - bufCap

	if diff < 0 {
		diff = 0
	}

	for i := 0; i < diff; i++ {
		if kv := lc.links.Back(); kv != nil {
			lc.purgeLRUItem(kv)
			lc.tryEvict(kv)
		}
	}

	lc.capacity = bufCap

	return diff
}

// LeastRecentlyUsed returns the least recently-used key / value pair, or nil if not extant
func (lc *LRUCache) LeastRecentlyUsed() (key interface{}, value interface{}) {
	kv := lc.links.Back()
	if kv != nil {
		n := kv.Value.(*pair)
		key, value = n.key, n.value
		return
	}
	return
}

/* Utilities */

func (lc *LRUCache) purgeLRUItem(e *list.Element) {
	lc.links.Remove(e)
	kv := e.Value.(*pair)
	delete(lc.cache, kv.key)
}

func (lc *LRUCache) tryEvict(e *list.Element) {
	if lc.onItemEvicted != nil {
		kv := e.Value.(*pair)
		lc.onItemEvicted(kv.key, kv.value)
	}
}
