package etcd

import (
	"context"
	"google.golang.org/grpc/connectivity"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatcherEventType int

const (
	WatcherEventTypeCurrent WatcherEventType = iota
	WatcherEventTypePut
	WatcherEventTypeDelete
)

const (
	watcherMaxGetFailedCount = 10
)

// WatcherEventTypeDelete: Kvs is the deleted kv
type WatcherResult struct {
	EventType WatcherEventType
	Kvs       []EtcdKV
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
	for i := 0; i < len(registerWatchers); i++ {
		watcher := registerWatchers[i]
		watcher.watchAsync()
	}
}

func (w *Watcher) watchAsync() {
	cli := GetClient()
	if cli == nil {
		log.Printf("[etcd error] etcd client is not found")
		return
	}

	// routine
	go func() {
		defer func() {
			log.Printf("[etcd] watcher routine end. watch key: %s", w.EtcdKey)
		}()

		// check if etcd is ready
		ready := false
		for getFailedCount := 0; getFailedCount < watcherMaxGetFailedCount; getFailedCount++ {
			if cli.ActiveConnection() != nil && cli.ActiveConnection().GetState() == connectivity.Ready {
				ready = true
				break
			}
			time.Sleep(10 * time.Second)
		}
		if !ready {
			log.Printf("[etcd error] etcd is not ready before timeout. key: %s", w.EtcdKey)
			return
		}
		watchChan := clientv3.NewWatcher(cli).Watch(context.TODO(), w.EtcdKey, clientv3.WithPrefix())

		// get current value
		currentKvs, err := GetWithTimeout(2*time.Second, w.EtcdKey, clientv3.WithPrefix())
		if err != nil {
			log.Printf("[etcd error] watcher failed to get current kv. key: %s, err: %s", w.EtcdKey, err)
			return
		}
		if w.Callback != nil {
			w.Callback(w.EtcdKey, WatcherResult{
				EventType: WatcherEventTypeCurrent,
				Kvs:       currentKvs,
			})
		}

		for resp := range watchChan {
			for _, event := range resp.Events {
				var res *WatcherResult
				switch event.Type {
				case clientv3.EventTypePut:
					res = &WatcherResult{
						EventType: WatcherEventTypePut,
						Kvs: []EtcdKV{{
							Key:   string(event.Kv.Key),
							Value: event.Kv.Value,
						}},
					}
				case clientv3.EventTypeDelete:
					res = &WatcherResult{
						EventType: WatcherEventTypeDelete,
						Kvs: []EtcdKV{{
							Key:   string(event.Kv.Key),
							Value: event.Kv.Value,
						}},
					}
				}

				if w.Callback != nil && res != nil {
					w.Callback(w.EtcdKey, *res)
				}
			}
		}
	}()
}
