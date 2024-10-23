package delayqueue

import (
	"fmt"
	"testing"
	"time"
)

func createDelayQueue() (*DelayQueue, <-chan string) {
	//在开启Group的时候就需要初始化延迟队列并在后台开启守护协程监听
	queue := NewDelayQueue()

	signal := make(chan string, 10)
	go DynamicKeyexpire(signal)

	return queue, signal
}

func TestPush(t *testing.T) {
	cases := []struct {
		key string
		ttl int64
	}{
		{"key1", 1},
		{"key2", 2},
		{"key3", 3},
		{"key4", 3},
		{"key5", 3},
	}

	queue, signal := createDelayQueue()
	for _, c := range cases {
		queue.Push(c.key, c.ttl)
	}

	start := time.Now()
	go func() {
		for key := range signal {
			fmt.Printf("%s expire... ttl:%d\n", key, time.Since(start).Milliseconds())
		}
	}()

	time.Sleep(10 * time.Second)
}
