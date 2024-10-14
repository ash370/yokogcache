package delayqueue

import (
	"context"
	"time"
	"yokogcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type DelayQueue struct {
	cli *clientv3.Client
}

func NewDelayQueue() *DelayQueue {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.LogrusObj.Errorf("[Build delayqueue - new clientv3 failed, err:%s", err)
		return nil
	}
	return &DelayQueue{cli}
}

func (d *DelayQueue) Push(key string, ttl int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lease, _ := d.cli.Grant(context.Background(), ttl)
	_, err := d.cli.Put(ctx, "keyprefix/"+key, "", clientv3.WithLease(lease.ID))
	if err != nil {
		logger.LogrusObj.Errorf("[Push key(%s) into delayqueue - pushing failed, err:%s", key, err)
		return err
	}
	logger.LogrusObj.Infof("[Push key(%s) into delayqueue - pushing success", key)
	return nil
}

func DynamicKeyexpire(signal chan string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.LogrusObj.Errorf("[DynamicKeyexpire - ]failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()
	//watch机制负责订阅发布功能
	watchChan := cli.Watch(context.Background(), "keyprefix", clientv3.WithPrefix())
	logger.LogrusObj.Infoln("[DelayQueue listening...]")

	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypeDelete:
				logger.LogrusObj.Warnf("expire:(%s,%s)\n", string(ev.Kv.Key), string(ev.Kv.Value))
				signal <- string(ev.Kv.Key[10:])
			}
		}
	}
}
