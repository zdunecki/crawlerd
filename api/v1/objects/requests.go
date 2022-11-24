package objects

import "github.com/zdunecki/crawlerd/pkg/meta/metav1"

// TODO: use struct from meta pkg
type RequestPostURL struct {
	URL string `json:"url"`
	// Deprecated: Interval is not important now
	Interval int `json:"interval"`
}

type RequestPatchURL struct {
	URL *string `json:"url" bson:"url,omitempty"`

	// Deprecated: Interval is not important now
	Interval *int `json:"interval" bson:"interval,omitempty"`
}

type RequestAppendSeed struct {
	Seed []*metav1.Seed `json:"seed"`
}
