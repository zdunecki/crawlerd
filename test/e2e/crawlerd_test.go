package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"crawlerd/pkg/storage/objects"
)

func TestE2ECrawlerd(t *testing.T) {
	var (
		apiHost = os.Getenv("API_HOST")
	)

	if apiHost == "" {
		apiHost = "api:8080"
	}

	data := map[string]interface{}{
		"url":      "https://httpbin.org/range/1",
		"interval": 10,
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	{
		_, err := http.Post(fmt.Sprintf("http://%s/api/urls", apiHost), "application/json", bytes.NewReader(dataB))
		if err != nil {
			t.Error(err)
		}
	}

	time.Sleep(time.Minute + time.Second * 3)

	{
		var history []objects.History

		res, err := http.Get(fmt.Sprintf("http://%s/api/urls/0/history", apiHost))
		if err != nil {
			t.Error(err)
		}

		if err := json.NewDecoder(res.Body).Decode(&history); err != nil {
			t.Error(err)
		}

		if len(history) < 6 {
			t.Error("should be crawled at least 6 times")
		}
	}
}
