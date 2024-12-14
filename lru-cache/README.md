# My own LRU cache

An LRU **(Least Recently Used)** cache is a type of cache that uses an algorithm to manage the items stored in it. Its main goal is to optimize access to frequently used data, reducing latency and improving overall system performance.

### How does it work?

The LRU algorithm is based on the principle that data that has been used recently is likely to be used again soon. Therefore, the cache keeps track of the order in which elements have been accessed. When the cache is full and space is needed for a new item:

1. The least recently used (LRU) item is identified. This is the element that has gone the longest without being accessed.
2. The LRU element is removed. The space it occupied is freed up.
3. The new item is added to the cache.

This way, the cache always contains the most recently used items, maximizing the likelihood that data requests will find the information in the cache rather than having to access a slower data source.

### How to run 
```bash
cd cmd
go run .
```
