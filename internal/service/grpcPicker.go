package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	discovery "yokogcache/internal/middleware/etcd/discovery2"
	"yokogcache/internal/service/consistenthash"
	"yokogcache/utils"
	"yokogcache/utils/logger"
	pb "yokogcache/utils/yokogcachepb"

	"google.golang.org/grpc"
)

type GRPCPool struct {
	pb.UnimplementedYokogCacheServer

	self         string //ip:port
	ring         *consistenthash.ConsistentHash
	mu           sync.Mutex
	grpcFetchers map[string]*grpcFetcher

	status     bool       //true:running  false:stop
	stopSignal chan error //通知register revoke(撤销)服务

	update chan bool //用于传递节点更新信号
}

func NewGRPCPool(self string, update chan bool) (*GRPCPool, error) {
	//检查地址格式
	if !utils.ValidPeerAddr(self) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", self)
	}
	return &GRPCPool{self: self, update: update}, nil
}

func (gp *GRPCPool) UpdatePeers(peerAddrs ...string) {
	gp.mu.Lock()

	gp.ring = consistenthash.NewConsistentHash(defaultReplicas, nil)
	gp.ring.AddTruthNodes(peerAddrs...)
	gp.grpcFetchers = map[string]*grpcFetcher{}

	for _, peerAddr := range peerAddrs {
		if !utils.ValidPeerAddr(peerAddr) {
			gp.mu.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid addr format, it should be x.x.x.x:port", peerAddr))
		}
		// YokogCache/localhost:8001
		// YokogCache/localhost:8002
		// YokogCache/localhost:8003
		// attention：服务发现原理建议看下 Endpoint 源码, key 是 serviceName/addr value 是 addr
		// 服务解析时按照 serviceName 进行前缀查询，找到所有服务节点
		// 而 clusters 前缀是为了拿到所有实例地址做一致性哈希使用的
		// 注意 serviceName 要和你在 protocol 文件中定义的服务名称一致
		service := fmt.Sprintf("YokogCache/server%s", peerAddr[10:])
		gp.grpcFetchers[peerAddr] = &grpcFetcher{serviceName: service}
		//gp.grpcFetchers[peerAddr] = &grpcFetcher{"YokogCache"} //服务名访问
	}
	gp.mu.Unlock()

	//开启一个goroutine用于监听节点变更
	go func() {
		for {
			select {
			case <-gp.update: //接收到true说明节点数量发生变化(新增或删除)，重构哈希环
				gp.reconstruct()
			case <-gp.stopSignal:
				gp.Stop()
			default:
				time.Sleep(time.Second * 2)
			}
		}
	}()
}

func (gp *GRPCPool) reconstruct() {
	serviceList, err := discovery.ListServicePeers("YokogCache")
	if err != nil { // 如果没有拿到服务实例列表，暂时先维持当前视图
		return
	}

	gp.mu.Lock()
	gp.ring = consistenthash.NewConsistentHash(defaultReplicas, nil)
	gp.ring.AddTruthNodes(serviceList...)
	gp.grpcFetchers = map[string]*grpcFetcher{}

	for _, peerAddr := range serviceList {
		if !utils.ValidPeerAddr(peerAddr) {
			gp.mu.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid addr format, it should be x.x.x.x:port", peerAddr))
		}
		service := fmt.Sprintf("YokogCache/server%s", peerAddr[10:])
		gp.grpcFetchers[peerAddr] = &grpcFetcher{serviceName: service}
		//gp.grpcFetchers[peerAddr] = &grpcFetcher{"YokogCache"} //服务名访问
	}
	gp.mu.Unlock()
	logger.LogrusObj.Infof("hash ring reconstruct, contain service peer %v", serviceList)
}

// 实现grpc.pb.go里的服务端接口，和本地缓存逻辑交互 group.Get(key)
func (gp *GRPCPool) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	response := &pb.GetResponse{}

	groupname := in.GetGroup()
	key := in.GetKey()

	logger.LogrusObj.Infof("[yokogcache-svr %s] Recv RPC Request - (%s)/(%s)", gp.self, groupname, key)

	group := groups[groupname]
	if group == nil {
		gp.Warn("no such group %v", groupname)
		return response, fmt.Errorf("no such group %v", groupname)
	}
	value, err := group.Get(key)
	if err != nil {
		gp.Warn("get key %v error %v", key, err)
		return response, err
	}
	response.Value = value.ByteSlice()
	return response, nil
}

// 日志打印时加上服务器名称
func (gp *GRPCPool) Warn(format string, v ...interface{}) {
	logger.LogrusObj.Warnf("[Server %s] %s", gp.self, fmt.Sprintf(format, v...))
}

func (gp *GRPCPool) Pick(key string) (Fetcher, bool) {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if peerAddr := gp.ring.GetTruthNode(key); peerAddr != "" && peerAddr != gp.self {
		logger.LogrusObj.Infof("[current peer %s] Pick remote peer %s", gp.self, peerAddr)
		return gp.grpcFetchers[peerAddr], true
	} else {
		//pick itself
		gp.Warn("pick itself")
	}
	return nil, false
}

func (gp *GRPCPool) Run() {
	gp.mu.Lock()
	if gp.status {
		gp.mu.Unlock()
		fmt.Printf("yokogcache-svr %s already started", gp.self)
		return
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
	gp.stopSignal = make(chan error)

	//3.
	port := strings.Split(gp.self, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("failed to listen %s, error: %v", gp.self, err)
		return
	}

	//4.
	grpcServer := grpc.NewServer()
	pb.RegisterYokogCacheServer(grpcServer, gp) //grpcServer会调用已注册的服务YokogCache来响应请求
	defer gp.Stop()

	//5.
	go func() {
		// Register never return unless stop signal received (blocked)
		err := discovery.Register("YokogCache/"+"server"+port, gp.self, gp.stopSignal)
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		//close channel
		close(gp.stopSignal)

		err = lis.Close()
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		logger.LogrusObj.Warnf("[%s] Revoke service and close tcp socket ok.", gp.self)
	}()

	logger.LogrusObj.Infof("[%s] register service ok\n", gp.self)

	gp.mu.Unlock()

	/*
		Serve 接受监听器 lis 上的传入连接，为每个连接创建一个新的 ServerTransfer 和 service goroutine。
		service goroutines 读取 gRPC 请求，然后调用已注册的服务来给出响应。
	*/
	if err := grpcServer.Serve(lis); gp.status && err != nil {
		logger.LogrusObj.Fatalf("failed to serve %s, error: %v", gp.self, err)
		return
	}
}

func (gp *GRPCPool) Stop() {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	if !gp.status {
		return
	}

	gp.stopSignal <- nil //通知停止心跳（Register函数会返回了）
	gp.status = false
	gp.grpcFetchers = nil
	gp.ring = nil
}
