package etcd

import (
	"context"
	"log"
	"sort"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/connectivity"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatcherEventType int

const (
	WatcherEventTypeCurrent WatcherEventType = iota
	WatcherEventTypePut
	WatcherEventTypeDelete
)

// WatcherEventTypeDelete: Kvs is the deleted kv
type WatcherResult struct {
	EventType WatcherEventType
	Kvs       []EtcdKV
}

type WatcherCallback func(string, WatcherResult)

type Watcher struct {
	Scale           int
	EtcdKey         string
	Callback        WatcherCallback
	isStarted       atomic.Bool
	isCurrentCalled atomic.Bool
}

type watchers []*Watcher

var registerWatchers watchers = make([]*Watcher, 0)

func (watcher *Watcher) Register() {
	registerWatchers = append(registerWatchers, watcher)
}

func (w watchers) Len() int {
	return len(w)
}

func (w watchers) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func (w watchers) Less(i, j int) bool {
	return w[i].Scale > w[j].Scale
}

func StartWatchers() {
	sort.Sort(registerWatchers)
	for i := 0; i < len(registerWatchers); i++ {
		watcher := registerWatchers[i]
		go watcher.start()
	}
}

func (w *Watcher) start() {
	if w.isStarted.Load() {
		log.Printf("[arda etcd error] watcher has been already started. key: %s", w.EtcdKey)
		return
	}
	w.isStarted.Store(true)

	defer func() {
		log.Printf("[arda etcd error] watcher routine end. watch key: %s", w.EtcdKey)
	}()

	for {
		cli := GetClient()
		if cli == nil {
			log.Printf("[arda etcd error] etcd client is not found. retry after 20 seconds")
			time.Sleep(20 * time.Second)
			continue
		}

		// etcd client is not ready
		if cli.ActiveConnection() == nil || cli.ActiveConnection().GetState() != connectivity.Ready {
			log.Printf("[arda etcd error] etcd is not ready, retry later. key: %s", w.EtcdKey)
			time.Sleep(20 * time.Second)
			continue
		}

		if !w.isCurrentCalled.Load() {
			// get current value
			currentKvs, err := GetWithTimeout(2*time.Second, w.EtcdKey, clientv3.WithPrefix())
			if err != nil {
				log.Printf("[arda etcd error] watcher failed to get current kv. key: %s, err: %s", w.EtcdKey, err)
				time.Sleep(20 * time.Second)
				continue
			}
			if w.Callback != nil {
				w.Callback(w.EtcdKey, WatcherResult{
					EventType: WatcherEventTypeCurrent,
					Kvs:       currentKvs,
				})
			}
			w.isCurrentCalled.Store(true)
		}

		watchChan := clientv3.NewWatcher(cli).Watch(context.TODO(), w.EtcdKey, clientv3.WithPrefix())
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
	}
}
