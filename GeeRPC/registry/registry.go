package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type GeeRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultGeeRegistry = New(defaultTimeout)

func (r *GeeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.servers[addr]
	if !ok {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now()
	}
}

func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

const GeerpcHeader = "X-Geerpc-Servers"

func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set(GeerpcHeader, strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		addr := req.Header.Get(GeerpcHeader)
		if addr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *GeeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultGeeRegistry.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()

}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := http.Client{}
	req, _ := http.NewRequest(http.MethodPost, registry, nil)
	req.Header.Set(GeerpcHeader, addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat error: ", err)
		return err
	}
	return nil
}
