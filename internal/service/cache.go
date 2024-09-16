package service

import (
	"sync"
	"yokogcache/internal/service/lru"
)

//负责对lru模块的并发控制 =》对lru.Cache加锁

// 将cache和淘汰算法解耦，如果修改了淘汰算法，只需要在cache里修改成员即可
type cache struct {
	mu       sync.Mutex
	lru      *lru.Cache
	capacity int64 //缓存最大容量
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
