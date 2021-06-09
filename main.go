package main

import (
	"container/list"
	"errors"
	"fmt"
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

func (lc *LRUCache) Put(key, value interface{}) (wasEvicted bool) {
	lc.lock.Lock()

	defer lc.lock.Unlock()
	if kv, ok := lc.cache[key]; ok {
		lc.links.MoveToFront(kv)

		kv.Value.(*pair).value = value

		return false
	} else {
		kv := &pair{key, value}

		k := lc.links.PushFront(kv)
		lc.cache[key] = k
	}

	if lc.links.Len() > lc.capacity {
		if kv := lc.links.Back(); kv != nil {
			lc.PurgeLRUItem(kv)
			lc.TryEvict(kv)

			return true
		}
	}

	return false
}

func (lc *LRUCache) Del(key interface{}) (wasDeleted bool) {
	lc.lock.Lock()

	defer lc.lock.Unlock()
	if kv, ok := lc.cache[key]; ok {
		lc.PurgeLRUItem(kv)

		return true
	}

	return false
}

func (lc *LRUCache) Keys() []interface{} {
	lc.lock.RLock()

	defer lc.lock.RUnlock()
	keys := make([]interface{}, len(lc.cache))

	i := 0
	for k, _ := range lc.cache {
		keys[i] = k
		i++
	}

	return keys
}

func (lc *LRUCache) Peek(key interface{}) (value interface{}) {
	lc.lock.RLock()

	defer lc.lock.RUnlock()
	if v, ok := lc.cache[key]; ok {
		return v.Value.(*pair).value
	}

	return nil
}

func (lc *LRUCache) Has(key interface{}) (ok bool) {
	lc.lock.RLock()

	defer lc.lock.RUnlock()
	_, ok = lc.cache[key]
	return ok
}

func (lc *LRUCache) Purge() {
	lc.lock.Lock()

	defer lc.lock.Unlock()
	for _, v := range lc.cache {
		if lc.onItemEvicted != nil {
			lc.PurgeLRUItem(v)
			lc.TryEvict(v)
		}
	}

	lc.links.Init()
}

func (lc *LRUCache) Size() int {
	lc.lock.Lock()

	defer lc.lock.Unlock()
	return lc.links.Len()
}

func (lc *LRUCache) AdjustCapacity(bufCap int) (numEvicted int) {
	lc.lock.RLock()

	defer lc.lock.RUnlock()
	diff := lc.links.Len() - bufCap

	if diff < 0 {
		diff = 0
	}

	for i := 0; i < diff; i++ {
		if kv := lc.links.Back(); kv != nil {
			lc.PurgeLRUItem(kv)
			lc.TryEvict(kv)
		}
	}

	lc.capacity = bufCap
	return diff
}

/* Utilities */

func (lc *LRUCache) PurgeLRUItem(e *list.Element) {
	lc.links.Remove(e)
	k := e.Value.(*pair)
	delete(lc.cache, k.key)
}

func (lc *LRUCache) TryEvict(e *list.Element) {
	if lc.onItemEvicted != nil {
		kv := e.Value.(*pair)
		lc.onItemEvicted(kv.key, kv.value)
	}
}

func main() {
	lru, ok := New(3, cb)
	if ok != nil {
		fmt.Println("error")
	}
	// | 1 |
	if ok := lru.Put(1, 1); ok {
		fmt.Println("Evicted!", 1)
	}

	// | 2 | 1 |
	if ok := lru.Put(2, 2); ok {
		fmt.Println("Evicted!", 2)
	}

	// | 3 | 2 | 1 |
	if ok := lru.Put(3, 3); ok {
		fmt.Println("Evicted!", 3)
	}

	// | 1 | 3 | 2 |
	v, got := lru.Get(1)

	if got {
		fmt.Println("Get", v)
	}

	// | 4 | 1 | 3 | 2 |
	if ok := lru.Put(4, 4); ok {
		fmt.Println("Evicted!", 4)
	}

	fmt.Println(lru.Keys())
}

func cb(key interface{}, value interface{}) {
	fmt.Println("key is", key)
	fmt.Println("value is", value)
}
