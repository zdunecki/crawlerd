package router

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"crawlerd/api"
	"crawlerd/api/v1/objects"
	"crawlerd/pkg/meta/metav1"
)

func (r *router) seedList(c api.Context) {
	seedList, err := r.store.Seed().List(c.RequestContext())
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	if seedList == nil {
		seedList = []*metav1.Seed{}
	}

	c.JSON(seedList)
}

func (r *router) seedAppend(c api.Context) {
	var req *objects.RequestAppendSeed

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

	if err := r.store.Seed().Append(c.RequestContext(), req.Seed); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	c.Created().JSON(&objects.ResponseAppendSeed{
		OK: true,
	})
}

func (r *router) seedDelete(c api.Context) {
	id := c.Param("id")

	if errors, err := r.store.Seed().DeleteMany(c.RequestContext(), id); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	} else {
		r.log.Error(errors)

		c.JSON(errors)
		return
	}

	c.NoContent()
}
