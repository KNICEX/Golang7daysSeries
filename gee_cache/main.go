package main

import (
	"fmt"
	"gee_cache/geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "12312",
	"Jack": "123123124",
	"Sam":  "3412412",
}

func main() {
	geecache.NewGroup("hana", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}),
	)

	addr := "localhost:8080"
	peers := geecache.NewHTTPPool(addr)
	log.Fatal(http.ListenAndServe(addr, peers))

}
