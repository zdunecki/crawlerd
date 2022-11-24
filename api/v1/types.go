package v1

import (
	"github.com/zdunecki/crawlerd/api/v1/objects"
	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
)

type URL interface {
	Create(*objects.RequestPostURL) (*objects.ResponsePostURL, error)
	Patch(id string, data *objects.RequestPatchURL) (*objects.ResponsePostURL, error)
	Delete(id string) error
	All() ([]*metav1.URL, error)
	History(urlID string) ([]*metav1.History, error)
}

type RequestQueue interface {
	BatchCreate([]*metav1.RequestQueueCreate) (*objects.ResponseRequestQueueCreate, error)
}

type Linker interface {
	All() ([]*metav1.LinkNode, error)
}

type V1 interface {
	URL() URL
	RequestQueue() RequestQueue
	Linker() Linker
}
