package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes int64
	// 当前已用
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}
type Value interface {
	Size() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(kv.value.Size()) + int64(len(kv.key))
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) update(value Value, ele *list.Element) {
	c.ll.MoveToFront(ele)
	kv := ele.Value.(*entry)
	c.nbytes += int64(value.Size()) - int64(kv.value.Size())
	kv.value = value
}
func (c *Cache) add(key string, value Value) {
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.nbytes += int64(value.Size()) + int64(len(key))
}
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.update(value, ele)
	} else {
		c.add(key, value)
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
func (c *Cache) CacheSize() int64 {
	return c.nbytes
}
