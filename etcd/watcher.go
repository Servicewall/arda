package etcd

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatcherEventType int

const (
	WatcherEventTypeCurrent WatcherEventType = iota
	WatcherEventTypePut
	WatcherEventTypeDelete
)

type WatcherResult struct {
	EventType WatcherEventType
	Kvs       []*mvccpb.KeyValue // only has value if EventType is WatcherEventTypeCurrent
}

type WatcherCallback func(string, WatcherResult)

type Watcher struct {
	EtcdKey  string
	Callback WatcherCallback
}

var registerWatchers []*Watcher = make([]*Watcher, 0)

func (watcher *Watcher) Register() {
	registerWatchers = append(registerWatchers, watcher)
}

func StartWatchers() {
	for _, watcher := range registerWatchers {
		go watcher.startWatch()
	}
}

func (w *Watcher) startWatch() {
	cli := GetClient()
	if cli == nil {
		fmt.Printf("[etcd] faild to watch for etcd key %s, no etcd client found.\n", w.EtcdKey)
		return
	}

	// get current value
	resp, err := clientv3.NewKV(cli).Get(context.TODO(), w.EtcdKey, clientv3.WithPrefix())
	if err == nil {
		w.Callback(w.EtcdKey, WatcherResult{
			EventType: WatcherEventTypeCurrent,
			Kvs:       resp.Kvs,
		})
	} else {
		fmt.Printf("[etcd] get current value faild for key %s.", w.EtcdKey)
	}

	watchChan := clientv3.NewWatcher(cli).Watch(context.TODO(), w.EtcdKey, clientv3.WithPrefix())

	for resp := range watchChan {
		for _, event := range resp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				w.Callback(w.EtcdKey, WatcherResult{
					EventType: WatcherEventTypePut,
					Kvs:       []*mvccpb.KeyValue{event.Kv},
				})
			case clientv3.EventTypeDelete:
				w.Callback(w.EtcdKey, WatcherResult{
					EventType: WatcherEventTypeDelete,
					Kvs:       []*mvccpb.KeyValue{event.Kv},
				})
			}
		}
	}
}
