# tenure-go

`Tenure-go` is a thread-safe LRU cache instance that uses hashmap lookups and an Open Doubly Linked List to enact the [Least-Recently Used algorithm](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)).

`Tenure-go`'s internal cache utilizes the Go's sync/mutex locking mechanism to ensure thread-safety.

## Usage

#### type Callback

```go
type Callback func(key interface{}, value interface{})
```


#### type LRUCache

```go
type LRUCache struct {
}
```


#### func  New

```go
func New(bufCap int, onItemEvicted Callback) (*LRUCache, error)
```
New initializes a new LRU cache with a buffer capacity of `bufCap` It accepts as
a second parameter a callback to be invoked upon successful invocation of the
Least Recently-Used cache policy i.e. when a key/value pair is removed All
transactions utilize locks and are therefore thread-safe

#### func (*LRUCache) AdjustCapacity

```go
func (lc *LRUCache) AdjustCapacity(bufCap int) (numEvicted int)
```
AdjustCapacity resizes the cache capacity Invoking this transaction will evict
all least recently-used items to adjust the cache, where necessary

#### func (*LRUCache) Capacity

```go
func (lc *LRUCache) Capacity() int
```
Capacity returns the current maximum buffer capacity of the cache

#### func (*LRUCache) Del

```go
func (lc *LRUCache) Del(key interface{}) (wasDeleted bool)
```
Del deletes an item corresponding to a given key from the cache, if extant A
boolean flag is returned, indicating whether of not the transaction occurred

#### func (*LRUCache) Drop

```go
func (lc *LRUCache) Drop()
```
Drop drops all items from the cache

#### func (*LRUCache) Get

```go
func (lc *LRUCache) Get(key interface{}) (value interface{}, ok bool)
```
Get attempts to retrieve the value for the given key from the cache Returns the
corresponding value and true if extant; else, returns nil, false Get
transactions will move the item to the head of the cache, designating it as most
recently-used

#### func (*LRUCache) Has

```go
func (lc *LRUCache) Has(key interface{}) (ok bool)
```
Has returns a boolean flag verifying the existence (or lack thereof) of a given
key in the cache without enacting the eviction policy

#### func (*LRUCache) Keys

```go
func (lc *LRUCache) Keys() []interface{}
```
Keys returns a slice of the keys currently extant in the cache

#### func (*LRUCache) LeastRecentlyUsed

```go
func (lc *LRUCache) LeastRecentlyUsed() (key interface{}, value interface{})
```
LeastRecentlyUsed returns the least recently-used key / value pair, or nil if
not extant

#### func (*LRUCache) Put

```go
func (lc *LRUCache) Put(key, value interface{}) (wasEvicted bool)
```
Put adds or inserts a given key / value pair into the cache Put transactions
will move the key to the head of the cache, designating it as 'most
recently-used' If the cache has reached the specified capacity, Put transactions
will also enact the eviction policy thereby removing the least recently-used
item Returns a boolean flag indicating whether an eviction occurred

#### func (*LRUCache) Size

```go
func (lc *LRUCache) Size() int
```
Size returns the current size of the cache

#### type LRUController

```go
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
```
