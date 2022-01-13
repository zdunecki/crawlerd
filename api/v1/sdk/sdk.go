package sdk

// TODO: client vs sdk vs something else ?

// TODO: auto generated client

import (
	"errors"
	"net/http"
	"time"

	v1 "crawlerd/api/v1"
)

func NewWithOpts(opts ...Option) (v1.V1, error) {
	opt := &options{}

	for _, o := range opts {
		o(opt)
	}

	if opt.addr != "" {
		return newHTTPClient(opt.addr+"/v1", &http.Client{
			Timeout: time.Second * 15, // TODO: timeout config
		}), nil
	}

	return nil, errors.New("api url is not defined")
}
