package worker

import (
	"context"
	"sync"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/apikit/pkg/scheduler"
	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	log "github.com/sirupsen/logrus"
)

type Controller interface {
	ReAttachResources(chan *metav1.RequestQueue)
}

type controller struct {
	scheduler    crawlerdpb.SchedulerClient
	requestQueue store.RequestQueue

	log *log.Entry
}

func NewController(scheduler crawlerdpb.SchedulerClient, requestQueue store.RequestQueue) Controller {
	return &controller{
		scheduler:    scheduler,
		requestQueue: requestQueue,

		log: log.WithFields(map[string]interface{}{
			"service": "controller",
		}),
	}
}

// TODO: what should do on k8s?
func (c *controller) ReAttachResources(urlC chan *metav1.RequestQueue) {
	c.log.Debugln("attach jobs to another workers...")

	wg := sync.WaitGroup{}

	if urlC == nil || len(urlC) == 0 {
		return
	}

	urlCLen := len(urlC)
	wg.Add(urlCLen)

	i := 0
	for rq := range urlC {
		func() {
			defer func() {
				wg.Done()
				i++
			}()
			c.log.Debugf("attaching id=%s", rq.ID)

			if _, err := c.requestQueue.DeleteOneByID(context.Background(), rq.ID); err != nil {
				c.log.Error(err)
				return
			}

			// TODO: retry
			if _, err := c.scheduler.AddURL(context.Background(), &crawlerdpb.RequestURL{
				Id: 0,
				//Id:       crawl.ID,
				Url: rq.URL,
				//Interval: crawl.Interval,
				Lease: true,
			}); err != nil && err != scheduler.ErrNoWorkers {
				c.log.Error(err)
			}

			return
		}()

		if i >= urlCLen {
			break
		}
	}

	wg.Wait()
}
