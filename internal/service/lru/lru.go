package lru

import (
	"container/list"
	"yokogcache/internal/service/persistent"
	"yokogcache/utils/logger"
)

//实现LRU最近最少使用淘汰算法，用于cache容量不够的情况下移除相应缓存记录
//这里并未实现并发机制

type OnEvicted func(string, Value) //当有记录被删除，可以调用该函数处理

type Value interface { //所有缓存的记录都是可计算长度的，需要实现该接口
	Len() int
}

// 使用LRU算法实现的缓存结构
type Cache struct {
	//添加记录时，添加的加上已使用的不能超过最大容量
	capacity int64 //Cache最大容量（byte）
	length   int64 //Cache当前已使用容量（byte）

	hashmap map[string]*list.Element
	ll      *list.List //链表头表示最近使用过

	callback OnEvicted

	snapshot *persistent.SnapShot
}

// 双向链表节点存储的对象，保存key方便查找和删除
type entry struct {
	key   string
	value Value
}

// 初始化一个缓存，指定其最大容量
// 如果maxBytes=0，表示无内存限制
func New(maxBytes int64, callback OnEvicted) *Cache {
	return &Cache{
		capacity: maxBytes,
		hashmap:  make(map[string]*list.Element),
		ll:       list.New(),
		callback: callback,
	}
}

// 向缓存添加数据
// 添加的数据必须是Value接口类型=》可统计长度的类型
func (c *Cache) Add(key string, value Value) {
	kvSize := int64(len(key)) + int64(value.Len()) //链表存储的对象既有键也有值
	//检查是否需要淘汰
	for c.capacity != 0 && c.length+kvSize > c.capacity {
		c.RemoveOldest()
	}
	if oldElem, ok := c.hashmap[key]; ok { //已存在就更新缓存
		//该key命中，移动
		c.ll.MoveToFront(oldElem)
		oldKv := oldElem.Value.(*entry)                           //将指向链表节点的指针的值断言成指向entry的指针
		c.length += int64(value.Len()) - int64(oldKv.value.Len()) //新数据和旧数据的差值是增加的，键的长度不变
		oldKv.value = value                                       //改变oldelem中存储的值(就是kv指向的值)
	} else { //不存在要新增
		elem := c.ll.PushFront(&entry{key: key, value: value}) //链表存储的是指向entry的指针
		c.hashmap[key] = elem
		c.length += kvSize
	}
}

// 访问缓存
// 给定key，返回value，同时返回是否查询成功
func (c *Cache) Get(key string) (value Value, ok bool) {
	//直接查哈希表
	if elem, ok := c.hashmap[key]; ok {
		//查到的数据插到链表头部=》最近访问
		c.ll.MoveToFront(elem)
		entry := elem.Value.(*entry)
		return entry.value, ok
	}
	return
}

func (c *Cache) RemoveOldest() {
	//链表尾是最近最少使用的
	elem := c.ll.Back()
	if elem != nil {
		c.ll.Remove(elem)
		kv := elem.Value.(*entry)
		delete(c.hashmap, kv.key) //从哈希表中删除映射
		c.length -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.callback != nil {
			//如果有回调函数
			c.callback(kv.key, kv.value)
		}
	}
}

func (c *Cache) Persist(log []string) {
	if c.snapshot == nil {
		c.snapshot = persistent.NewSnapshot("dump.spst")
	}
	err := c.snapshot.BgSave(c.hashmap)
	if err != nil {
		logger.LogrusObj.Error("持久化失败, err:", err)
	}

	//完成之后还要追加增量log
}
