package main

import (
	"flag"
	"fmt"
	"log"
	discovery "yokogcache/internal/middleware/etcd/discovery1"
	"yokogcache/internal/service"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *service.Group {
	return service.NewGroup("scores", 2<<10, service.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 分别在端口9999、10000、100001启动服务器节点组成集群
func main() {
	var port int
	flag.IntVar(&port, "port", 9999, "Yokogcache server port")
	flag.Parse()

	yoko := createGroup()
	addr := fmt.Sprintf("localhost:%d", port)
	svr, err := service.NewGRPCPool(addr)
	if err != nil {
		log.Fatalf("fail to init server at %s, err: %v", addr, err)
	}

	//添加peers
	addrs, err := discovery.GetPeers("clusters") //获取etcd集群中该前缀下的所有地址
	if err != nil {
		addrs = []string{"localhost:8001"}
	}
	svr.UpdatePeers(addrs...)
	yoko.RegisterServer(svr)
	log.Println("groupcache is running at ", addr)
	// 启动服务（注册服务至 etcd、计算一致性 hash）
	err = svr.Run()
	if err != nil {
		log.Fatal(err)
	}
}
