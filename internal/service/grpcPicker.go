package service

import (
	"sync"
	"yokogcache/internal/service/consistenthash"
	pb "yokogcache/utils/yokogcachepb"
)

type GRPCPool struct {
	pb.UnimplementedYokogCacheServer

	self string //ip:port
	ring *consistenthash.ConsistentHash
	mu   sync.Mutex
}
