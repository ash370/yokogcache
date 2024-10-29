package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"yokogcache/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	//初始化
	cli, err := clientv3.New(config.DefaultEtcdConfig)

	if err != nil {
		fmt.Println("new clientv3 failed,err:", err)
		return
	}

	fmt.Println("connect to etcd success!")
	defer cli.Close()

	//put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = cli.Put(ctx, "clusters/localhost:8003", "localhost:8003")
	if err != nil {
		log.Fatal("put groupcache service to etcd failed")
		return
	}

	fmt.Println("put groupcache service to etcd success!")
}
