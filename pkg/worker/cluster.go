package worker

import (
	"context"
	"fmt"
	"strings"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// TODO: worker cluster (k8s, etcd)

type EventType int

const PrefixKeyWorker = "worker."

const (
	DELETE EventType = iota
	PUT
)

type ClusterType int

type WorkerMeta struct {
	ID   string
	Addr string
}

type Cluster interface {
	Register(ctx context.Context, w Worker) error
	Unregister(ctx context.Context, w Worker) error

	GetAll(ctx context.Context) ([]*WorkerMeta, error)

	DeleteByID(ctx context.Context, id string) error

	Watch(func(EventType))
}

type etcdCluster struct {
	etcd *clientv3.Client
}

func NewETCDCluster(etcd *clientv3.Client) Cluster {
	return &etcdCluster{
		etcd: etcd,
	}
}

func (c etcdCluster) Register(ctx context.Context, w Worker) error {
	if _, err := c.etcd.Put(ctx, c.workerID(w.ID()), w.Addr()); err != nil {
		return err
	}

	return nil
}

func (c etcdCluster) Unregister(ctx context.Context, w Worker) error {
	if _, err := c.etcd.Delete(ctx, c.workerID(w.ID())); err != nil {
		return err
	}

	return nil
}

func (c etcdCluster) GetAll(ctx context.Context) ([]*WorkerMeta, error) {
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

func (c etcdCluster) DeleteByID(ctx context.Context, id string) error {
	if _, err := c.etcd.Delete(ctx, c.workerID(id)); err != nil {
		return err
	}

	return nil
}

func (c etcdCluster) Watch(f func(EventType)) {
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
}

func (c etcdCluster) workerID(id string) string {
	return fmt.Sprintf("%s%s", PrefixKeyWorker, id)
}
