package lru

import (
	"log"
	"testing"
)

type String string

func (s String) Size() int {
	return len(s)
}

var (
	k1, k2, k3 = "k1", "k2", "k3"
	v1, v2, v3 = "value1", "value2", "value3"
)

func TestCache(t *testing.T) {
	capacity := len(k1 + k2 + v1 + v2)
	lru := New(int64(capacity), func(s string, value Value) {
		log.Printf("%s --> %v removed\n", s, value)
	})
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatal("removing key1 failed")
	}

}
