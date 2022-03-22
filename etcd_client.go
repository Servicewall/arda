package arda

import (
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClientConfig struct {
	clientv3.Config
}

func (cfg *EtcdClientConfig) NewEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(cfg.Config)
}

func NewEtcdClient(endpoints string) error {
	cfg := EtcdClientConfig{
		clientv3.Config{
			Endpoints:   strings.Split(endpoints, ","),
			DialTimeout: 5 * time.Second,
		},
	}
	_, err := cfg.NewEtcdClient()
	return err
}
