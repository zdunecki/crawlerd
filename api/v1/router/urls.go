package router

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/zdunecki/crawlerd/api"
	"github.com/zdunecki/crawlerd/api/v1/objects"
	"github.com/zdunecki/crawlerd/crawlerdpb"
	"github.com/zdunecki/crawlerd/pkg/meta/metav1"
)

func (r *router) urlsGetAll(c api.Context) {
	urls, err := r.store.URL().FindAll(c.RequestContext())
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	if urls == nil {
		urls = []metav1.URL{}
	}

	c.JSON(urls)
}

func (r *router) urlsHistoryGet(c api.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	history, err := r.store.History().FindByID(c.RequestContext(), id)
	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if history == nil {
		history = []metav1.History{}
	}

	c.JSON(history)
}

func (r *router) urlsCreate(c api.Context) {
	var req *objects.RequestPostURL

	data, err := ioutil.ReadAll(c.Request().Body)
	if data != nil && len(data) >= objects.DefaultMaxPOSTContentLength.Int() {
		c.RequestEntityTooLarge()
		return
	}

	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	if req.Interval < objects.IntervalMinValue {
		c.BadRequest()
		return
	}

	done, seq, err := r.store.URL().InsertOne(c.RequestContext(), req.URL, req.Interval)

	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if !done {
		c.BadRequest()
		return
	}

	if err := r.schedulerRetry(func() error {
		_, e := r.scheduler.AddURL(c.RequestContext(), &crawlerdpb.RequestURL{
			Id:       int64(seq),
			Url:      req.URL,
			Interval: int64(req.Interval),
		})
		return e
	}); err != nil {
		r.log.Error(err)
	}

	c.Created().JSON(&objects.ResponsePostURL{
		ID: seq,
	})
}

func (r *router) urlsPatch(c api.Context) {
	var req objects.RequestPatchURL

	id, err := c.ParamInt("id")
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	data, err := ioutil.ReadAll(c.Request().Body)
	if data != nil && len(data) >= objects.DefaultMaxPOSTContentLength.Int() {
		c.RequestEntityTooLarge()
		return
	}

	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	done, err := r.store.URL().UpdateOneByID(c.RequestContext(), id, objects.RequestPatchURL{
		URL:      req.URL,
		Interval: req.Interval,
	})
	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if !done {
		c.NotFound()
		return
	}

	if err := r.schedulerRetry(func() error {
		_, e := r.scheduler.UpdateURL(c.RequestContext(), &crawlerdpb.RequestURL{
			Id:       int64(id),
			Url:      *req.URL,
			Interval: int64(*req.Interval),
		})
		return e
	}); err != nil {
		r.log.Error(err)
	}

	c.JSON(&objects.ResponsePostURL{
		ID: id,
	})
}

func (r *router) urlsDelete(c api.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	//r.store.Linker().
	done, err := r.store.URL().DeleteOneByID(c.RequestContext(), id)

	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if !done {
		c.NotFound()
		return
	}

	if err := r.schedulerRetry(func() error {
		_, e := r.scheduler.DeleteURL(c.RequestContext(), &crawlerdpb.RequestDeleteURL{
			Id: int64(id),
		})
		return e
	}); err != nil {
		r.log.Error(err)
	}

	c.NoContent()
}
