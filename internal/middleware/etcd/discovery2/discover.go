package discovery2

import (
	"context"
	"yokogcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
ClientConn 表示与概念端点的虚拟连接，用于执行 RPC，
ClientConn 可根据配置、负载等情况，与端点自由建立零个或多个实际连接。
*/
// 向grpc请求服务，返回connection
func Discovery(c *clientv3.Client, serviceName string) (*grpc.ClientConn, error) {
	etcdResolve, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	return grpc.NewClient(
		"etcd:///"+serviceName,
		grpc.WithResolvers(etcdResolve),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
}

// 根据服务名发现节点
func ListServicePeers(serviceName string) ([]string, error) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return []string{}, err
	}

	endPointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("create endpoints manager failed, %v", err)
		return []string{}, err
	}

	key2EndpointMap, err := endPointsManager.List(context.Background())
	if err != nil {
		logger.LogrusObj.Errorf("enpoint manager list op failed, %v", err)
		return []string{}, err
	}

	peers := []string{}
	for key, endpoint := range key2EndpointMap {
		peers = append(peers, endpoint.Addr)
		logger.LogrusObj.Infof("found endpoint %s (%s):(%s)", key, endpoint.Addr, endpoint.Metadata)
	}
	return peers, nil
}

// 动态监听节点变更
func DynamicServices(update chan bool, serviceName string) {
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()

	//watch机制负责订阅发布功能
	watchChan := cli.Watch(context.Background(), serviceName, clientv3.WithPrefix())

	// 每次用户往指定的服务中添加或者删除新的实例地址时，watchChan 后台都能通过 WithPrefix() 扫描到实例数量的变化并以  watchResp.Events 事件的方式返回
	// 当发生变更时，往 update channel 发送一个信号，告知 endpoint manager 重新构建哈希映射
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- true // 通知 endpoint manager 重构哈希环
				logger.LogrusObj.Warnf("Service endpoint added or updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- true // 通知 endpoint manager 重构哈希环
				logger.LogrusObj.Warnf("Service endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}
