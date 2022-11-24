package sdk

import (
	"bytes"
	"encoding/json"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	v1 "github.com/zdunecki/crawlerd/api/v1"
	"github.com/zdunecki/crawlerd/api/v1/objects"
)

type rest interface {
	get(resource string, outPtr interface{}) error

	post(resource string, body interface{}, outPtr interface{}) error

	patch(resource string, body interface{}, outPtr interface{}) error

	put(resource string, body interface{}, outPtr interface{}) error

	delete(resource string, body interface{}, outPtr interface{}) error
}

type httpClient struct {
	apiURL string
	http   *http.Client

	url          v1.URL
	requestQueue v1.RequestQueue
	linker       v1.Linker
}

func newHTTPClient(apiURL string, http *http.Client) v1.V1 {
	c := &httpClient{
		apiURL: apiURL,
		http:   http,
	}

	c.url = &httpURL{
		rest: &httpClient{
			apiURL: apiURL + "/urls",
			http:   http,
		},
	}
	c.requestQueue = &httpRequestQueue{
		client: c,
	}
	c.linker = &httpLinker{
		client: c,
	}

	return c
}

func (c *httpClient) URL() v1.URL {
	return c.url
}

func (c *httpClient) RequestQueue() v1.RequestQueue {
	return c.requestQueue

}

func (c *httpClient) Linker() v1.Linker {
	return c.linker
}

func (c *httpClient) get(resource string, outPtr interface{}) error {
	resp, err := c.http.Get(c.apiURL + resource)
	if err != nil {
		return err
	}

	return jsoniter.NewDecoder(resp.Body).Decode(outPtr)
}

func (c *httpClient) post(resource string, body interface{}, outPtr interface{}) error {
	return c.request(http.MethodPost, resource, body, outPtr)
}

func (c *httpClient) patch(resource string, body interface{}, outPtr interface{}) error {
	return c.request(http.MethodPatch, resource, body, outPtr)
}

func (c *httpClient) put(resource string, body interface{}, outPtr interface{}) error {
	return c.request(http.MethodPut, resource, body, outPtr)
}

func (c *httpClient) delete(resource string, body interface{}, outPtr interface{}) error {
	return c.request(http.MethodDelete, resource, body, outPtr)
}

func (c *httpClient) request(method, resource string, body interface{}, outPtr interface{}) error {
	var bodyReader *bytes.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}

		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.apiURL+resource, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application-json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		if outPtr != nil {
			return jsoniter.NewDecoder(resp.Body).Decode(outPtr)
		}

		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < 600 {
		apiErr := &objects.APIError{}

		if err := jsoniter.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return err
		}

		return apiErr
	}

	return nil
}
