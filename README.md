# tenure-go

`Tenure-go` is a thread-safe LRU cache instance that uses hashmap lookups and an Open Doubly Linked List to enact the [Least-Recently Used algorithm](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)).

`Tenure-go`'s internal cache utilizes the Go's sync/mutex locking mechanism to ensure thread-safety.
