package geecache

import (
	lru2 "gee_cache/geecache/lru"
	"sync"
)

type Cache interface {
	Get(key string) (lru2.Value, bool)
	Add(key string, value lru2.Value)
}

type cache struct {
	mutex      sync.RWMutex
	lru        Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		//c.lru = lru.New(c.cacheBytes, nil)
		c.lru = lru2.NewKCache(c.cacheBytes, 2, 0, nil)
	}
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
