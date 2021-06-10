package tenure

import (
	"testing"
)

func TestEvictionPolicy(t *testing.T) {
	maxcap := 256
	evictions := 0

	incr := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evicted instances not synced; Have (k=%v,v=%v), Want (k=v)", k, v)
		}
		evictions++
	}

	lru, err := New(maxcap, incr)

	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	for i := 0; i < maxcap*2; i++ {
		lru.Put(i, i)
	}

	if lru.Size() != maxcap {
		t.Fatalf("Cache capacity failure; Have %v, Want %v", lru.Size(), maxcap)
	}

	if evictions != maxcap {
		t.Fatalf("Cache eviction failure; Have %v, Want %v", evictions, maxcap)
	}

	for i, k := range lru.Keys() {
		v, ok := lru.Get(k)

		if !ok {
			t.Fatalf("Key retrieval failure")
		}

		if v != k {
			t.Fatalf("Invalid key; Have %v, Want %v", v, k)
		}

		if v != i+maxcap {
			t.Fatalf("Invalid key; Have %v, Want %v", v, i+maxcap)
		}
	}

	for i := 0; i < maxcap; i++ {
		if ok := lru.Has(i); ok {
			t.Fatalf("Cache contains stale value; %v should have been evicted", i)
		}
	}

	for i := maxcap; i < maxcap*2; i++ {
		if _, ok := lru.Get(i); !ok {
			t.Fatalf("Premature cache eviction; %v should not have been evicted", i)
		}
	}

}

func TestRemoval(t *testing.T) {
	maxcap := 9

	noop := func(k interface{}, v interface{}) {}

	lru, err := New(maxcap, noop)
	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	for i := 0; i <= maxcap; i++ {
		lru.Put(i, i)
	}

	if lru.Size() != maxcap {
		t.Fatalf("Size mismatch; Have %v, Want %v", lru.Size(), maxcap)
	}

	c := 0
	r := func(k int) {
		c++
		if ok := lru.Del(k); !ok {
			t.Fatalf("Failed to delete item %v", k)
		}

		if lru.Size() != maxcap-c {
			t.Fatalf("Size mismatch; Have %v, Want %v", lru.Size(), maxcap-c)
		}

		if lru.Has(k) {
			t.Fatalf("Failed to delete key %v", k)
		}
	}

	r(maxcap)
	r(maxcap - 1)
	r(maxcap - 3)
}

func TestLeastRecentlyUsed(t *testing.T) {
	maxcap := 3
	evictions := 0

	incr := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evicted instances not synced; Have (k=%v,v=%v), Want (k=v)", k, v)
		}
		evictions++
	}

	lru, err := New(maxcap, incr)

	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	lru.Put(1, 1)
	lru.Put(2, 2)
	lru.Get(1)
	lru.Put(3, 3)
	lru.Get(1)
	lru.Put(4, 4)
	lru.Get(1)
	lru.Put(5, 5)

	if lru.Size() != maxcap {
		t.Fatalf("Eviction policy failure; Have size %v, Want size %v", lru.Size(), maxcap)
	}

	if _, v := lru.LeastRecentlyUsed(); v != 4 {
		t.Fatalf("Least recently used failure; Have %v, Want %v", v, 4)
	}

	if evictions != 2 {
		t.Fatalf("Eviction policy failure; Have %v evictions, Want %v evictions", evictions, 2)
	}

	evictions = 0
	lru.Drop()

	if evictions != maxcap {
		t.Fatalf("Expected drop to remove all cached values; Have %v evictions, Want %v evictions", evictions, maxcap)
	}

	if len(lru.Keys()) != 0 {
		t.Fatalf("Expected drop to remove all keys; Have %v keys, Want %v keys", len(lru.Keys()), 0)
	}
}

func TestHasIsInconsequential(t *testing.T) {
	maxcap := 9
	evictions := 0

	incr := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evicted instances not synced; Have (k=%v,v=%v), Want (k=v)", k, v)
		}
		evictions++
	}

	lru, err := New(maxcap, incr)
	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	for i := maxcap + 1; i < maxcap*2; i++ {
		lru.Put(i, i)
		lru.Has(i)
	}
	if evictions != 0 {
		t.Fatalf("Has should not trigger the eviction policy; Have %v evictions, Want %v evictions", evictions, 0)
	}
}

func TestCapAdjustment(t *testing.T) {
	maxcap := 9
	evictions := 0

	incr := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evicted instances not synced; Have (k=%v,v=%v), Want (k=v)", k, v)
		}
		evictions++
	}

	lru, err := New(maxcap, incr)
	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	for i := 0; i < maxcap; i++ {
		lru.Put(i, i)
	}

	lru.AdjustCapacity(maxcap / 3)
	if evictions != 6 {
		t.Fatalf("Eviction policy failed; Have %v evictions, Want %v evictions", evictions, 6)
	}

	lru.AdjustCapacity(lru.Size() + maxcap)

	for i := 0; i <= lru.Capacity(); i++ {
		lru.Put(i, i)
	}

	if evictions != 6+1 {
		t.Fatalf("Eviction policy failed; Have %v evictions, Want %v evictions", evictions, 6+1)
	}
}

func TestMitigations(t *testing.T) {
	maxcap := 9
	evictions := 0

	incr := func(k interface{}, v interface{}) {
		if k != v {
			t.Fatalf("Evicted instances not synced; Have (k=%v,v=%v), Want (k=v)", k, v)
		}
		evictions++
	}

	lru, err := New(maxcap, incr)
	if err != nil {
		t.Fatalf("Failed to initialize a new LRU cache instance; see %v", err)
	}

	if k, v := lru.LeastRecentlyUsed(); k != nil || v != nil {
		t.Fatal("LRU should be nil")
	}

	if lru.Del(9) {
		t.Fatal("Deleting a non-extant value should return false")
	}

	if lru.Has(9) {
		t.Fatal("Has used with a non-extant key should return false")
	}
}
