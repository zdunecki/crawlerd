package worker

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"crawlerd/pkg/util"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	//TODO: configurable chan buf length
	DefaultQueueLength     = 10000
	HTTPClientTimeout      = time.Second * 5
	HeaderContentLengthMax = 5 * util.MB
)

const (
	pagePubSubTopic = "page"
)

type (
	QueueInterval int64
	CrawlerStopCB func(chan objects.CrawlURL)
)

type Crawler interface {
	Enqueue(objects.CrawlURL)
	Update(objects.CrawlURL)
	Dequeue(int64)

	Stop(CrawlerStopCB)

	newIntervalQue(crawlURL objects.CrawlURL)
	crawl(*time.Ticker, QueueInterval)
	fetchContent(crawl objects.CrawlURL) error
	httpRequest(string) (*http.Response, error)
}

type crawler struct {
	mu     sync.RWMutex
	wgStop *sync.WaitGroup

	httpClient *http.Client

	worker   Worker
	history  storage.HistoryRepository
	registry storage.RegistryRepository

	processor  *processor
	pageLookup *pageLookup
	pubsub     pubsub.PubSub
	compressor Compressor

	stopC       map[QueueInterval]chan bool
	ticker      map[QueueInterval]*time.Ticker
	queueC      map[QueueInterval]chan objects.CrawlURL
	closeQueueC chan objects.CrawlURL

	log *log.Entry
}

func NewCrawler(storage storage.Storage, worker Worker, pubsub pubsub.PubSub, compressor Compressor, httpClient *http.Client) Crawler {
	c := &crawler{
		wgStop: &sync.WaitGroup{},

		httpClient: httpClient,

		worker:     worker,
		history:    storage.History(),
		registry:   storage.Registry(),
		processor:  NewProcessor(),
		pageLookup: NewPageLookup(),
		pubsub:     pubsub,
		compressor: compressor,

		stopC:       make(map[QueueInterval]chan bool),
		ticker:      make(map[QueueInterval]*time.Ticker),
		queueC:      make(map[QueueInterval]chan objects.CrawlURL),
		closeQueueC: make(chan objects.CrawlURL, DefaultQueueLength),

		log: log.WithFields(map[string]interface{}{
			"service": "crawler",
		}),
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: HTTPClientTimeout,
		}
	}

	return c
}

func (c *crawler) Dequeue(id int64) {
	if err := c.registry.DeleteURLByID(context.Background(), int(id)); err != nil {
		c.log.Error(err)
	}
}

func (c *crawler) Update(crawlURL objects.CrawlURL) {
	if err := c.registry.PutURL(context.Background(), crawlURL); err != nil {
		c.log.Error(err)
	}
}

