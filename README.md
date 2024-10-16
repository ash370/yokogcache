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
 
  * 使用类似于事件回调的处理机制，根据节点的 PUT、DEL 事件更新节点状态（无需手动导入服务节点，各节点能够感知到其他服务节点的上线）
 
  * failover 容错机制，节点失效后请求将转发到其他节点处理；即使所有节点下线，只要其中一个节点完成重启仍可继续提供服务
 
  * 增加缓存穿透的防御策略（将不存在的 key 的空值存到缓存中，设置合理过期时间，防止不存在的 key 的大量并发请求打穿数据库）

  * grpc会根据其负载均衡算法从在线的服务节点列表中选出一个目标节点发起请求，很可能会将请求再次打回当前节点，由于当前节点设置了singleflight，所以收到的第二个请求被认为是重复请求而阻塞，然而第一个请求由于rpc失败会出现fail to call peer错误。
  * 注册时使用serviceName/addr作为服务名，动态管理节点时监听serviceName，因为etcd是根据前缀查找的，所以查找serviceName可以找到所有服务节点，当前节点调用peer时使用serviceName/addr就可以直接将请求发送到一致性哈希算法选出的节点