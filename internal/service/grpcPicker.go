package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"yokogcache/internal/service/consistenthash"
	pb "yokogcache/utils/yokogcachepb"

	"google.golang.org/grpc"
)

var _ NodePicker = (*GRPCPool)(nil)

type GRPCPool struct {
	pb.UnimplementedYokogCacheServer

	self         string //ip:port
	ring         *consistenthash.ConsistentHash
	mu           sync.Mutex
	grpcFetchers map[string]*grpcFetcher
}

func NewGRPCPool(self string) *GRPCPool {
	return &GRPCPool{self: self}
}

func (gp *GRPCPool) UpdatePeers(peers ...string) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.ring = consistenthash.NewConsistentHash(defaultReplicas, nil)
	gp.ring.AddTruthNodes(peers...)
	gp.grpcFetchers = map[string]*grpcFetcher{}

	for _, peer := range peers {
		gp.grpcFetchers[peer] = &grpcFetcher{serviceName: peer}
	}
}

// 实现grpc.pb.go里的服务端接口，处理请求
func (gp *GRPCPool) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	gp.Log("%s %s", in.Group, in.Key)
	response := &pb.GetResponse{}

	groupname := in.Group
	key := in.Key

	group := groups[groupname]
	if group == nil {
		gp.Log("no such group %v", groupname)
		return response, fmt.Errorf("no such group %v", groupname)
	}
	value, err := group.Get(key)
	if err != nil {
		gp.Log("get key %v error %v", key, err)
		return response, err
	}
	response.Value = value.ByteSlice()
	return response, nil
}

// 日志打印时加上服务器名称
func (gp *GRPCPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", gp.self, fmt.Sprintf(format, v...))
}

func (gp *GRPCPool) Pick(key string) (Fetcher, bool) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if peer := gp.ring.GetTruthNode(key); peer != "" && peer != gp.self {
		gp.Log("Pick peer %s", peer)
		return gp.grpcFetchers[peer], true
	}
	return nil, false
}

func (gp *GRPCPool) Run() {
	lis, err := net.Listen("tcp", gp.self)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	pb.RegisterYokogCacheServer(server, gp)

	err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
}