func (c *crawler) Enqueue(crawlURL objects.CrawlURL) {
	c.mu.Lock()
	defer c.mu.Unlock()

	crawlURL.WorkerID = c.worker.ID()

	intervalID := QueueInterval(crawlURL.Interval)

	_, intervalQueExists := c.ticker[intervalID]

	if intervalQueExists {
		c.queueC[intervalID] <- crawlURL
		if err := c.registry.PutURL(context.Background(), crawlURL); err != nil {
			return
		}

		if err := c.fetchContent(crawlURL); err != nil {
			c.log.Error(err)
		}

		return
	}

	c.newIntervalQue(crawlURL)
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

func (c *crawler) newIntervalQue(crawlURL objects.CrawlURL) {
	intervalID := QueueInterval(crawlURL.Interval)

	ticker := time.NewTicker(time.Second * time.Duration(crawlURL.Interval))
	c.stopC[intervalID] = make(chan bool)
	c.ticker[intervalID] = ticker
	c.queueC[intervalID] = make(chan objects.CrawlURL, DefaultQueueLength)
	c.queueC[intervalID] <- crawlURL

	if err := c.registry.PutURL(context.Background(), objects.CrawlURL{
		Id:       crawlURL.Id,
		Url:      crawlURL.Url,
		Interval: crawlURL.Interval,
		WorkerID: crawlURL.WorkerID,
	}); err != nil {
		return
	}

	go func(crawlURL objects.CrawlURL) {
		if err := c.fetchContent(crawlURL); err != nil {
			c.log.Error(err)
		}
		c.crawl(ticker, QueueInterval(crawlURL.Interval))
	}(crawlURL)
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
						defer wg.Done()

						if err := c.fetchContent(crawlQ); err != nil {
							c.log.Error(err)
						}
					}()

					// TODO: in-memory may increase performance? consider how to get current value in memory instead of in distributed kv db.
					// TODO: on of the solutions may be high efficient in-memory cache with kv watchers about update's information.
					reCrawlUrl, err := c.registry.GetURLByID(context.Background(), int(crawlQ.Id))
					if err != nil {
						c.log.Error(err)
						continue
					}

					if reCrawlUrl == nil {
						c.log.Warn("url is nil")
						continue
					}

					c.log.Debug("try requeue")

					crawlIsUpdated := cmp.Diff(*reCrawlUrl, crawlQ) != ""

					if !crawlIsUpdated {
						c.queueC[intervalID] <- *reCrawlUrl
						c.log.Debug("requeued successfully")
						continue
					}

					newInterval := QueueInterval(reCrawlUrl.Interval)
					c.log.Debug("requeued queue must be updated")

					if newInterval == intervalID {
						c.queueC[intervalID] <- *reCrawlUrl
						c.log.Debug("requeued queue updated successfully")
						continue
					}

					c.mu.Lock()

					_, queueExists := c.queueC[newInterval]

					if !queueExists {
						c.log.Debug("create new channel for queue")
						c.queueC[newInterval] = make(chan objects.CrawlURL, DefaultQueueLength)

						newTick := time.NewTicker(time.Second * time.Duration(newInterval))
						go c.crawl(newTick, newInterval)
					}

					c.queueC[newInterval] <- *reCrawlUrl
					c.mu.Unlock()
					c.log.Debug("requeued to another QueueInterval successfully")
				}
			}

			wg.Wait()
		}
	}
}

// TODO: gridfs vs ioutil.ReadAll
func (c *crawler) fetchContent(crawl objects.CrawlURL) error {
	c.log.Debug("trying fetch content")

	start := time.Now()
	resp, err := c.httpRequest(crawl.Url)
	if err != nil {
		return err
	}

	createHistory := func(body []byte, finish time.Time) error {
		l := c.log.WithFields(log.Fields{
			"id": crawl.Id,
		})

		writeData := body

		if c.compressor != nil {
			compressed, err := c.compressor(writeData)
			if err != nil {
				l.Debugf("err: %v", err)
			} else if compressed != nil {
				writeData = compressed
			}
		}

		if _, _, err := c.history.InsertOne(context.Background(), int(crawl.Id), writeData, finish.Sub(start), start); err != nil {
			return err
		}

		l.Debug("content added to history")
		return nil
	}

	// TODO: stream to db instead of read all to memory
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.log.Error(err)
		body = nil
		return createHistory(body, time.Now())
	}

	status := c.pageLookup.status(resp, crawl.Url)

	go func() {
		c.log.Debugf("page_discovery_status: %s", status)

		switch status {
		case PageStatusProcess:
			// TODO: what if body is not html
			if err := c.processor.processHTML(body); err != nil {
				c.log.Error(err)
			}

		case PageStatusStream: // TODO: queue
			topicMsg := &crawlerdpb.TopicMessage{
				Url:  crawl.Url,
				Body: body,
			}
			msg, err := proto.Marshal(topicMsg)
			if err != nil {
				c.log.Error(err)
				return
			}

			if err := c.pubsub.Publish(pagePubSubTopic, msg); err != nil {
				c.log.Error(err)
			}
		}
	}()

	return createHistory(body, time.Now())
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
