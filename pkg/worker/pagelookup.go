package worker

import (
	"net/http"

	"github.com/zdunecki/crawlerd/pkg/util"
)

type PageStatus string

const (
	PageStatusProcess PageStatus = "process"
	PageStatusStream  PageStatus = "stream"
)

type pageLookup struct{}

func NewPageLookup() *pageLookup {
	return &pageLookup{}
}

// TODO: return "status" based on current crawlURL
func (pd *pageLookup) status(response *http.Response, crawlURL string) PageStatus {
	isHTML := util.IsHTMLHeader(response.Header)

	if isHTML {
		return PageStatusProcess
	}

	return PageStatusProcess
	// TODO: stream
	// return PageStatusStream
}
