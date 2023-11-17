package geecache

import (
	"fmt"
	pb "gee_cache/geecache/geecachepb"
	"gee_cache/geecache/singleflight"
	"log"
	"math/rand"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 接口型函数，使用者可以传入函数或者实现接口的结构体
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type CacheStrategy interface {
	ShouldSave(isRemote bool, groupName string, key string) bool
}

type CacheStrategyFunc func(isRemote bool, groupName string, key string) bool

func (f CacheStrategyFunc) ShouldSave(isRemote bool, groupName string, key string) bool {
	return f(isRemote, groupName, key)
}

type Group struct {
	name          string
	getter        Getter
	cacheStrategy CacheStrategy
	mainCache     *cache
	mutex         sync.RWMutex
	peers         PeerPicker
	loader        singleflight.Loader
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 默认本地获取加缓存, 其他节点获取 1/10 概率加入本地缓存
func defaultCacheStrategy(isRemote bool, _ string, _ string) bool {
	if isRemote {
		if rand.Intn(10) > 8 {
			return true
		} else {
			return false
		}
	}
	return true
}

func newBaseGroup(name string, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	return &Group{
		name:          name,
		getter:        getter,
		cacheStrategy: CacheStrategyFunc(defaultCacheStrategy),
		loader:        singleflight.NewChanLoader(),
	}
}

// NewGroup 基于lru-1的Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	mu.Lock()
	defer mu.Unlock()
	g := newBaseGroup(name, getter)
	g.mainCache = newCache(cacheBytes)
	groups[name] = g
	return g
}

// NewKGroup 基于lru-k的Group
func NewKGroup(name string, cacheBytes int64, k int, maxHistory int, getter Getter) *Group {
	if k < 1 {
		panic("k > 0 required")
	}
	mu.Lock()
	defer mu.Unlock()
	g := newBaseGroup(name, getter)
	g.mainCache = newKCache(cacheBytes, k, maxHistory)
	groups[name] = g
	return g
}

// CacheStrategy 设置缓存策略
func (g *Group) CacheStrategy(strategy CacheStrategy) {
	if strategy == nil {
		panic("nil strategy")
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.cacheStrategy = strategy
}

// OnEvicted 设置数据被清除缓存的回调
func (g *Group) OnEvicted(fn func(key string, val ByteView)) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.mainCache.onEvicted(fn)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// Loader 设置阻止缓存击穿的Loader singleflight包提供两个实现
func (g *Group) Loader(loader singleflight.Loader) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.loader = loader
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit key:", key)
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	view, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				} else {
					log.Println("[GeeCache] Failed to get from peer ", err)
				}
			}
		}
		return g.getLocally(key)
	})

	if err != nil {
		return ByteView{}, err
	}
	return view.(ByteView), err

}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	byteView := ByteView{res.Value}
	if g.cacheStrategy.ShouldSave(true, g.name, key) {
		g.populateCache(key, byteView)
	}
	return byteView, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	res := ByteView{bytes}
	if g.cacheStrategy.ShouldSave(false, g.name, key) {
		g.populateCache(key, res)
	}
	return ByteView{bytes}, err
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
