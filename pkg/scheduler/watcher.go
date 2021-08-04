package scheduler

import (
	"context"
	"sync"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/meta/v1"
	"crawlerd/pkg/storage"
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
	mu            *sync.RWMutex
	jobDoneC      chan bool
	isJobRunning  bool
	retryJob      bool

	urlTimerTimeout time.Duration
	urlTimer        *time.Timer

	log *log.Entry
}

func NewWatcher(workerCluster worker.Cluster, url storage.URLRepository, registry storage.RegistryRepository, timerTimeout time.Duration) Watcher {
	return &watcher{
		workerCluster:   workerCluster,
		url:             url,
		registry:        registry,
		urlTimerTimeout: timerTimeout,
		urlTimer:        time.NewTimer(timerTimeout),
		mu:              &sync.RWMutex{},
		jobDoneC:        make(chan bool),
		log:             log.WithFields(map[string]interface{}{}),
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
// WatchNewURLs loop through all url's and send to crawl if neither worker didn't do that
// it's helpful especially at start process if any url already exist in database
func (w watcher) WatchNewURLs(f func(*crawlerdpb.RequestURL)) {
	justNow := time.NewTimer(time.Second)

	// TODO: add tests - multiple jobs in short time
	job := func() {
		w.log.Debug("try to find new url's to start crawl")

		w.mu.Lock()
		if w.isJobRunning {
			w.retryJob = true
			w.mu.Unlock()
			return
		}
		w.retryJob = false
		w.isJobRunning = true
		w.mu.Unlock()

		defer func() {
			w.mu.Lock()
			w.isJobRunning = false
			w.mu.Unlock()
		}()

		// TODO: consider better solution than scrolling whole urls
		if err := w.url.Scroll(context.Background(), func(urls []v1.URL) {
			w.log.Debugf("found url's candidates to start crawl, len=%d", len(urls))

			for _, url := range urls {
				go func(url v1.URL) {
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

		// TODO: consider when watch urls should done
		//w.jobDoneC <- true
	}

	for {
		select {
		case <-w.urlTimer.C:
			job()
			w.urlTimer.Reset(w.urlTimerTimeout)
		case <-justNow.C:
			job()
			w.urlTimer.Reset(w.urlTimerTimeout)
		case <-w.jobDoneC:
			w.mu.RLock()
			if w.retryJob {
				w.mu.RUnlock()
				job()
				return
			}
			w.mu.RUnlock()
		}
	}
}

func (w *watcher) ResetTimer() {
	w.urlTimer.Reset(w.urlTimerTimeout)
}
