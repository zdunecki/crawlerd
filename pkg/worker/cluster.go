package worker

import (
	"context"
)

// TODO: worker cluster (k8s, etcd)

type EventType int

const (
	DELETE EventType = iota
	PUT
)

type ClusterType string

const ClusterTypeETCD ClusterType = "etcd"
const ClusterTypeK8s ClusterType = "k8s"

type WorkerMeta struct {
	ID   string
	Addr string
}

type Cluster interface {
	Register(ctx context.Context, w Worker) error
	GetAll(ctx context.Context) ([]*WorkerMeta, error)

	DeleteByID(ctx context.Context, id string) error

	Watch(func(EventType)) error

	// TODO: better name
	WorkerAddr() (id, host string, err error)

	Type() ClusterType
}
