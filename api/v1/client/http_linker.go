package client

import (
	metav1 "crawlerd/pkg/meta/v1"
)

type httpLinker struct {
	client *httpClient
}

func (c *httpLinker) All() ([]*metav1.LinkNode, error) {
	resp := make([]*metav1.LinkNode, 0)

	if err := c.client.get("/linker", &resp); err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp, nil
}
