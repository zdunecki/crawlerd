package router

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"regexp"

	"crawlerd/api"
	"crawlerd/api/v1/objects"
	"crawlerd/pkg/meta/metav1"
)

// TODO urls are now queues
// TODO: auth
// TODO: batch errors
// TODO: ScrapeLinksPattern, FollowLinks on frontend side
func (r *router) requestQueueBatchCreate(c api.Context) {
	var req []*metav1.RequestQueueCreateAPI
	rq := make([]*metav1.RequestQueueCreate, 0)

	rs := r.store.Runner()

	// TODO: body limitations?
	data, err := ioutil.ReadAll(c.Request().Body)
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

	{
		linkNodes := make([]*metav1.LinkNodeCreate, 0)

		for _, ri := range req {
			if err := ri.Validate(); err != nil {
				r.log.Error(err)
				c.BadRequest().JSON(&objects.APIError{
					Type:    objects.ErrorTypeValidation,
					Message: err.Error(),
				})
				return
			}

			runner, err := rs.GetByID(c.RequestContext(), ri.RunID)
			if err != nil {
				c.InternalError().JSON(&objects.APIError{
					Type: objects.ErrorTypeInternal,
				})
				return
			}

			addLinkNode := func() {
				linkNodes = append(linkNodes, &metav1.LinkNodeCreate{
					URL: metav1.NewLinkURL(ri.URL),
				})
			}

			if runner.RunnerConfig.ScrapeLinksPattern != "" {
				if re, err := regexp.Compile(runner.RunnerConfig.ScrapeLinksPattern); err != nil {
					r.log.Error(err)
				} else {
					if re.Match([]byte(ri.URL)) {
						addLinkNode()
					}
				}
			} else {
				addLinkNode()
			}

			addRequestQue := func() {
				rq = append(rq, &metav1.RequestQueueCreate{
					RunID:  ri.RunID,
					URL:    ri.URL,
					Depth:  ri.Depth,
					Status: metav1.RequestQueueStatusQueued,
				})
			}

			if runner.RunnerConfig.FollowLinks == nil {
				addRequestQue()
			} else {
				shouldAddRq := false
				for _, filter := range runner.RunnerConfig.FollowLinks {
					if filter.Match != "" {
						re, _ := regexp.Compile(filter.Match)
						if re.MatchString(ri.URL) {
							shouldAddRq = true
							break
						}
					} else if filter.Is != "" {
						if ri.URL == filter.Is {
							shouldAddRq = true
							break
						}
					}
				}

				if shouldAddRq {
					addRequestQue()
				}
			}

		}

		_, err := r.store.Linker().InsertManyIfNotExists(c.RequestContext(), linkNodes)
		if err != nil {
			r.log.Error(err)
			c.InternalError().JSON(&objects.APIError{
				Type: objects.ErrorTypeInternal,
			})
			return
		}
	}

	ids, err := r.store.RequestQueue().InsertMany(c.RequestContext(), rq)
	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	c.Created().JSON(&objects.ResponseRequestQueueCreate{
		IDs: ids,
	})
}
