package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	WorkerWatcherEvent int
)

const (
	WorkerWatcherEventPut WorkerWatcherEvent = iota
	WorkerWatcherEventDelete
	WorkerWatcherEventTicker
)

type Watcher interface {
	WatchWorkers(func(WorkerWatcherEvent))
	WatchNewURLs(f func(*crawlerdpb.RequestURL))
	ResetTimer()
}

type etcdwatcher struct {
	etcd    *clientv3.Client
	storage storage.Client

	urlTimerTimeout time.Duration
	urlTimer        *time.Timer
}

func NewETCDWatcher(etcd *clientv3.Client, storage storage.Client, timerTimeout time.Duration) Watcher {
	return &etcdwatcher{
		etcd:            etcd,
		storage:         storage,
		urlTimerTimeout: timerTimeout,
		urlTimer:        time.NewTimer(timerTimeout),
	}
}

func (w *etcdwatcher) WatchWorkers(f func(WorkerWatcherEvent)) {
	go func() {
		tick := time.NewTicker(time.Minute)

		for {
			select {
			case <-tick.C:
				f(WorkerWatcherEventTicker)
			default:

			}
		}
	}()

	for {
		watchWorkers := <-w.etcd.Watch(context.Background(), "worker.", clientv3.WithPrefix())
		for _, ev := range watchWorkers.Events {
			switch ev.Type {
			case mvccpb.DELETE:
				f(WorkerWatcherEventDelete)
			case mvccpb.PUT:
				f(WorkerWatcherEventPut)
			}
		}
	}
}

func (w etcdwatcher) WatchNewURLs(f func(*crawlerdpb.RequestURL)) {
	justNow := time.NewTimer(time.Second)

	wg := sync.WaitGroup{}

	job := func(wait bool) {
		defer func() {
			if wait {
				wg.Done()
			}
		}()
		if wait {
			wg.Add(1)
		}

		if err := w.storage.URL().Scroll(context.Background(), func(urls []objects.URL) {
			for _, url := range urls {
				go func(url objects.URL) {
					resp, err := w.etcd.Get(context.Background(), fmt.Sprintf("crawl.%s", strconv.Itoa(url.ID)))
					if err != nil {
						log.Error(err)
						return
					}

					isCrawling := resp.Kvs != nil && len(resp.Kvs) > 0

					if isCrawling {
						return
					}

					f(&crawlerdpb.RequestURL{
						Id:       int64(url.ID),
						Url:      url.URL,
						Interval: int64(url.Interval),
					})
				}(url)
			}
		}); err != nil {
			log.Error(err)
		}
	}

	for {
		select {
		case <-w.urlTimer.C:
			job(true)
			wg.Wait()
			w.urlTimer.Reset(w.urlTimerTimeout)
		case <-justNow.C:
			job(false)
		}
	}
}

func (w *etcdwatcher) ResetTimer() {
	w.urlTimer.Reset(w.urlTimerTimeout)
}
