package client

import (
	"errors"

	v1 "crawlerd/api/v1"
)

func NewWithOpts(opts ...Option) (v1.V1, error) {
	opt := &options{}

	for _, o := range opts {
		o(opt)
	}

	if opt.addr != "" {
		return &httpClient{
			url: newHTTPURL(opt.addr),
		}, nil
	}

	return nil, errors.New("api url is not defined")
}
