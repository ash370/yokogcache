package discovery

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

func Register(serviceName string, addr string) error {
	//创建Etcd客户端
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}

	//为服务创建节点管理器
	em, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		return err
	}

	//创建租约
	lease, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}

	//将服务地址添加到etcd并与租约关联
	err = em.AddEndpoint(context.TODO(), serviceName+"/"+addr, endpoints.Endpoint{Addr: addr, Metadata: "YokogCache services"}, clientv3.WithLease(lease.ID))
	if err != nil {
		return nil
	}

	//启动一个 goroutine 来持续接收租约的保活信号
	alive, err := cli.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return err
	}

	log.Printf("EtcdRegistry %s/%s ", serviceName, addr)
	go func() {
		for {
			<-alive
			//fmt.Println("etcd server keep alive")
		}
	}()
	return nil
}
