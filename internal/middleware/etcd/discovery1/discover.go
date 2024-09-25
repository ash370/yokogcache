package discovery1

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
ClientConn 表示与概念端点的虚拟连接，用于执行 RPC，ClientConn 可根据配置、负载等情况，与端点自由建立零个或多个实际连接。
*/
// EtcdDial向grpc请求服务，返回connection
func EtcdDial(c *clientv3.Client, serviceName string) (*grpc.ClientConn, error) {
	addr := serviceName[11:]
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// 从etcd获取peers
func GetPeers(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		log.Fatalf("failed to create etcd client, err: %v", err)
		return []string{}, err
	}

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println("get peerAddrs from etcd failed, err", err)
		return []string{}, err
	}

	peers := []string{}
	for _, kv := range resp.Kvs {
		peers = append(peers, string(kv.Value))
		//peers = append(peers, kv.String())
		//这里踩了一个小坑，需要取Value(得到的是value值即地址)，不能使用String()方法直接得到字符串类型
	}
	log.Println("get peerAddrs list from etcd success, peers: ", peers)
	return peers, nil
}
