package service

import (
	"fmt"
	"sync"
	"yokogcache/internal/service/lru"
	"yokogcache/utils/logger"
)

var limit = 15

//负责对lru模块的并发控制 =》对lru.Cache加锁

// 将cache和淘汰算法解耦，如果修改了淘汰算法，只需要在cache里修改成员即可
type cache struct {
	mu       sync.Mutex
	lru      *lru.Cache
	capacity int64 //缓存最大容量

	cnt        int  //put操作的计数器
	isSnapshot bool //标记是否在快照期间
	log        []string
}

func newCache(cap int64, signal <-chan string) *cache {
	c := &cache{
		capacity: cap,
		log:      []string{},
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

	if c.isSnapshot {
		//snapshot期间，先存入切片缓存
		c.log = append(c.log, "add"+key+value.String())
	}

	c.lru.Add(key, value)

	c.cnt++
	if c.cnt == limit {
		logger.LogrusObj.Info("写入次数达到阈值，触发后台快照...")
		c.isSnapshot = true
		go c.lru.Persist(c.log)
	}

	if !c.isSnapshot {
		//不在snapshot期间，写log文件
	}
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
