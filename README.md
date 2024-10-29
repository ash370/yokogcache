# yoko Group Cache(yokogcache)

## 实现功能

* v1(basic) version

  * 并发访问控制singleflight
 
  * 负载均衡consistenthash一致性哈希算法
 
  * 缓存淘汰LRU
 
  * 分布式缓存节点之间基于HTTP协议通信

  * 分布式缓存节点之间基于gRPC协议通信

  * 简单的服务注册发现（手动导入服务节点）
 

* v2 version

  * 服务注册发现和动态节点管理（使用 endpoint manager 和 watch channel 实现类似于服务订阅发布的能力）
 
  * 使用类似于事件回调的处理机制，根据节点的 PUT、DEL 事件更新哈希环（无需手动导入服务节点，各节点能够感知到其他服务节点的上线下线状态，同时重构哈希环）
 
  * 增加缓存穿透的防御策略（将不存在的 key 的空值存到缓存中，设置合理过期时间，防止不存在的 key 的大量并发请求打穿数据库）

  * 注册时将每个节点看成一个服务，直接注册节点服务名到etcd，每个服务名下再放实际的地址，而一致性哈希时根据服务名做路由，各个节点之间的服务发现监听每个节点的服务名，解决一致性哈希动态路由和grpc的负载均衡不匹配问题。用户请求集群时的服务发现监听最上层服务名。

  * 增加TTL机制，基于etcd实现延迟队列，做到及时删除过期缓存

  ## 项目结构

```
.
├── cmd
│   ├── grpc                        // run grpc server
│   └── http                        // run http server
├── config
│   ├── config.go
│   └── config.yml                  // mysql, etcd, etc. configuration
├── discovery                       // service register and discovery
│   ├── discovery.go
│   └── registry.go
├── etcd                            // etcd cluster used as service registry
│   └── cluster
│       ├── Procfile
│       └── runcluster.md           // run etcd cluster guide
├── go.mod
├── go.sum
├── internal
│   ├── db
│   │   ├── cmd                     // DB client
│   │   ├── dbservice               // init DB
│   │   └── model
│   ├── middleware
│   │   ├── etcd
│   │   │   ├── cluster
│   │   │   ├── delayqueue
│   │   │   └── discovery
│   │   └── rabbitmq
│   └── service
│       ├── consistenthash          // for load-balance
│       ├── lru
│       ├── singleflight.go         // single flight for concurrent access control
│       ├── byteview.go             // for data security
│       ├── cache.go                // main cache logic          
│       ├── grpcfetcher.go          
│       ├── grpcpicker.go
│       ├── httpfetcher.go
│       ├── httphelper.go           // run http server
│       ├── httppicker.go
│       ├── interface.go
│       └── yokogcache.go
├── client.go                       // client
├── script/test
│   ├── grpc                        // grpc client
│   │   ├── grpc_test1.sh
│   │   └── buildforcheck           // run grpc client separatly
│   └── http                        // http client
│       └── http_test1.sh
└── utils
    ├── logger
    └── utils.go                    // ip address validation
```