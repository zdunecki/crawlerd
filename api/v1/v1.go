package v1

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"crawlerd/api"
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
)

const (
	IntervalMinValue = 5
)

type V1URL interface {
	Create(*RequestPostURL) (*ResponsePostURL, error)
	Patch(id string, data *RequestPatchURL) (*ResponsePostURL, error)
	Delete(id string) error
	All() ([]*objects.URL, error)
	History(urlID string) ([]*objects.History, error)
}

type V1 interface {
	URL() V1URL
}

type v1 struct {
	storage   storage.Storage
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

func (v *v1) Serve(addr string, v1 api.API) error {
	if v.storage == nil {
		return ErrNoStorage
	}

	if v.scheduler == nil {
		return ErrNoScheduler
	}

	v1.Post("/api/urls", func(ctx api.Context) {
		var req *RequestPostURL

		data, err := ioutil.ReadAll(ctx.Request().Body)
		if data != nil && len(data) >= DefaultMaxPOSTContentLength.Int() {
			ctx.RequestEntityTooLarge()
			return
		}

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if req.Interval < IntervalMinValue {
			ctx.BadRequest()
			return
		}

		done, seq, err := v.storage.URL().InsertOne(ctx.RequestContext(), req.URL, req.Interval)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.BadRequest()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.AddURL(ctx.RequestContext(), &crawlerdpb.RequestURL{
				Id:       int64(seq),
				Url:      req.URL,
				Interval: int64(req.Interval),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.Created().JSON(&ResponsePostURL{
			ID: seq,
		})
	}, api.WithMaxBytes(DefaultMaxPOSTContentLength))

	v1.Patch("/api/urls/{id}", func(ctx api.Context) {
		var req RequestPatchURL

		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		data, err := ioutil.ReadAll(ctx.Request().Body)
		if data != nil && len(data) >= DefaultMaxPOSTContentLength.Int() {
			ctx.RequestEntityTooLarge()
			return
		}

		if err != nil {
			log.Error(err)
			ctx.InternalError()
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.storage.URL().UpdateOneByID(ctx.RequestContext(), id, RequestPatchURL{
			URL:      req.URL,
			Interval: req.Interval,
		})
		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.NotFound()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.UpdateURL(ctx.RequestContext(), &crawlerdpb.RequestURL{
				Id:       int64(id),
				Url:      *req.URL,
				Interval: int64(*req.Interval),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.JSON(&ResponsePostURL{
			ID: id,
		})
	}, api.WithMaxBytes(DefaultMaxPOSTContentLength))

	v1.Delete("/api/urls/{id}", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.storage.URL().DeleteOneByID(ctx.RequestContext(), id)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.NotFound()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.DeleteURL(ctx.RequestContext(), &crawlerdpb.RequestDeleteURL{
				Id: int64(id),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.NoContent()
	})

	v1.Get("/api/urls", func(ctx api.Context) {
		urls, err := v.storage.URL().FindAll(ctx.RequestContext())
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if urls == nil {
			urls = []objects.URL{}
		}

		ctx.JSON(urls)
	})

	v1.Get("/api/urls/{id}/history", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		history, err := v.storage.History().FindByID(ctx.RequestContext(), id)
		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if history == nil {
			history = []objects.History{}
		}

		ctx.JSON(history)
	})

	v.log.Info("listening on: ", addr)
	return http.ListenAndServe(addr, v1.Handler())
}

func (v *v1) schedulerRetry(f func() error) error {
	return backoff.Retry(f, v.schedulerBackoff)
}
