package etcd

import (
	"context"
	"errors"

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
		go func() {
			_ = watcher.startSyncWatch()
		}()
	}
}

func (w *Watcher) startSyncWatch() error {
	cli := GetClient()
	if cli == nil {
		return errors.New("etcd client not found")
	}

	// get current value
	resp, err := Get(w.EtcdKey, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if w.Callback != nil {
		w.Callback(w.EtcdKey, WatcherResult{
			EventType: WatcherEventTypeCurrent,
			Kvs:       resp,
		})
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

	return nil
}
