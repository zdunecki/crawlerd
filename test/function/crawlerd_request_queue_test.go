package main

import (
	"testing"

	metav1 "crawlerd/pkg/meta/v1"
)

func TestRequestQueueCreateAPIEndpoint(t *testing.T) {
	clientv1, done, err := testMongoDBAPI()
	defer done()

	if err != nil {
		t.Error(err)
	}

	rq := clientv1.RequestQueue()

	// TODO: test linker - should have be inserted once

	{
		data := []*metav1.RequestQueueCreate{
			{
				URL: "https://example.com",
			},
		}

		resp, err := rq.BatchCreate(data)
		if err != nil {
			t.Error(err)
		}

		if resp == nil || len(resp.IDs) != 1 {
			t.Error("batch create ids len should be equal 1")
		}
	}

	{
		data := []*metav1.RequestQueueCreate{
			{
				URL: "https://example.com",
			},
		}

		resp, err := rq.BatchCreate(data)
		if err != nil {
			t.Error(err)
		}

		if resp == nil || len(resp.IDs) != 1 {
			t.Error("batch create ids len should be equal 1")
		}
	}
}
