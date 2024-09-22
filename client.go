package main

import (
	"context"
	"log"
	"time"

	pb "yokogcache/utils/yokogcachepb"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	/*resp, err := cli.Get(ctx, "clusters", clientv3.WithPrefix())
	if err != nil {
		log.Fatalln("从 etcd 获取节点地址失败")
		return
	}

	addr := string(resp.Kvs[0].Value)
	log.Printf("获取地址: %s", addr)*/

	//conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("etcd:///YokogCache", grpc.WithResolvers(etcdResolver), grpc.WithTransportCredentials(insecure.NewCredentials()) /*, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`)*/)
	//"etcd:///YokogCache/localhost:9999"无法访问，=》要么用服务名访问，要么直接用地址
	//使用etcd后，每次的查询请求可能会发给不同的节点

	if err != nil {
		log.Fatalln("获取 grpc 通道失败")
		return
	}
	log.Println("从 etcd 获取 grpc 通道成功")

	client_stub := pb.NewYokogCacheClient(conn)

	response, err := client_stub.Get(ctx, &pb.GetRequest{Key: "Jack", Group: "scores"})
	if err != nil {
		log.Fatalln("没有查询到这个人的记录", err.Error())
		return
	}
	log.Printf("成功从 RPC 返回调用结果：%s\n", string(response.GetValue()))
}
