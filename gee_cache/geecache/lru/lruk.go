package lru

import (
	"container/list"
)

type historyEntry struct {
	key   string
	count int
}

type KCache struct {
	k          int
	hl         *list.List
	maxHistory int
	history    map[string]*list.Element
	*Cache
}

func NewKCache(maxBytes int64, k int, maxHistory int, onEvicted func(string, Value)) *KCache {
	return &KCache{
		k:          k,
		hl:         list.New(),
		maxHistory: maxHistory,
		history:    make(map[string]*list.Element),
		Cache:      New(maxBytes, onEvicted),
	}
}

func (k *KCache) RemoveOldestHistory() {
	ele := k.hl.Back()
	if ele != nil {
		k.hl.Remove(ele)
		delete(k.history, ele.Value.(*historyEntry).key)
	}
}

func (k *KCache) Add(key string, value Value) {
	if ele, ok := k.Cache.cache[key]; ok {
		// 在缓存中, 更新位置
		k.Cache.update(value, ele)
	} else {
		if history, ok := k.history[key]; ok {
			// 在历史记录中 增加记录count
			history.Value.(*historyEntry).count++
			if history.Value.(*historyEntry).count >= k.k {
				// 达到k次添加进Cache
				k.Cache.add(key, value)
				// 从history移除
				k.hl.Remove(history)
				delete(k.history, history.Value.(*historyEntry).key)
			} else {
				k.hl.MoveToFront(history)
			}
		} else {
			// 添加进history
			if k.k == 1 {
				k.Cache.add(key, value)
			} else {
				ele := k.hl.PushFront(&historyEntry{key: key, count: 1})
				k.history[key] = ele
			}
			if k.maxHistory != 0 && k.hl.Len() > k.maxHistory {
				k.RemoveOldestHistory()
			}
		}
	}
	for k.maxBytes != 0 && k.maxBytes < k.nbytes {
		k.RemoveOldest()
	}
}

func (k *KCache) HistoryLen() int {
	return k.hl.Len()
}
