package roundrobin

import (
	"sync"
)

type roundrobin struct {
	mu       sync.Mutex
	next     int
	itemsLen int
	items    []interface{}
}

func New(items []interface{}) *roundrobin {
	return &roundrobin{
		items:    items,
		itemsLen: len(items),
	}
}

func (rr *roundrobin) Next() interface{} {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	r := rr.items[rr.next]
	rr.next = (rr.next + 1) % rr.itemsLen

	return r
}
