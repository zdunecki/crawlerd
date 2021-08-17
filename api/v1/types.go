package v1

import (
	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/util"
)

const (
	DefaultMaxPOSTContentLength = util.KB * 4
)

type URL interface {
	Create(*RequestPostURL) (*ResponsePostURL, error)
	Patch(id string, data *RequestPatchURL) (*ResponsePostURL, error)
	Delete(id string) error
	All() ([]*metav1.URL, error)
	History(urlID string) ([]*metav1.History, error)
}

type RequestQueue interface {
	BatchCreate([]*metav1.RequestQueueCreate) (*ResponseRequestQueueCreate, error)
}

type Linker interface {
	All() ([]*metav1.LinkNode, error)
}

type V1 interface {
	URL() URL
	RequestQueue() RequestQueue
	Linker() Linker
}

// TODO: use struct from meta pkg

type (
	RequestPostURL struct {
		URL string `json:"url"`
		// Deprecated: Interval is not important now
		Interval int `json:"interval"`
	}

	RequestPatchURL struct {
		URL *string `json:"url" bson:"url,omitempty"`

		// Deprecated: Interval is not important now
		Interval *int `json:"interval" bson:"interval,omitempty"`
	}
)

type (
	ResponsePostURL struct {
		ID int `json:"id"`
	}
)

// below dont move/delete

type ResponseRequestQueueCreate struct {
	IDs []string `json:"ids"`
}
