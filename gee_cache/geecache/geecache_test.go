package geecache

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

var db = map[string]string{
	"Tom":  "12312",
	"Jack": "123123124",
	"Sam":  "3412412",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))

	var mutex sync.RWMutex

	gee := NewGroup("score", 1024, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key: ", key)
			mutex.Lock()
			loadCounts[key]++
			mutex.Unlock()
			time.Sleep(time.Millisecond * 20)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return make([]byte, 0), fmt.Errorf("not found")
		},
	))
	start := time.Now()
	var n sync.WaitGroup
	for i := 3000; i > 0; i-- {
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

func TestSync(t *testing.T) {
	lock.RLock()
	lock.Lock()
	lock.Unlock()
	lock.RUnlock()

}
