package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	metav1 "crawlerd/pkg/meta/v1"
)

// TODO: esbuild tests
// TODO: fix tests after refactor
func TestRunner(t *testing.T) {
	var functions = map[string]string{
		"test1": `
		(() => {
			return {
				message: "hello world"
			}
		})()
	`,
	}

	var expect = map[string]string{
		"test1": "hello world",
	}

	getFunction := func(c context.Context, id string) (string, error) {
		return functions[id], nil
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	getFunction(nil, "")

	c, addr, err := testRunner(handler, nil)
	if err != nil {
		t.Error(err)
		return
	}

	runID := "test1"
	out, err := c.Run(&metav1.RunnerUpCreate{
		ID:  runID,
		URL: addr + "/some-url",
	})

	if err != nil {
		t.Error(err)
		return
	}

	type output struct {
		Message string `json:"message"`
	}

	b, _ := json.Marshal(out)

	var o output
	if err := json.Unmarshal(b, &o); err != nil {
		t.Error(err)
		return
	}

	shouldBe := expect[runID]
	if o.Message != shouldBe {
		t.Errorf("should be: %s, but currently is: %s", shouldBe, o.Message)
	}
}
