package main

import (
	"context"
	"log"

	"yokogcache/config"
	discovery "yokogcache/internal/middleware/etcd/discovery2"
	"yokogcache/utils/logger"
	pb "yokogcache/utils/yokogcachepb"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	config.InitConfig()

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	// 服务发现（直接根据服务名字获取与服务的虚拟端连接）
	conn, err := discovery.Discovery(cli, "YokogCache")

	if err != nil {
		panic(err)
	}
	logger.LogrusObj.Debug("Discovery continue")

	client_stub := pb.NewYokogCacheClient(conn)

	response, err := client_stub.Get(context.TODO(), &pb.GetRequest{Key: "Ella Robinson", Group: "scores"})
	if err != nil {
		log.Fatalln("没有查询到这个人的记录", err.Error())
		return
	}
	logger.LogrusObj.Infof("成功从 RPC 返回调用结果：%s\n", string(response.GetValue()))
}
