package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"yokogcache/internal/db/dbservice"
	discovery "yokogcache/internal/middleware/etcd/discovery2"
	"yokogcache/internal/service"
	"yokogcache/utils/logger"
)

/*var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}*/

func createGroup() *service.Group {
	return service.NewGroup("scores", 2<<10, service.RetrieverFunc(
		func(key string) ([]byte, error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			db := dbservice.NewStudentDB(ctx)
			logger.LogrusObj.Infoln("[SlowDB] search key", key)
			if v, err := db.Load(key); err != nil {
				return nil, fmt.Errorf("%s not exist", key)
			} else {
				return v, nil
			}
		}))
}

// 分别在端口9999、10000、100001启动服务器节点组成集群
func main() {
	var port int
	flag.IntVar(&port, "port", 9999, "Yokogcache server port")
	flag.Parse()

	yoko := createGroup()
	addr := fmt.Sprintf("localhost:%d", port)

	updateChan := make(chan bool)
	svr, err := service.NewGRPCPool(addr, updateChan)
	if err != nil {
		logger.LogrusObj.Errorf("fail to init server at %s, err: %v", addr, err)
		return
	}

	// get a grpc service instance（通过通信来共享内存而不是通过共享内存来通信）

	//监听服务节点的变更
	go discovery.DynamicServices(updateChan, "YokogCache")

	//添加peers
	addrs, err := discovery.ListServicePeers("YokogCache") //获取etcd集群中该前缀下的所有地址
	if err != nil {
		addrs = []string{"localhost:8001"}
	}
	svr.UpdatePeers(addrs...)
	yoko.RegisterServer(svr)
	logger.LogrusObj.Infoln("groupcache is running at ", addr)
	// 启动服务（注册服务至 etcd、计算一致性 hash）
	svr.Run()
	if err != nil {
		log.Fatal(err)
	}
}
