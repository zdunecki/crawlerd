package sdk

import (
	v1 "crawlerd/api/v1"
	metav1 "crawlerd/pkg/meta/metav1"
)

type httpRequestQueue struct {
	client *httpClient
}

func (c *httpRequestQueue) BatchCreate(body []*metav1.RequestQueueCreate) (*v1.ResponseRequestQueueCreate, error) {
	resp := &v1.ResponseRequestQueueCreate{}

	if err := c.client.post("/request-queue/batch", body, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
