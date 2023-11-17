package geecache

import (
	"fmt"
	"gee_cache/geecache/consistenthash"
	pb "gee_cache/geecache/geecachepb"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath   = "/_geecache/"
	defaultUpdatePath = "/_geecache/_update"
	defaultReplicas   = 50
)

/// /_geecache/:groupName/:key

//								no such group
//	API -> /:group/:key -> Peer -- > 404
//							| hasGroup
//							|			hasKey
//							|--> Group --> key --> value -> return
//									|
//									|	not such key											found
//									| -> peers: hash(key) -> addr!=self -> addr/:group/:key -> value -> judgeSave -> return
//													|											| not found
//													|--> addr==self -> self.Get(key) <----------|
//																			|		found
//																			|----> value -> judgeSave -> return
//																			|		not found
//																			|----> 404
//

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.RWMutex
	peers       *consistenthash.Map    // key -> addr
	httpGetters map[string]*httpGetter // addr -> getter
}

type peersUpdate struct {
	added   []string
	removed []string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...any) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// TODO 处理update peers请求

	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

func (p *HTTPPool) SetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	p.add(peers...)
}

func (p *HTTPPool) add(peers ...string) {
	p.peers.Add(peers...)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) remove(peers ...string) {
	p.peers.Remove(peers...)
	for _, peer := range peers {
		delete(p.httpGetters, peer)
	}
}

func (p *HTTPPool) UpdatePeers(update peersUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.remove(update.removed...)
	p.add(update.added...)
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	// 确保read失败有close兜底
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}
