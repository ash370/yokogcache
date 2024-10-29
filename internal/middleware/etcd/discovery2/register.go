package discovery2

import (
	"context"
	"fmt"
	"time"
	"yokogcache/config"
	"yokogcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

func Register(serviceName string, addr string, stop chan error) error {
	//创建Etcd客户端
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Fatalf("err: %v", err)
		return err
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
	err = em.AddEndpoint(cli.Ctx(), serviceName+"/"+addr, endpoints.Endpoint{Addr: addr, Metadata: "YokogCache services"}, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to add services as endpoint to etcd endpoint Manager: %v", err)
	}

	//设置服务心跳检测
	//自动对租约续租
	ch, err := cli.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	logger.LogrusObj.Debugf("EtcdRegistry %s ok\n", serviceName)

	for {
		select {
		case err := <-stop: //监听服务取消注册的信号
			etcdDel(cli, serviceName, addr)
			if err != nil {
				logger.LogrusObj.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): //etcd连接被断开
			return fmt.Errorf("etcd client connect broken")
		case _, ok := <-ch: //监听租约撤销信号
			//监听租约
			if !ok {
				logger.LogrusObj.Error("keepalive channel closed, revoke given lease") // 比如 etcd 断开服务，通知 server 停止
				//撤销租约
				etcdDel(cli, serviceName, addr)
				return fmt.Errorf("keepalive channel closed, revoke given lease") // 返回非 nil 的 error，上层就会关闭 stopSignalChan 从而关闭 server
			}
		default:
			time.Sleep(200 * time.Millisecond)
			//log.Printf("Recv reply from service: %s/%s, ttl:%d", serviceName, addr, lease.TTL)
		}
	}
}

func etcdDel(client *clientv3.Client, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	return endPointsManager.DeleteEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr), nil)
}
