package layer

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"crawlerd/pkg/worker"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type layer struct {
	id         uint
	httpClient *http.Client
}

func NewLayerEngine() *layer {
	return &layer{
		httpClient: http.DefaultClient,
		id:         0,
	}
}

func (l *layer) Diff(bodyA, bodyB io.Reader) (string, error) {
	dmp := diffmatchpatch.New()

	b, err := ioutil.ReadFile("./page.json")
	if err != nil {
		return "", nil
	}
	diffs := dmp.DiffMain(string(b), string(b), true)

	return dmp.DiffText1(diffs), nil
}

func (l *layer) httpRequest(endpoint string) (*http.Response, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	head, err := http.Head(endpoint)
	if err != nil {
		return nil, err
	}

	if head.StatusCode != http.StatusOK {
		return nil, worker.ErrHTTPNotOK
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	return l.httpClient.Do(req)
}
