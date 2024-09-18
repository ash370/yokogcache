package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	discovery "yokogcache/internal/middleware/etcd/discovery1"
	"yokogcache/internal/service/consistenthash"
	"yokogcache/utils"
	pb "yokogcache/utils/yokogcachepb"

	"google.golang.org/grpc"
)

type GRPCPool struct {
	pb.UnimplementedYokogCacheServer

	self         string //ip:port
	ring         *consistenthash.ConsistentHash
	mu           sync.Mutex
	grpcFetchers map[string]*grpcFetcher

	status bool //true:running  false:stop
	//stopSignal chan error //通知register revoke(撤销)服务
}

func NewGRPCPool(self string) (*GRPCPool, error) {
	//检查地址格式
	if !utils.ValidPeerAddr(self) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", self)
	}
	return &GRPCPool{self: self}, nil
}

func (gp *GRPCPool) UpdatePeers(peerAddrs ...string) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	gp.ring = consistenthash.NewConsistentHash(defaultReplicas, nil)
	gp.ring.AddTruthNodes(peerAddrs...)
	gp.grpcFetchers = map[string]*grpcFetcher{}

	// 注意：此操作是覆写操作，peersIP 必须满足 x.x.x.x:port 的格式
	for _, peerAddrs := range peerAddrs {
		if !utils.ValidPeerAddr(peerAddrs) {
			panic(fmt.Sprintf("[peer %s] invalid addr format, it should be x.x.x.x:port", peerAddrs))
		}
		// YokogCache/localhost:9999
		// YokogCache/localhost:10000
		// YokogCache/localhost:10001
		// attention：服务发现原理建议看下 Endpoint 源码, key 是 serviceName/addr value 是 addr
		// 服务解析时按照 serviceName 进行前缀查询，找到所有服务节点
		// 而 clusters 前缀是为了拿到所有实例地址做一致性哈希使用的
		// 注意 serviceName 要和你在 protocol 文件中定义的服务名称一致
		service := fmt.Sprintf("YokogCache/%s", peerAddrs) //这个前缀后面是所有服务节点地址
		gp.grpcFetchers[peerAddrs] = &grpcFetcher{serviceName: service}
	}
}

// 实现grpc.pb.go里的服务端接口，和本地缓存逻辑交互 group.Get(key)
func (gp *GRPCPool) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	response := &pb.GetResponse{}

	groupname := in.Group
	key := in.Key

	log.Printf("[yokogcache-svr %s] Recv RPC Request - (%s)/(%s)", gp.self, groupname, key)

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

	if peerAddr := gp.ring.GetTruthNode(key); peerAddr != "" && peerAddr != gp.self {
		gp.Log("[current peer %s] Pick remote peer %s", gp.self, peerAddr)
		return gp.grpcFetchers[peerAddr], true
	}
	return nil, false
}

func (gp *GRPCPool) Run() error {
	gp.mu.Lock()
	if gp.status {
		gp.mu.Unlock()
		return fmt.Errorf("yokogcache-svr %s already started", gp.self)
	}

	/*
		-----------------启动服务----------------------
		 1. 设置status=true 表示服务器已在运行
		 2. 初始化stop channal,后续用于通知register stop keep alive
		 3. 初始化tcp socket并开始监听gp.self开启的端口
		 4. 注册当前自定义服务至grpc 这样grpc收到request可以分发给节点处理
		 5. 将自己的服务名/Host地址注册至etcd 这样client可以通过etcd
		    获取服务Host地址 从而进行通信。这样的好处是client只需知道服务名
		    以及etcd的Host即可获取对应服务IP 无需写死至client代码中
		 ----------------------------------------------
	*/

	//1. 2.
	gp.status = true
	//gp.stopSignal = make(chan error)

	//3.
	port := strings.Split(gp.self, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen %s, error: %v", gp.self, err)
	}

	//4.
	grpcServer := grpc.NewServer()
	pb.RegisterYokogCacheServer(grpcServer, gp) //grpcServer会调用已注册的服务YokogCache来响应请求

	//5.
	go func() {
		// Register never return unless stop signal received (blocked)
		err = discovery.Register("YokogCache", gp.self)
		if err != nil {
			log.Fatal(err)
		}
	}()

	gp.mu.Unlock()

	/*
		Serve 接受监听器 lis 上的传入连接，为每个连接创建一个新的 ServerTransfer 和 service goroutine。
		service goroutines 读取 gRPC 请求，然后调用已注册的服务来给出响应。
	*/
	if err := grpcServer.Serve(lis); gp.status && err != nil {
		return fmt.Errorf("grpcServer failed to serve %s, error: %v", gp.self, err)
	}
	return nil
}
