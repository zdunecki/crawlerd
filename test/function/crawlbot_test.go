package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/test"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestCrawlBot(t *testing.T) {
	type TestCase struct {
		Body   string   `yaml:"body"`
		Expect []string `yaml:"expect"`
	}

	var testCases []TestCase

	{
		b, _ := ioutil.ReadFile("./crawlbot_test.yaml")
		yaml.Unmarshal(b, &testCases)
	}

	linkToCrawlBotJS := "../../runners/crawlbot/index.js"
	crawlbotB, err := ioutil.ReadFile(linkToCrawlBotJS)
	if err != nil {
		t.Error(err)
		return
	}

	var functions = map[string]string{
		"test1": string(crawlbotB),
	}

	getFunction := func(c context.Context, id string) (string, error) {
		return functions[id], nil
	}

	api, store, done, err := testMongoDBAPI()
	defer done()

	handlerBody := "" // TODO: it's not a concurrent solution
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(handlerBody))
	})

	runner, fakeServerAddr, err := testRunner(getFunction, handler, store)
	if err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range testCases {
		handlerBody = testCase.Body

		{
			runID := "test1"
			_, err := runner.Run(&metav1.RunnerUpCreate{
				ID:  runID,
				URL: fakeServerAddr + "/some-url",
			})

			if err != nil {
				t.Error(err)
				return
			}
		}

		expectLinks := make([]string, len(testCase.Expect))

		for i, link := range testCase.Expect {
			if strings.HasPrefix(link, "$FAKE_SERVER_ADDR") {
				expectLinks[i] = strings.ReplaceAll(link, "$FAKE_SERVER_ADDR", fakeServerAddr)
			} else {
				expectLinks[i] = link
			}
		}

		linker := api.Linker()
		nodes, err := linker.All()

		if err != nil {
			t.Error(err)
			return
		}

		test.Diff(t, "nodes len should be equal", len(expectLinks), len(nodes))

		if nodes == nil || len(nodes) != len(expectLinks) {
			t.Errorf("invalid linker nodes len, expect=%d, but currently is=%d", len(expectLinks), len(nodes))
			return
		}

		nodeURLs := make([]string, len(nodes))
		for i, n := range nodes {
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
