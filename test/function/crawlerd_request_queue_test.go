package main

import (
	"testing"

	metav1 "crawlerd/pkg/meta/metav1"
)

func TestRequestQueueCreateAPIEndpoint(t *testing.T) {
	apiv1, _, _, done, err := testMongoDBAPI()
	defer done()

	if err != nil {
		t.Error(err)
		return
	}

	rq := apiv1.RequestQueue()

	// TODO: test linker - should have be inserted once

	{
		data := []*metav1.RequestQueueCreate{
			{
				RunID: "123",
				URL:   "https://example.com",
			},
		}

		resp, err := rq.BatchCreate(data)
		if err != nil {
			t.Error(err)
			return
		}

		if resp == nil || len(resp.IDs) != 1 {
			t.Error("batch create ids len should be equal 1")
			return
		}
	}

	{
		data := []*metav1.RequestQueueCreate{
			{
				RunID: "123",
				URL:   "https://example.com",
			},
		}

		resp, err := rq.BatchCreate(data)
		if err != nil {
			t.Error(err)
			return
		}

		if resp == nil || len(resp.IDs) != 1 {
			t.Error("batch create ids len should be equal 1")
			return
		}
	}
}
