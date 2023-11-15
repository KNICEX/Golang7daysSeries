package lru

import "testing"

func TestLRUK(t *testing.T) {
	lruk := NewKCache(1024, 2, 10, nil)
	lruk.Add(k1, String(v1))
	lruk.Add(k2, String(v2))
	lruk.Add(k3, String(v3))
	if _, ok := lruk.Get(k1); ok {
		t.Fatal("k is not valid")
	}
	if _, ok := lruk.Get(k2); ok {
		t.Fatal("k is not valid")
	}
	lruk.Add(k2, String(v2))
	if _, ok := lruk.Get(k2); !ok {
		t.Fatal()
	}
	if lruk.HistoryLen() != 2 {
		t.Fatal()
	}
}
