package service

import (
	"fmt"
	"log"
	"sync"
	"yokogcache/internal/service/singleflight"
)

//在cache上再封装一层，能够为缓存命名、填充缓存等

var (
	mu     sync.RWMutex              //读写groups并发控制
	groups = make(map[string]*Group) //所有group都加入到全局的group集合里
)

// 缓存的命名空间，包括唯一的名字、回调函数以及并发缓存(cache(加锁的lru.Cache))
type Group struct {
	name       string
	localCache *cache     //本地的缓存
	retriever  Retriever  //用于从数据源获取数据
	server     NodePicker //分布式节点服务器
	flight     *singleflight.Flight
}

func NewGroup(name string, cap int64, retriever Retriever) *Group {
	//传入接口的好处是用户可以传进来回调函数也可以传结构体
	if retriever == nil {
		panic("nil retriever")
	}
	g := &Group{
		name:       name,
		localCache: &cache{capacity: cap},
		retriever:  retriever,
		flight:     &singleflight.Flight{},
	}
	//先加锁(并发写需要加锁，可以并发读)，再将当前group加入全局的groups映射里
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}

// 将 实现了 Picker 接口的节点池注入到 Group 中
func (g *Group) RegisterServer(p NodePicker) {
	if g.server != nil {
		panic("group had been registed server")
	}
	g.server = p
}

// 获取指定名字的缓存空间
func GetGroup(name string) *Group {
	//可以并发读，不能并发读写
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 从缓存空间里获取缓存值，这里封装三种获取缓存的途径
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	if v, ok := g.localCache.get(key); ok { //在本地已被缓存
		log.Println("cache hit")
		return v, nil
	}
	//cache missing,get it another way
	return g.load(key)
}

func (g *Group) load(key string) (val ByteView, err error) {
	view, err := g.flight.Fly(key, func() (interface{}, error) {
		if g.server != nil {
			if node, ok := g.server.Pick(key); ok { //选出节点
				if val, err = g.getFromNode(node, key); err == nil {
					return val, err
				}
				log.Println("[YokoGcache] Failed to get from peer", err)
			}
		}
		//没有分布式节点，从本地数据库获取数据
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// 本地向Retriever取回数据并填充缓存（用户定义Retriever）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)} //取回的原始数据是字节切片，存其深拷贝的值，防止原始数据被篡改
	g.populateCache(key, value)             //数据存入缓存
	return value, nil
}

// 提供填充缓存的能力
func (g *Group) populateCache(key string, value ByteView) {
	g.localCache.add(key, value)
}

// 从远程节点获取缓存
func (g *Group) getFromNode(node Fetcher, key string) (ByteView, error) {
	bytes, err := node.Fetch(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}