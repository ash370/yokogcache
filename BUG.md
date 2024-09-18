# 待解决问题

在`client.go`客户端文件中发起rpc调用使用
```go
conn, err := grpc.NewClient("etcd:///YokogCache", grpc.WithResolvers(etcdResolver), grpc.WithTransportCredentials(insecure.NewCredentials()))
```
直接用服务名访问没有问题，但是在`discover.go`中这样调用就会出现`watch close`的错误...

`discover.go`是当本地没有缓存 向peer节点发出请求时进行服务发现，可以建立grpc连接，但是Get失败

改成
```go
addr := serviceName[11:] //todo: 为什么不取出地址无法访问...
	return grpc.NewClient(
		//"etcd:///"+serviceName,
		addr,
		grpc.WithResolvers(etcdResolve),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
```
请求成功...

修改：不是Get失败。是报warn，虽然不影响获取数据，但是不明白为什么会报warn