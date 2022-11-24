package scheduler

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zdunecki/crawlerd/crawlerdpb"
	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
	"github.com/zdunecki/crawlerd/pkg/store"
	"github.com/zdunecki/crawlerd/pkg/worker"
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
	linker        store.Linker
	mu            *sync.RWMutex
	jobDoneC      chan bool
	isJobRunning  bool
	retryJob      bool

	urlTimerTimeout time.Duration
	urlTimer        *time.Timer

	log *log.Entry
}

func NewWatcher(workerCluster worker.Cluster, linker store.Linker, timerTimeout time.Duration) Watcher {
	return &watcher{
		workerCluster:   workerCluster,
		linker:          linker,
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
// TODO: watcher should work similar to message queues, another option is to use production ready message queue like NATs but everything should be under watcher abstraction

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
		// TODO: scroll not acked request queues
		if err := w.linker.Scroll(context.Background(), func(nodes []*metav1.LinkNode) {
			w.log.Debugf("found url's candidates to start crawl, len=%d", len(nodes))

			for _, node := range nodes {
				go func(node *metav1.LinkNode) {
					resp, err := w.linker.Live().FindOneByID(context.Background(), node.ID)

					if err != nil {
						log.Error(err)
						return
					}

					isCrawling := resp != nil

					if isCrawling {
						return
					}

					f(&crawlerdpb.RequestURL{
						//Id:  int64(node.ID),
						Url: node.URL.ToString(),
						//Interval: int64(url.Interval),
					})
				}(node)
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
