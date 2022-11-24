package router

import (
	"github.com/zdunecki/crawlerd/api"
	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
)

func (r *router) linkerGetAll(c api.Context) {
	nodes, err := r.store.Linker().FindAll(c.RequestContext())
	if err != nil {
		r.log.Error(err)
		c.BadRequest()
		return
	}

	if nodes == nil {
		nodes = []*metav1.LinkNode{}
	}

	c.JSON(nodes)
}
