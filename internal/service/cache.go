package service

import (
	"fmt"
	"sync"
	"yokogcache/internal/service/lru"
	"yokogcache/utils/logger"
)

//负责对lru模块的并发控制 =》对lru.Cache加锁

// 将cache和淘汰算法解耦，如果修改了淘汰算法，只需要在cache里修改成员即可
type cache struct {
	mu       sync.Mutex
	lru      *lru.Cache
	capacity int64 //缓存最大容量
}

func newCache(cap int64, signal <-chan string) *cache {
	c := &cache{
		capacity: cap,
	}
	//go func(){
	//	listening...
	//	c.delete(key)
	//}
	go func() {
		for key := range signal {
			logger.LogrusObj.Infof("Cache execute deleting key %s ...", key)
			c.delete(key)
			logger.LogrusObj.Infof("key %s deleted", key)
		}
	}()
	return c
}

// 并发读写，封装add方法和get方法
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.capacity, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}

func (c *cache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	fmt.Printf("key %s expire, deleting...\n", key)
}
