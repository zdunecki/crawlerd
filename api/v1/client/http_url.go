package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	v1 "crawlerd/api/v1"
	metav1 "crawlerd/pkg/meta/v1"
	jsoniter "github.com/json-iterator/go"
)

type httpURL struct {
	apiURL     string
	httpClient *http.Client
}

func (h *httpURL) Create(url *v1.RequestPostURL) (*v1.ResponsePostURL, error) {
	resp := &v1.ResponsePostURL{}

	if err := h.post("", url, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *httpURL) Patch(id string, data *v1.RequestPatchURL) (*v1.ResponsePostURL, error) {
	resp := &v1.ResponsePostURL{}

	if err := h.patch("/"+id, data, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *httpURL) Delete(id string) error {
	return h.delete("/"+id, nil, nil)
}

// TODO: scroll
func (h *httpURL) All() ([]*metav1.URL, error) {
	resp := make([]*metav1.URL, 0)

	if err := h.get("", &resp); err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp, nil
}

// TODO: scroll
func (h *httpURL) History(urlID string) ([]*metav1.History, error) {
	resp := make([]*metav1.History, 0)

	if err := h.get(fmt.Sprintf("/%s/history", urlID), &resp); err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp, nil
}

func (h *httpURL) get(resource string, outPtr interface{}) error {
	resp, err := h.httpClient.Get(h.apiURL + resource)
	if err != nil {
		return err
	}

	return jsoniter.NewDecoder(resp.Body).Decode(outPtr)
}

func (h *httpURL) post(resource string, body interface{}, outPtr interface{}) error {
	return h.request(http.MethodPost, resource, body, outPtr)
}

func (h *httpURL) patch(resource string, body interface{}, outPtr interface{}) error {
	return h.request(http.MethodPatch, resource, body, outPtr)
}

func (h *httpURL) put(resource string, body interface{}, outPtr interface{}) error {
	return h.request(http.MethodPut, resource, body, outPtr)
}

func (h *httpURL) delete(resource string, body interface{}, outPtr interface{}) error {
	return h.request(http.MethodDelete, resource, body, outPtr)
}

func (h *httpURL) request(method, resource string, body interface{}, outPtr interface{}) error {
	var bodyReader *bytes.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}

		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, h.apiURL+resource, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application-json")
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}

	if outPtr != nil {
		return jsoniter.NewDecoder(resp.Body).Decode(outPtr)
	}

	return nil
}

// TODO: timeout
func newHTTPURL(apiAddr string) v1.V1URL {
	return &httpURL{
		apiURL: apiAddr + "/api/urls",
		httpClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}
