package v1

import "crawlerd/pkg/util"

const (
	DefaultMaxPOSTContentLength = util.KB * 4
)

type (
	RequestPostURL struct {
		URL      string `json:"url"`
		Interval int    `json:"interval"`
	}

	RequestPatchURL struct {
		URL      *string `json:"url" bson:"url,omitempty"`
		Interval *int    `json:"interval" bson:"interval,omitempty"`
	}
)


type (
	ResponsePostURL struct {
		ID int `json:"id"`
	}
)