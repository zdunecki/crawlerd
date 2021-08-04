package worker

import (
	"context"
	"sync"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/apikit/pkg/scheduler"
	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/storage"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	ReAttachResources(chan metav1.CrawlURL)
}

type controller struct {
	scheduler crawlerdpb.SchedulerClient
	registry  storage.RegistryRepository

	log *log.Entry
}

func NewController(scheduler crawlerdpb.SchedulerClient, registry storage.RegistryRepository) Controller {
	return &controller{
		scheduler: scheduler,
		registry:  registry,

		log: log.WithFields(map[string]interface{}{
			"service": "controller",
		}),
	}
}

// TODO: what should do on k8s?
func (c *controller) ReAttachResources(urlC chan metav1.CrawlURL) {
	c.log.Debugln("attach jobs to another workers...")

	wg := sync.WaitGroup{}

	if urlC == nil || len(urlC) == 0 {
		return
	}

	urlCLen := len(urlC)
	wg.Add(urlCLen)

	i := 0
	for crawl := range urlC {
		func(crawl metav1.CrawlURL) {
			defer func() {
				wg.Done()
				i++
			}()
			c.log.Debugf("attaching id=%d", crawl.Id)

			if err := c.registry.DeleteURL(context.Background(), crawl); err != nil {
				c.log.Error(err)
				return
			}

			// TODO: retry
			if _, err := c.scheduler.AddURL(context.Background(), &crawlerdpb.RequestURL{
				Id:       crawl.Id,
				Url:      crawl.Url,
				Interval: crawl.Interval,
				Lease:    true,
			}); err != nil && err != scheduler.ErrNoWorkers {
				c.log.Error(err)
			}

			return
		}(crawl)

		if i >= urlCLen {
			break
		}
	}

	wg.Wait()
}
