package v1

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"github.com/zdunecki/crawlerd/api"
	"github.com/zdunecki/crawlerd/api/v1/objects"
	"github.com/zdunecki/crawlerd/api/v1/router"
	"github.com/zdunecki/crawlerd/crawlerdpb"
	"github.com/zdunecki/crawlerd/pkg/store"
)

type v1 struct {
	store store.Repository

	scheduler crawlerdpb.SchedulerClient

	schedulerBackoff *backoff.ExponentialBackOff

	log *log.Entry
}

func New(opts ...Option) (*v1, error) {
	v := &v1{
		log: log.WithFields(map[string]interface{}{
			"service": "api",
		}),
	}

	for _, o := range opts {
		if err := o(v); err != nil {
			return nil, err
		}
	}

	{
		bo := backoff.NewExponentialBackOff()
		bo.MaxInterval = time.Second * 2
		bo.MaxElapsedTime = time.Second * 15

		v.schedulerBackoff = bo
	}

	return v, nil
}

func (apiv1 *v1) Serve(addr string, v1 api.API) error {
	if apiv1.store == nil {
		return objects.ErrNoStorage
	}

	if apiv1.scheduler == nil {
		return objects.ErrNoScheduler
	}

	if err := router.New(v1, apiv1.store, apiv1.scheduler, apiv1.schedulerRetry, apiv1.log); err != nil {
		return err
	}

	apiv1.log.Info("listening on: ", addr)

	return http.ListenAndServe(addr, v1.Handler())
}

func (apiv1 *v1) Store() store.Repository {
	return apiv1.store
}

func (apiv1 *v1) schedulerRetry(f func() error) error {
	return backoff.Retry(f, apiv1.schedulerBackoff)
}
