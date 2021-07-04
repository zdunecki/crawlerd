package scheduler

import (
	"context"
	"sync"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"crawlerd/pkg/worker"
	log "github.com/sirupsen/logrus"
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

type watcher struct {
	workerCluster worker.Cluster
	url           storage.URLRepository
	registry      storage.RegistryRepository

	urlTimerTimeout time.Duration
	urlTimer        *time.Timer
}

func NewWatcher(workerCluster worker.Cluster, url storage.URLRepository, timerTimeout time.Duration) Watcher {
	return &watcher{
		workerCluster:   workerCluster,
		url:             url,
		urlTimerTimeout: timerTimeout,
		urlTimer:        time.NewTimer(timerTimeout),
	}
}

func (w *watcher) WatchWorkers(f func(WorkerWatcherEvent)) {
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

	if err := w.workerCluster.Watch(func(event worker.EventType) {
		switch event {
		case worker.DELETE:
			f(WorkerWatcherEventDelete)
		case worker.PUT:
			f(WorkerWatcherEventPut)
		}
	}); err != nil {
		// TODO: log, backoff
	}
}

// TODO: tests
func (w watcher) WatchNewURLs(f func(*crawlerdpb.RequestURL)) {
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

		if err := w.url.Scroll(context.Background(), func(urls []objects.URL) {
			for _, url := range urls {
				go func(url objects.URL) {
					resp, err := w.registry.GetURLByID(context.Background(), url.ID)

					if err != nil {
						log.Error(err)
						return
					}

					isCrawling := resp != nil

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

func (w *watcher) ResetTimer() {
	w.urlTimer.Reset(w.urlTimerTimeout)
}
