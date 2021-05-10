package worker

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/util"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

const (
	//TODO: configurable chan buf length
	DefaultQueueLength     = 10000
	HTTPClientTimeout      = time.Second * 5
	HeaderContentLengthMax = 256 * util.KB
)

type (
	QueueInterval int64
	CrawlerStopCB func(chan CrawlURL)
)

type (
	// TODO: protobuf instead of json
	CrawlURL struct {
		Id       int64
		Url      string
		Interval int64
		WorkerID string
	}
)

type Crawler interface {
	Enqueue(CrawlURL)
	Update(CrawlURL)
	Dequeue(int64)

	Stop(CrawlerStopCB)

	crawl(*time.Ticker, QueueInterval)
	httpRequest(string) (*http.Response, error)
}

type crawler struct {
	mu     sync.RWMutex
	wgStop *sync.WaitGroup

	httpClient *http.Client

	worker   Worker
	storage  storage.Client
	registry Registry

	stopC       map[QueueInterval]chan bool
	ticker      map[QueueInterval]*time.Ticker
	queueC      map[QueueInterval]chan CrawlURL
	closeQueueC chan CrawlURL
}

func NewCrawler(storage storage.Client, registry Registry, worker Worker) Crawler {
	return &crawler{
		wgStop: &sync.WaitGroup{},

		httpClient: &http.Client{
			Timeout: HTTPClientTimeout,
		},

		worker:   worker,
		storage:  storage,
		registry: registry,

		stopC:       make(map[QueueInterval]chan bool),
		ticker:      make(map[QueueInterval]*time.Ticker),
		queueC:      make(map[QueueInterval]chan CrawlURL),
		closeQueueC: make(chan CrawlURL, DefaultQueueLength),
	}
}

func (c *crawler) Dequeue(id int64) {
	if err := c.registry.DeleteURLByID(int(id)); err != nil {
		log.Error(err)
	}
}

func (c *crawler) Update(crawlURL CrawlURL) {
	if err := c.registry.PutURL(crawlURL); err != nil {
		log.Error(err)
	}
}

func (c *crawler) Enqueue(crawlURL CrawlURL) {
	c.mu.Lock()
	defer c.mu.Unlock()

	crawlURL.WorkerID = c.worker.ID()

	intervalID := QueueInterval(crawlURL.Interval)

	_, exists := c.ticker[intervalID]

	if exists {
		c.queueC[intervalID] <- crawlURL
		if err := c.registry.PutURL(crawlURL); err != nil {
			return
		}

		if err := c.fetchContent(crawlURL); err != nil {
			log.Error(err)
		}

		return
	}

	ticker := time.NewTicker(time.Second * time.Duration(crawlURL.Interval))
	c.stopC[intervalID] = make(chan bool)
	c.ticker[intervalID] = ticker
	c.queueC[intervalID] = make(chan CrawlURL, DefaultQueueLength)
	c.queueC[intervalID] <- crawlURL

	if err := c.registry.PutURL(crawlURL); err != nil {
		return
	}

	go func(crawlURL CrawlURL) {
		if err := c.fetchContent(crawlURL); err != nil {
			log.Error(err)
		}
		c.crawl(ticker, QueueInterval(crawlURL.Interval))
	}(crawlURL)
}

func (c *crawler) Stop(cb CrawlerStopCB) {
	if c.queueC == nil || len(c.queueC) == 0 {
		return
	}

	if len(c.ticker) > 0 {
		c.wgStop.Add(1)
	}
	for intervalID, _ := range c.ticker {
		c.stopC[intervalID] <- true
	}

	c.wgStop.Wait()
	cb(c.closeQueueC)
}

func (c *crawler) crawl(ticker *time.Ticker, intervalID QueueInterval) {
	for {
		select {
		case <-c.stopC[intervalID]:
			ticker.Stop()
			queue := c.queueC[intervalID]

			queueLen := len(queue)

			if queueLen > 0 {
				i := 0
				for crawlQ := range queue {
					i++
					c.closeQueueC <- crawlQ

					if i >= queueLen {
						break
					}
				}
			}

			c.wgStop.Done()
		case <-ticker.C:
			queue := c.queueC[intervalID]
			queueLen := len(queue)

			wg := sync.WaitGroup{}

			crawledLen := 0
			wg.Add(queueLen)

			for {
				if crawledLen >= queueLen {
					break
				}
				crawlQ := <-queue

				select {
				default:
					crawledLen++

					go func() {
						log.Info("trying fetch resource")
						defer wg.Done()

						if err := c.fetchContent(crawlQ); err != nil {
							log.Error(err)
						}
					}()

					reCrawlUrl, err := c.registry.GetURLByID(int(crawlQ.Id))
					if err != nil {
						log.Error(err)
						continue
					}

					if reCrawlUrl == nil {
						log.Warn("url is nil")
						continue
					}

					log.Info("try requeue")

					crawlIsUpdated := cmp.Diff(*reCrawlUrl, crawlQ) != ""

					if !crawlIsUpdated {
						c.queueC[intervalID] <- *reCrawlUrl
						log.Info("requeued successfully")
						continue
					}

					newInterval := QueueInterval(reCrawlUrl.Interval)
					log.Info("requeued queue must be updated")

					if newInterval == intervalID {
						c.queueC[intervalID] <- *reCrawlUrl
						log.Info("requeued queue updated successfully")
						continue
					}

					c.mu.Lock()

					_, queueExists := c.queueC[newInterval]

					if !queueExists {
						log.Info("create new channel for queue")
						c.queueC[newInterval] = make(chan CrawlURL, DefaultQueueLength)

						newTick := time.NewTicker(time.Second * time.Duration(newInterval))
						go c.crawl(newTick, newInterval)
					}

					c.queueC[newInterval] <- *reCrawlUrl
					c.mu.Unlock()
					log.Info("requeued to another QueueInterval successfully")
				}
			}

			wg.Wait()
		}
	}
}

func (c *crawler) fetchContent(crawl CrawlURL) error {
	log.Info("trying fetch resource")

	start := time.Now()
	resp, err := c.httpRequest(crawl.Url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	finish := time.Now()

	if err != nil {
		log.Error(err)
		body = nil
	}

	if _, _, err := c.storage.History().InsertOne(context.Background(), int(crawl.Id), body, finish.Sub(start), start); err != nil {
		return err
	}

	log.Info("content added to history")

	return nil
}

func (c *crawler) httpRequest(endpoint string) (*http.Response, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	head, err := http.Head(endpoint)
	if err != nil {
		return nil, err
	}

	if head.StatusCode != http.StatusOK {
		return nil, ErrHTTPNotOK
	}

	if head.ContentLength > HeaderContentLengthMax.Int64() {
		return nil, ErrHTTPMaxContentLength
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}
