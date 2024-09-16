package main

import (
	"flag"
	"fmt"
	"log"
	"yokogcache/internal/service"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *service.Group {
	return service.NewGroup("scores", 2<<10, service.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Yokogcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	yoko := createGroup()
	if api {
		go service.StartHTTPAPIServer(apiAddr, yoko)
	}
	service.StartHTTPCacheServer(addrMap[port], addrs, yoko)
}
