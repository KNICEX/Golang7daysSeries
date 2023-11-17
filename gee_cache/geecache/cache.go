package geecache

import (
	"gee_cache/geecache/lru"
	"sync"
)

type Cache interface {
	Get(key string) (lru.Value, bool)
	Add(key string, value lru.Value)
	Delete(key string)
	OnEvicted(fn func(key string, value lru.Value))
}

type cache struct {
	mutex sync.RWMutex
	lru   Cache
}

func newCache(cacheBytes int64) *cache {
	return &cache{
		lru: lru.New(cacheBytes, nil),
	}
}
func newKCache(cacheBytes int64, k int, maxHistory int) *cache {
	return &cache{
		lru: lru.NewKCache(cacheBytes, k, maxHistory, nil),
	}
}

func (c *cache) onEvicted(fn func(key string, view ByteView)) {
	c.lru.OnEvicted(func(key string, value lru.Value) {
		fn(key, value.(ByteView))
	})
}

func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

func (c *cache) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lru.Delete(key)
}
