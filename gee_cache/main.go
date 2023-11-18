package main

import (
	"flag"
	"fmt"
	"gee_cache/geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":      "12312",
	"Jack":     "123123124",
	"Sam":      "3412412",
	"Chtholly": "123123123",
	"Willem":   "123123123",
	"Almaria":  "123123123",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("hana", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}),
	)
}
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	})
	log.Println("frontend server is running at ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr, nil))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewGrpcPool(addr)
	peers.SetPeers(addrs...)
	gee.RegisterPeers(peers)
	peers.Run()
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start an api server?")
	flag.Parse()

	apiAddr := "localhost:9999"
	addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)

}
