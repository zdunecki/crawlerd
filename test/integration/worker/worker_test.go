package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func post(url string, interval int) error {
	var (
		apiHost = os.Getenv("API_HOST")
	)
	if apiHost == "" {
		apiHost = "api:8080"
	}

	data := map[string]interface{}{
		"url":      url,
		"interval": interval,
	}
	dataB, err := json.Marshal(data)
	if err != nil {
		return err
	}
	{
		_, err := http.Post(fmt.Sprintf("http://%s/api/urls", apiHost), "application/json", bytes.NewReader(dataB))
		if err != nil {
			return nil
		}
	}

	return nil
}

// TODO: multiple tests, multi worker tests
func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	{
		work, err := newFixtures()
		if err != nil {
			t.Error(err)
		}

		go func() {
			if err := work.Serve(ctx); err != nil {
				t.Error(err)
			}
		}()
	}

	timeN := 60
	t1 := 10

	if err := post("https://httpbin.org/range/1", t1); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second*time.Duration(timeN) + time.Second*3)
	cancel()

	storage, err := newFixtureStorage()

	if err != nil {
		t.Log(err)
	}

	{
		history, err := storage.History().FindByID(context.Background(), 0)
		if err != nil {
			t.Error(err)
		}

		n := timeN / t1
		if len(history) < n {
			t.Errorf("should be crawled at least %d times", n)
		}
	}

	t.Log("finished")
}
