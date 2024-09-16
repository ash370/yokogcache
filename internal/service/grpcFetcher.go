package service

import (
	"context"
	"log"

	pb "yokogcache/utils/yokogcachepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcFetcher struct {
	serviceName string //yokogcache/ip:port
}

var _ Fetcher = (*grpcFetcher)(nil)

func (gf *grpcFetcher) Fetch(group string, key string) ([]byte, error) {
	//建立rpc连接
	conn, err := grpc.NewClient(gf.serviceName, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	//基于rpc连接创建一个客户端，直接调用远程服务
	yokogcacheClient := pb.NewYokogCacheClient(conn)
	response, err := yokogcacheClient.Get(context.Background(), &pb.GetRequest{Group: group, Key: key})
	if err != nil {
		log.Fatalf("could not call service: %s err: %v", gf.serviceName, err)
	}
	return response.Value, err
}
