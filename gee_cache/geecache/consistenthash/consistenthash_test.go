package consistenthash

import (
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	hash.Add("6", "4", "2")
	// 虚拟节点对应表
	// 6 -> 06 16 26
	// 4 -> 04 14 24
	// 2 -> 02 12 22

	testCases := map[string]string{
		"2":  "2", // 2 <= 02
		"11": "2", // 11 <= 12
		"23": "4", // 23 <= 24
		"27": "2", // 27  > 26 -> 02
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Fatalf("Asking for %s, should yielded %s", k, v)
		}
	}

	hash.Add("8")
	// 8 -> 08 18 28

	// 27 <= 28

	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Fatalf("Asking for %s, should yielded %s", k, v)
		}
	}

}
