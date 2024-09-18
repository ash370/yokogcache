package service

import (
	"context"
	"fmt"
	"log"
	"time"
	discovery "yokogcache/internal/middleware/etcd/discovery1"

	pb "yokogcache/utils/yokogcachepb"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type grpcFetcher struct {
	serviceName string //服务名称 YokogCache
}

var _ Fetcher = (*grpcFetcher)(nil)

func (gf *grpcFetcher) Fetch(group string, key string) ([]byte, error) {
	//创建一个etcd客户端（这里创建在etcd服务端的对外服务端口2379）
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	//发现服务，etcd客户端向grpc请求服务，返回和目标服务器的连接
	conn, err := discovery.EtcdDial(cli, gf.serviceName)
	if err != nil {
		return nil, err
	}
	log.Printf("connect to %s", gf.serviceName)
	//基于grpc连接创建一个peer对应的客户端，直接调用peer的服务
	yokogcacheClient := pb.NewYokogCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//使用带有超时自动取消的上下文 和 指定 请求调用 客户端的 Get 方法发起 rpc 请求调用
	resp, err := yokogcacheClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not call service: %s err: %v", gf.serviceName, err)
	}
	return resp.Value, nil
}
