package singleflight

import (
	"sync"
	"yokogcache/utils/logger"
)

//提供缓存击穿的保护
//当cache并发访问节点获取缓存时，如果节点未缓存该值则会向db发送大量的请求 导致db的压力骤增
//因此 将所有由key产生的请求抽象成flight
//这个flight只会起飞一次(single)可以缓解击穿的可能性
//flight载有我们要的缓存数据 称为packet

type packet struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Flight struct {
	mu     sync.Mutex
	flight map[string]*packet //key映射到请求
}

// Fly负责key航班的飞行，fn是获取packet的方法（确保其只执行一次）
func (f *Flight) Fly(key string, fn func() (interface{}, error)) (interface{}, error) {
	f.mu.Lock()
	if f.flight == nil {
		f.flight = make(map[string]*packet)
	}
	if p, ok := f.flight[key]; ok {
		//已经有其他goroutine在查询
		logger.LogrusObj.Warnf("%s 已经在查询，阻塞... 等待其他goroutine返回...", key)
		f.mu.Unlock()
		p.wg.Wait()
		return p.val, p.err
	}
	p := new(packet)
	p.wg.Add(1)
	f.flight[key] = p
	f.mu.Unlock()

	p.val, p.err = fn()
	p.wg.Done()

	f.mu.Lock()
	delete(f.flight, key) //航班已完成
	f.mu.Unlock()

	return p.val, p.err
}
