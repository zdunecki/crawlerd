package router

import (
	"context"

	"github.com/zdunecki/crawlerd/api"
	"github.com/zdunecki/crawlerd/api/v1/objects"
	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
)

func (r *router) jobsGetAll(c api.Context) {
	if jobs, err := r.store.Job().FindAll(context.TODO()); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	} else {
		c.JSON(jobs)
	}
}

func (r *router) jobsGetByID(c api.Context) {
	id := c.Param("id")

	if job, err := r.store.Job().FindOneByID(context.TODO(), id); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	} else {
		c.JSON(job)
	}
}

func (r *router) jobsCreate(c api.Context) {
	req := &metav1.JobCreate{}

	if err := c.Bind(req); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	if err := req.Validate(); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeValidation,
		})
		return
	}

	if id, err := r.store.Job().InsertOne(context.TODO(), req); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	} else {
		c.JSON(map[string]string{
			"id": id,
		})
	}
}

func (r *router) jobsPath(c api.Context) {
	req := &metav1.JobPatch{}

	if err := c.Bind(req); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	id := c.Param("id")

	if job, err := r.store.Job().FindOneByID(context.TODO(), id); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	} else {
		req.ApplyJob(job)

		if err := req.Validate(); err != nil {
			r.log.Error(err)
			c.BadRequest().JSON(&objects.APIError{
				Type: objects.ErrorTypeValidation,
			})
			return
		}
	}

	if err := r.store.Job().PatchOneByID(context.TODO(), id, req); err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	c.JSON("ok")
}
