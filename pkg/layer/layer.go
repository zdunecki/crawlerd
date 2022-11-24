package layer

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/zdunecki/crawlerd/pkg/worker"
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
	resp, err := l.httpRequest("https://livesession.io")
	if err != nil {
		return "", err
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return "", err
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
