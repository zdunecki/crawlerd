package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zdunecki/crawlerd/pkg/util"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const KeyWorker = "worker"
const PrefixKeyWorker = KeyWorker + "."

const registerTTL = 60
const bumpRegisterTTL = registerTTL / 2

type etcdCluster struct {
	etcd *clientv3.Client
	once *sync.Once
}

// TODO: logger

func NewETCDCluster(etcd *clientv3.Client) Cluster {
	return &etcdCluster{
		etcd: etcd,
		once: &sync.Once{},
	}
}

func (c *etcdCluster) Register(ctx context.Context, w Worker) error {
	c.once.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(bumpRegisterTTL))

			for {
				select {
				case <-ticker.C:
					workerResp, err := c.etcd.Get(context.Background(), c.workerID(w.ID()))
					if err != nil {
						continue
					}
					if workerResp == nil || len(workerResp.Kvs) == 0 {
						continue
					}

					c.etcd.KeepAlive(context.Background(), clientv3.LeaseID(workerResp.Kvs[0].Lease))
				}
			}
		}()
	})

	lease, err := c.etcd.Grant(context.TODO(), registerTTL)
	if err != nil {
		return err
	}

	if _, err := c.etcd.Put(ctx, c.workerID(w.ID()), w.Addr(), clientv3.WithLease(lease.ID)); err != nil {
		return err
	}

	return nil
}

func (c *etcdCluster) GetAll(ctx context.Context) ([]*WorkerMeta, error) {
	resp, err := c.etcd.Get(ctx, c.workerID(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	meta := make([]*WorkerMeta, 0)
	if resp.Kvs != nil {
		for _, kv := range resp.Kvs {
			keyAtoms := strings.Split(string(kv.Key), ".")
			meta = append(meta, &WorkerMeta{
				ID:   keyAtoms[1],
				Addr: string(kv.Value),
			})
		}
	}

	return meta, nil
}

func (c *etcdCluster) DeleteByID(ctx context.Context, id string) error {
	if _, err := c.etcd.Delete(ctx, c.workerID(id)); err != nil {
		return err
	}

	return nil
}

func (c *etcdCluster) Watch(f func(EventType)) error {
	go func() {
		for {
			watchWorkers := <-c.etcd.Watch(context.Background(), PrefixKeyWorker, clientv3.WithPrefix())
			for _, ev := range watchWorkers.Events {
				switch ev.Type {
				case mvccpb.DELETE:
					f(DELETE)
				case mvccpb.PUT:
					f(PUT)
				}
			}
		}
	}()

	return nil
}

func (c *etcdCluster) WorkerAddr() (id, host string, err error) {
	workerID := util.RandomString(10)
	// TODO: attach another port if already exists
	workerPort := strconv.Itoa(util.Between(9111, 9555))
	workerHost, err := os.Hostname()
	hostEnv := os.Getenv("WORKER_HOST")
	if hostEnv != "" {
		workerHost = hostEnv
	} else {
		if err != nil {
			return "", "", err
		}
	}

	workerAddr := net.JoinHostPort(workerHost, workerPort)

	return workerID, workerAddr, nil
}

func (c *etcdCluster) Type() ClusterType {
	return ClusterTypeETCD
}

func (c *etcdCluster) workerID(id string) string {
	return fmt.Sprintf("%s%s", PrefixKeyWorker, id)
}
