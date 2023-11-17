package geecache

import (
	"fmt"
	"gee_cache/geecache/singleflight"
	"log"
	"sync"
	"testing"
	"time"
)

var db = make(map[string]string, 1000)

func initDB() {
	for i := 0; i < 30; i++ {
		db[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
}

func TestBatch(t *testing.T) {
	initDB()
	itemGet(t)
	start := time.Now()
	for i := 3; i > 0; i-- {
		itemGet(t)
	}
	fmt.Println("average time ", time.Since(start)/3)
}

func itemGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))

	var mutex sync.RWMutex

	gee := NewKGroup("score", 2<<10, 3, 0, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key: ", key)
			mutex.Lock()
			loadCounts[key]++
			mutex.Unlock()
			time.Sleep(time.Millisecond * 40)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			} else {
				return nil, fmt.Errorf("not found key %s", key)
			}
		},
	))
	gee.Loader(singleflight.NewWGLoader())
	start := time.Now()
	var n sync.WaitGroup
	for i := 300; i > 0; i-- {
		go func() {
			n.Add(1)
			for k, v := range db {
				if view, err := gee.Get(k); err != nil || view.String() != v {
					t.Error("failed to find find ", k)
					return
				}
			}
			n.Done()
		}()
	}

	n.Wait()
	fmt.Println("spent time ", time.Since(start))
	fmt.Println(loadCounts)
}

var lock sync.RWMutex
