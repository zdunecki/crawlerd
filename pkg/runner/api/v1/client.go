package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	metav1 "crawlerd/pkg/meta/v1"
	jsoniter "github.com/json-iterator/go"
)

type httpClient struct {
	apiURL string
	http   *http.Client
}

func NewHTTPClient(addr string, http *http.Client) V1 {
	if strings.HasPrefix(addr, ":") {
		addr = "http://localhost:" + addr[1:]
	}
	c := &httpClient{
		apiURL: addr + "/v1",
		http:   http,
	}

	return c
}

func (c *httpClient) Run(r *metav1.RunnerUpCreate) (interface{}, error) {
	out := new(interface{})
	if err := c.post("/run", r, out); err != nil {
		return nil, err
	}

	return out, nil
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

	if outPtr != nil {
		return jsoniter.NewDecoder(resp.Body).Decode(outPtr)
	}

	return nil
}
