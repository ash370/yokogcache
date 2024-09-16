package lru

import (
	"reflect"
	"testing"
)

// 为存储的数据类型实现Value接口
type Str string

func (s Str) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	cache := New(int64(0), nil) //初始化一个无限缓存
	cache.Add("key1", Str("1234"))
	if v, ok := cache.Get("key1"); !ok || string(v.(Str)) != "1234" { //检查是否正确命中
		t.Fatalf("cache hit key1=1234 failed!")
	}
	if _, ok := cache.Get("key2"); ok { //检查是否正确未命中
		t.Fatalf("cache miss key2 failed!")
	}
}

func TestAdd(t *testing.T) {
	cache := New(int64(0), nil)
	cache.Add("k1", Str("value1"))
	cache.Add("k1", Str("value12"))

	//检查是否正确更新缓存
	if cache.length != int64(len("k1"))+int64(len("value12")) {
		t.Fatal("expected 9 but got", cache.length)
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	cache := New(int64(cap), nil)
	cache.Add(k1, Str(v1))
	cache.Add(k2, Str(v2))
	cache.Add(k3, Str(v3))
	if _, ok := cache.Get(k1); ok {
		t.Fatalf("RemoveOldest k1 failed!")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	cache := New(int64(10), callback)
	cache.Add("key1", Str("123456"))
	cache.Add("k1", Str("v1"))
	cache.Add("k2", Str("v2"))
	cache.Add("k3", Str("v3"))

	expect := []string{"key1", "k1"}
	if !reflect.DeepEqual(keys, expect) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
