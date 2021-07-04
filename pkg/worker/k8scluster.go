package worker

import (
	"context"
	"net"
	"os"
	"strconv"

	"crawlerd/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type k8sCluster struct {
	client    kubernetes.Interface
	namespace string

	cfg *Config
}

func NewK8sCluster(client kubernetes.Interface, namespace string, cfg *Config) Cluster {
	return &k8sCluster{
		client:    client,
		namespace: namespace,

		cfg: cfg,
	}
}

func (k *k8sCluster) Register(ctx context.Context, w Worker) error {
	// k8s register pods itself?

	return nil
}

func (k *k8sCluster) Unregister(ctx context.Context, w Worker) error {
	// k8s unregister pods itself?

	return nil
}

func (k *k8sCluster) GetAll(ctx context.Context) ([]*WorkerMeta, error) {
	// TODO: label/namespace matcher because scheduler and api can run on the same kubernetes
	pods, err := k.client.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
	})

	if err != nil {
		return nil, err
	}

	workerMetas := make([]*WorkerMeta, 0)

	for _, pod := range pods.Items {
		workerMetas = append(workerMetas, &WorkerMeta{
			ID: pod.Name,
			// TODO: below
			/*
				TODO: addr to worker not pod
				TODO: it can be resolved by passing env but then how to deal with dynamic/multiple ports because every worker could have different ports in theory
				TODO: find best solution for this issue
			*/
			Addr: net.JoinHostPort(pod.Status.PodIP, k.cfg.WorkerGRPCAddr),
		})
	}

	return workerMetas, nil
}

func (k *k8sCluster) DeleteByID(ctx context.Context, id string) error {
	return k.client.CoreV1().Pods(k.namespace).Delete(ctx, id, metav1.DeleteOptions{})
}

func (k *k8sCluster) Watch(f func(EventType)) error {
	// TODO: label/namespace matcher because scheduler and api can run on the same kubernetes
	watchPods, err := k.client.CoreV1().Pods(k.namespace).Watch(context.Background(), metav1.ListOptions{
		TypeMeta:             metav1.TypeMeta{},
		LabelSelector:        "",
		FieldSelector:        "",
		Watch:                false,
		AllowWatchBookmarks:  false,
		ResourceVersion:      "",
		ResourceVersionMatch: "",
		TimeoutSeconds:       nil,
		Limit:                0,
		Continue:             "",
	})

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case result := <-watchPods.ResultChan():
				switch result.Type {
				// TODO: watch.Modified?
				case watch.Added:
					f(PUT)
				case watch.Deleted:
					f(DELETE)
				}
			}
		}
	}()

	return nil
}

// TODO: golang standard lib for getting pod ip should be better
func (k *k8sCluster) WorkerAddr() (id, host string, err error) {
	workerHost, err := os.Hostname()
	if err != nil {
		return "", "", err
	}

	var workerPort string

	// TODO: attach another port if already exists
	// TODO: in k8s it's should be get from named port
	if k.cfg != nil && k.cfg.WorkerGRPCAddr != "" {
		workerPort = k.cfg.WorkerGRPCAddr
	} else {
		workerPort = strconv.Itoa(util.Between(9111, 9555))
	}

	hostEnv := os.Getenv("WORKER_HOST")
	if hostEnv != "" {
		workerHost = hostEnv
	}

	//ip, err := util.ResolveHostIP()
	pod, err := k.client.CoreV1().Pods(k.namespace).Get(context.Background(), workerHost, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	workerID := workerHost
	workerAddr := net.JoinHostPort(pod.Status.PodIP, workerPort)
	//workerAddr := net.JoinHostPort(ip, workerPort)

	return workerID, workerAddr, nil
}

func (k *k8sCluster) Type() ClusterType {
	return ClusterTypeK8s
}
