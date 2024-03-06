package etcd

import (
	"context"
	"errors"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceRegister struct {
	LeaseTTL       int64
	Renewal        time.Duration
	Key            string
	Value          string
	lease          clientv3.Lease
	leaseId        *clientv3.LeaseID
	cancelRegister context.CancelFunc
	stop           bool
}

func (sr *ServiceRegister) Start() error {
	cli := GetClient()
	if cli == nil {
		log.Printf("[etcd error] etcd client is not found")
		return errors.New("etcd client not found")
	}

	//renewalTimer := time.NewTicker(sr.Renewal)
	sr.stop = false

	// start routine to do service register
	go func() {
		for !sr.stop {
			var etcdLease clientv3.Lease
			var etcdLeaseId *clientv3.LeaseID

			if etcdLease == nil {
				etcdLease = clientv3.NewLease(cli)
				sr.lease = etcdLease
			}

			if etcdLeaseId == nil {
				ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
				leaseGrantResp, err := etcdLease.Grant(ctx, sr.LeaseTTL)
				cancel()
				if err != nil {
					log.Printf("[etcd error] grant lease failed. key: %s, error: %s\n", sr.Key, err.Error())
					time.Sleep(2 * time.Second)
					continue
				}
				etcdLeaseId = &leaseGrantResp.ID
				sr.leaseId = etcdLeaseId
				err = PutWithTimeout(2*time.Second, sr.Key, sr.Value, clientv3.WithLease(*etcdLeaseId))
				if err != nil {
					log.Printf("[etcd error] lease put kv failed. key: %s, error: %s\n", sr.Key, err.Error())
					time.Sleep(2 * time.Second)
					continue
				}
			}

			ctx, cancel := context.WithCancel(context.TODO())
			sr.cancelRegister = cancel
			respChan, err := etcdLease.KeepAlive(ctx, *etcdLeaseId)
			if err != nil {
				etcdLeaseId = nil
				log.Printf("[etcd error] lease keep alive failed. key: %s, error: %s\n", sr.Key, err.Error())
				continue
			}

			for range respChan {
			}
			etcdLeaseId = nil
			log.Printf("[etcd] cancel service register.")
		}
		log.Printf("[etcd] stop service register.")
	}()

	go func() {
		timer := time.NewTicker(30 * time.Second)
		for {
			<-timer.C
			if sr.stop {
				log.Print("[etcd] stop get register key loop. ")
				break
			}
			kvs, err := GetWithTimeout(5*time.Second, sr.Key)
			if err != nil {
				log.Printf("[etcd error] get register key failed, %s", err.Error())
			} else if len(kvs) == 0 {
				log.Print("[etcd] register key died,re-register. ")
				sr.cancelRegister()
			}
		}
	}()

	return nil
}

func (sr *ServiceRegister) Stop() (err error) {
	sr.stop = true
	sr.cancelRegister()
	err = sr.deregister()
	return
}

func (sr *ServiceRegister) deregister() (err error) {
	if cli := GetClient(); cli == nil {
		err = errors.New("etcd client not found")
		return
	}

	if sr.lease != nil && sr.leaseId != nil {
		// revoke etcd lease
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		_, err = sr.lease.Revoke(ctx, *sr.leaseId)
		cancel()
		if err != nil {
			log.Printf("[etcd error] revoke lease failed. key: %s, error: %s\n", sr.Key, err.Error())
		}
	}
	return
}
