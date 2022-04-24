package etcd

import (
	"context"
	"errors"
	"log"
	"time"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceRegister struct {
	LeaseTTL    int64
	Renewal     time.Duration
	Key         string
	Value       string
	etcdLeaseId *clientv3.LeaseID
	etcdLease   clientv3.Lease
	stopChan    chan bool
}

func (sr *ServiceRegister) Start() error {
	cli := GetClient()
	if cli == nil {
		return errors.New("etcd client not found")
	}

	lease := clientv3.NewLease(cli)
	leaseGrantResp, err := lease.Grant(context.TODO(), sr.LeaseTTL)
	if err != nil {
		return err
	}
	leaseID := leaseGrantResp.ID
	err = Put(sr.Key, sr.Value, clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}

	// try to stop previous routine
	sr.Stop()

	// start a new routine to do service register
	stopChan := make(chan bool, 2)
	renewalTimer := time.NewTicker(sr.Renewal)

	go func() {
		loopFlag := true
		for {
			select {
			case <-renewalTimer.C:
				_, err := lease.KeepAliveOnce(context.TODO(), leaseID)
				if err == rpctypes.ErrLeaseNotFound {
					// 正常 renewal 时, etcd lease 未找到
					// 停止当前 routine， 启动新的 routine
					// 比如当使用断点调试导致 etcd lease ttl 触发后删除了 lease 的情况
					log.Printf("[etcd] etcd lease id [%d] is not found. start a new one\n", leaseID)
					loopFlag = false
					_ = sr.Start()
				}
			case <-stopChan:
				loopFlag = false
			}

			if !loopFlag {
				break
			}
		}
	}()

	sr.etcdLeaseId = &leaseID
	sr.etcdLease = lease
	sr.stopChan = stopChan

	return nil
}

func (sr *ServiceRegister) Stop() (err error) {
	cli := GetClient()
	if cli == nil {
		err = errors.New("etcd client not found")
		return
	}

	if sr.stopChan != nil {
		// stop previous registery routine
		sr.stopChan <- true
		close(sr.stopChan)
		sr.stopChan = nil
	}

	// revoke etcd lease
	if sr.etcdLease != nil && sr.etcdLeaseId != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		_, err = sr.etcdLease.Revoke(ctx, *sr.etcdLeaseId)
		cancel()
		if err != nil {
			return
		}
	}
	return
}
