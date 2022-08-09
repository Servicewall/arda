package etcd

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func Get(key string, opts ...clientv3.OpOption) (kvs []EtcdKV, err error) {
	cli := GetClient()
	if cli == nil {
		err = errors.New("etcd client not found")
		return
	}

	resp, err := clientv3.NewKV(cli).Get(context.TODO(), key, opts...)
	if err != nil {
		err = errors.New("etcd can not get value.")
		return
	}
	kvs = []EtcdKV{}
	for _, kv := range resp.Kvs {
		kvs = append(kvs, EtcdKV{
			Key:   string(kv.Key),
			Value: kv.Value,
		})
	}
	return
}

func Put(key, val string, opts ...clientv3.OpOption) error {
	cli := GetClient()
	if cli == nil {
		return errors.New("etcd client not found")
	}

	_, err := clientv3.NewKV(cli).Put(context.TODO(), key, val, opts...)
	if err != nil {
		return err
	}
	return nil
}

func PutWithTimeout(timeout time.Duration, key, val string, opts ...clientv3.OpOption) error {
	cli := GetClient()
	if cli == nil {
		return errors.New("etcd client not found")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	_, err := clientv3.NewKV(cli).Put(ctx, key, val, opts...)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

func Del(key string, opts ...clientv3.OpOption) error {
	cli := GetClient()
	if cli == nil {
		return errors.New("etcd client not found")
	}

	_, err := clientv3.NewKV(cli).Delete(context.TODO(), key, opts...)
	if err != nil {
		return err
	}
	return nil
}
