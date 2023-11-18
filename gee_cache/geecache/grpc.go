package geecache

import (
	"context"
	"fmt"
	"gee_cache/geecache/consistenthash"
	pb "gee_cache/geecache/geecachepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sync"
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

const (
	defaultAddr     = "8848"
	defaultReplicas = 50
)

type GrpcPool struct {
	pb.UnimplementedGroupCacheServer
	self        string
	mu          sync.Mutex
	peers       *consistenthash.Map
	grpcGetters map[string]*grpcGetter
}

type peersUpdate struct {
	added   []string
	removed []string
}

func NewGrpcPool(self string) *GrpcPool {
	return &GrpcPool{
		self:        self,
		peers:       consistenthash.New(defaultReplicas, nil),
		grpcGetters: make(map[string]*grpcGetter),
	}
}

func (p *GrpcPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *GrpcPool) add(peers ...string) {
	p.peers.Add(peers...)
	for _, peer := range peers {
		p.grpcGetters[peer] = &grpcGetter{addr: peer}
	}
}

func (p *GrpcPool) remove(peers ...string) {
	p.peers.Remove(peers...)
	for _, peer := range peers {
		delete(p.grpcGetters, peer)
	}
}

func (p *GrpcPool) UpdatePeers(update peersUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.remove(update.removed...)
	p.add(update.added...)
}

func (p *GrpcPool) SetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.add(peers...)
}

func (p *GrpcPool) PickPeer(key string) (PeerGetter, bool) {
	peer := p.peers.Get(key)
	if peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.grpcGetters[peer], true
	}
	return nil, false
}

func (p *GrpcPool) Get(_ context.Context, in *pb.Request) (*pb.Response, error) {
	p.Log("grpc get group: %s key: %s", in.Group, in.Key)
	response := &pb.Response{}
	group := GetGroup(in.Group)
	if group == nil {
		p.Log("no such group: %s", in.Group)
		return response, fmt.Errorf("no such group: %v", in.Group)
	}

	value, err := group.Get(in.Key)
	if err != nil {
		p.Log("get key %v error %v", in.Key, err)
		return response, err
	}

	response.Value = value.ByteSlice()
	return response, nil
}

func (p *GrpcPool) Run() {
	lis, err := net.Listen("tcp", p.self)
	if err != nil {
		panic(err)
	}

	creds, err := credentials.NewServerTLSFromFile("./geecache/key/test.pem", "./geecache/key/test.key")
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterGroupCacheServer(server, p)
	reflection.Register(server)
	p.Log("grpc server running in %s", p.self)
	err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
}

type grpcGetter struct {
	addr string
}

func (g *grpcGetter) Get(in *pb.Request, out *pb.Response) error {

	creds, err := credentials.NewClientTLSFromFile("./geecache/key/test.pem", "*.chtholly.com")
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(g.addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewGroupCacheClient(conn)
	response, err := client.Get(context.Background(), in)
	if response != nil {
		out.Value = response.Value
	}

	return err
}
