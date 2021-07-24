package worker

import (
	"context"
	"net"
	"os"
	"strconv"

	"crawlerd/pkg/util"
	k8sv1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type k8sCluster struct {
	client         kubernetes.Interface
	namespace      string
	workerSelector string

	cfg *Config
}

func NewK8sCluster(client kubernetes.Interface, namespace, workerSelector string, cfg *Config) Cluster {
	return &k8sCluster{
		client:         client,
		namespace:      namespace,
		workerSelector: workerSelector,

		cfg: cfg,
	}
}

func (k *k8sCluster) Register(ctx context.Context, w Worker) error {
	// k8s register pods itself?

	return nil
}

func (k *k8sCluster) GetAll(ctx context.Context) ([]*WorkerMeta, error) {
	labelSelector := ""

	if k.workerSelector != "" {
		selector, err := k8slabels.Parse(k.workerSelector)
		if err != nil {
			return nil, err
		}

		if selector != nil && selector.String() != "" {
			labelSelector = selector.String()
		}
	}

	// TODO: label/namespace matcher because scheduler and api can run on the same kubernetes namespace
	pods, err := k.client.CoreV1().Pods(k.namespace).List(ctx, k8smetav1.ListOptions{
		TypeMeta:             k8smetav1.TypeMeta{},
		LabelSelector:        labelSelector,
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

	// TODO: wait for terminating status

	for _, pod := range pods.Items {
		// TODO: wait for pending pods

		if pod.Status.Phase != k8sv1.PodRunning || pod.DeletionTimestamp != nil {
			continue
		}

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
	// k8s delete pods itself?
	return nil
}

func (k *k8sCluster) Watch(f func(EventType)) error {
	// TODO: label/namespace matcher because scheduler and api can run on the same kubernetes namespace
	watchPods, err := k.client.CoreV1().Pods(k.namespace).Watch(context.Background(), k8smetav1.ListOptions{
		TypeMeta:             k8smetav1.TypeMeta{},
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
	pod, err := k.client.CoreV1().Pods(k.namespace).Get(context.Background(), workerHost, k8smetav1.GetOptions{})
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
