package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	"crawlerd/pkg/storage/objects"
	"crawlerd/test"
)

func TestE2ECrawlerd(t *testing.T) {
	var (
		apiHost = os.Getenv("API_HOST")
	)

	if apiHost == "" {
		apiHost = "api:8080"
	}

	interval := 10

	data := map[string]interface{}{
		"url":      "https://httpbin.org/range/1",
		"interval": interval,
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

	firstRunDontRunInterval := time.Second * time.Duration(interval)
	wait := time.Minute - firstRunDontRunInterval

	cycles := int(math.Round(wait.Seconds() / float64(interval)))

	time.Sleep(wait)

	{
		var history []objects.History

		res, err := http.Get(fmt.Sprintf("http://%s/api/urls/0/history", apiHost))
		if err != nil {
			t.Error(err)
		}

		if err := json.NewDecoder(res.Body).Decode(&history); err != nil {
			t.Error(err)
		}

		test.Diff(t, "should crawl n times", cycles, len(history))
	}
}
