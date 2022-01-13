package sdk

import (
	"fmt"

	v1 "crawlerd/api/v1"
	metav1 "crawlerd/pkg/meta/metav1"
)

// TODO: delete all

type httpURL struct {
	rest rest
}

func (c *httpURL) Create(url *v1.RequestPostURL) (*v1.ResponsePostURL, error) {
	resp := &v1.ResponsePostURL{}

	if err := c.rest.post("", url, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *httpURL) Patch(id string, data *v1.RequestPatchURL) (*v1.ResponsePostURL, error) {
	resp := &v1.ResponsePostURL{}

	if err := c.rest.patch("/"+id, data, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *httpURL) Delete(id string) error {
	return c.rest.delete("/"+id, nil, nil)
}

// TODO: scroll
func (c *httpURL) All() ([]*metav1.URL, error) {
	resp := make([]*metav1.URL, 0)

	if err := c.rest.get("", &resp); err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp, nil
}

// TODO: scroll
func (c *httpURL) History(urlID string) ([]*metav1.History, error) {
	resp := make([]*metav1.History, 0)

	if err := c.rest.get(fmt.Sprintf("/%s/history", urlID), &resp); err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, nil
	}

	return resp, nil
}
