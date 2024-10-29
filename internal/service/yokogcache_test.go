package service

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestPush(t *testing.T) {
	yoko := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	yoko.Put("key1", ByteView{[]byte("124")}, 1)
	yoko.Put("key2", ByteView{[]byte("124")}, 3)
	yoko.Put("key3", ByteView{[]byte("124")}, 3)
	yoko.Put("key4", ByteView{[]byte("124")}, 3)
	yoko.Put("key5", ByteView{[]byte("124")}, 4)

	time.Sleep(10 * time.Second)
}

func TestRetriever(t *testing.T) {
	var f Retriever = RetrieverFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.retrieve("key"); !reflect.DeepEqual(v, expect) {
		t.Fatal("callback failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db)) //统计调用回调函数的次数
	gee := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) { //回调函数直接从数据库中获取数据
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok { //当前key取不到值，说明这是第一次调用，如果出现多次调用，说明第一次调用后没有将数据存入缓存
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key) //从数据库获取失败
		}))

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	//下面再读取，就不会调用回调函数了，上面的调用已经将数据写进了gee这个实例的loaclCache字段
	tom, _ := gee.Get("Tom")
	fmt.Println("tom:", string(tom.String()))
	fmt.Println("getter counts:", loadCounts["Tom"])

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
func TestGetGroup(t *testing.T) {
	groupname := "scores"
	NewGroup(groupname, 2<<10, RetrieverFunc(
		func(key string) (bytes []byte, err error) { return },
	))
	if group := GetGroup(groupname); group == nil || group.name != groupname {
		t.Fatalf("group %s not exist", groupname)
	}
	if group := GetGroup(groupname + "11"); group != nil {
		t.Fatalf("expect nil,but %s got", group.name)
	}
}
