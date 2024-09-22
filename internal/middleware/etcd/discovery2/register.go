package discovery2

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
		DialTimeout: 10 * time.Second,
	}
)

func Register(serviceName string, addr string, stop chan error) error {
	//创建Etcd客户端
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}
	defer cli.Close()

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
	err = em.AddEndpoint(cli.Ctx(), serviceName+"/"+addr, endpoints.Endpoint{Addr: addr, Metadata: "YokogCache services"}, clientv3.WithLease(lease.ID))
	if err != nil {
		log.Fatal(err)
	}

	//设置服务心跳检测
	//自动对租约续租
	ch, err := cli.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	log.Printf("EtcdRegistry %s/%s ok\n", serviceName, addr)
	for {
		select {
		case err := <-stop: //监听服务取消注册的信号
			if err != nil {
				log.Fatal(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): //监听服务被取消的信号
			log.Println("service closed")
		case _, ok := <-ch: //监听租约撤销信号
			//监听租约
			if !ok {
				log.Println("keepalive channel closed")
				//撤销租约
				_, err := cli.Revoke(context.Background(), lease.ID)
				return err
			}
			//log.Printf("Recv reply from service: %s/%s, ttl:%d", serviceName, addr, lease.TTL)
		}
	}
}
