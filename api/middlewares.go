package api

import (
	"net/http"

	"crawlerd/pkg/util"
)

type MiddleWare func(w http.ResponseWriter, r *http.Request)

func WithMaxBytes(n util.Byte) MiddleWare {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, int64(n))
	}
}
