package service

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultBasePath = "/_yokogcache/"
	defaultReplicas = 50
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"}, //本地etcd服务默认在2379端口监听客户端请求
		DialTimeout: 5 * time.Second,            //建立连接的超时时间
	}
)

//抽象出的接口

// 负责挑选节点，即根据一致性哈希算法找到查询的key应该访问集群中的哪个节点
type PeerPicker interface {
	Pick(key string) (Fetcher, bool)
}

// 负责查询指定缓存空间中key的值
// 集群中每个节点都必须实现该接口，当收到远程节点请求时会调用接口内方法获取缓存
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

// 要求实现从数据源获取数据能力（当本地没有缓存时的一种选择）
type Retriever interface {
	retrieve(string) ([]byte, error)
}

type RetrieverFunc func(key string) ([]byte, error)

// RetrieverFunc实现retriever方法，使得任意匿名函数func
// 通过被RetrieverFunc(func)类型强制转换后，实现了Retriever接口的能力
func (f RetrieverFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}
