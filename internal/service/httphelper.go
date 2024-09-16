package service

import (
	"log"
	"net/http"
)

// 启动缓存服务器：创建HTTPPool，添加节点信息，注册到Group里
func StartHTTPCacheServer(addr string, addrs []string, yokogcache *Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	yokogcache.RegisterServer(peers)
	log.Println("yokogcache is running at", addr)   //addr是本地服务端
	log.Fatal(http.ListenAndServe(addr[7:], peers)) //监听远程节点是否有请求
}

// 启动一个 API 服务（端口 9999），与用户进行交互
// 当用户通过 /api 路径发送请求时，服务器（分布式的）会从缓存组中获取数据并返回
//
// todo：gin路由拆分请求负载
func StartHTTPAPIServer(apiAddr string, yokogcache *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := yokogcache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
