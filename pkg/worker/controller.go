package worker

import (
	context "context"
	"sync"

	"crawlerd/crawlerdpb"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	ReAttachResources(chan CrawlURL)
}

type controller struct {
	scheduler crawlerdpb.SchedulerClient
	registry  Registry
}

func NewController(scheduler crawlerdpb.SchedulerClient, registry Registry) Controller {
	return &controller{
		scheduler: scheduler,
		registry:  registry,
	}
}

func (c *controller) ReAttachResources(urlC chan CrawlURL) {
	log.Infoln("attach jobs to another workers...")

	wg := sync.WaitGroup{}

	if urlC == nil || len(urlC) == 0 {
		return
	}

	urlCLen := len(urlC)
	wg.Add(urlCLen)

	i := 0
	for crawl := range urlC {
		func(crawl CrawlURL) {
			defer func() {
				wg.Done()
				i++
			}()
			log.Infof("attaching id=%d", crawl.Id)

			if err := c.registry.DeleteURL(crawl); err != nil {
				log.Error(err)
				return
			}

			if _, err := c.scheduler.AddURL(context.Background(), &crawlerdpb.RequestURL{
				Id:       crawl.Id,
				Url:      crawl.Url,
				Interval: crawl.Interval,
			}); err != nil {
				log.Error(err)
			}

			return
		}(crawl)

		if i >= urlCLen {
			break
		}
	}

	wg.Wait()
}
