package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/runner/testkit"
	"crawlerd/test"
)

func TestCrawlBot(t *testing.T) {
	var testCases []CrawlBotTestCase

	runID := "test1"

	crawlbotB, err := ioutil.ReadFile("../../runners/crawlbot/index.js")
	if err != nil {
		t.Error(err)
		return
	}

	var functions = map[string]string{
		runID: string(crawlbotB),
	}

	getFunction := func(c context.Context, id string) (string, error) {
		return functions[id], nil
	}

	api, store, storeOptions, done, err := testMongoDBAPI()
	defer done()

	handlerBody := "" // TODO: it's not a concurrent solution
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(handlerBody))
	})

	rf := testkit.NewTestRunnerFunctions(getFunction)
	storeOptions.CustomRunnerFunctions(rf).Apply()

	runner, fakeServerURL, err := testRunner(handler, store)
	if err != nil {
		t.Error(err)
		return
	}

	props := map[string]string{
		"fakeServer": fakeServerURL,
	}
	if err := jsxData("crawlbot_test.jsx", props, &testCases); err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range testCases {
		handlerBody = testCase.Body

		{
			_, err := runner.Run(&metav1.RunnerUpCreate{
				ID:       runID,
				URL:      props["fakeServer"] + "/some-url",
				MaxDepth: testCase.MaxDepth,
			})

			if err != nil {
				t.Error(err)
				return
			}
		}

		expectLinks := make([]string, len(testCase.Expect))

		for i, link := range testCase.Expect {
			expectLinks[i] = link
		}

		linker := api.Linker()
		linkerNodes, err := linker.All()

		if err != nil {
			t.Error(err)
			return
		}

		test.Diff(t, "linkerNodes len should be equal", len(expectLinks), len(linkerNodes))

		if linkerNodes == nil || len(linkerNodes) != len(expectLinks) {
			t.Errorf("invalid linker linkerNodes len, expect=%d, but currently is=%d", len(expectLinks), len(linkerNodes))
			return
		}

		nodeURLs := make([]string, len(linkerNodes))
		for i, n := range linkerNodes {
			nodeURLs[i] = fmt.Sprintf("%s", n.URL)
		}

		test.Diff(t, "expect links to be", expectLinks, nodeURLs)

		runners, err := store.Runner().List(context.Background())
		if err != nil {
			t.Error(err)
			return
		}

		for _, runner := range runners {
			test.Diff(t, "runner status should be equal", metav1.RunnerStatusSuccessed, runner.Status)
		}
	}
}
