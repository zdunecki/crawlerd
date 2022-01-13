package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	metav1 "crawlerd/pkg/meta/metav1"
	"crawlerd/pkg/runner/testkit"
	"crawlerd/test"
	"gopkg.in/h2non/gock.v1"
)

func TestCrawlBot(t *testing.T) {
	defer gock.Off()

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

	rfs := testkit.NewTestRunnerFunctions(getFunction)
	storeOptions.CustomRunnerFunctions(rfs).Apply()

	runner, err := testRunner(store)
	if err != nil {
		t.Error(err)
		return
	}

	type fakeServer struct {
		Handler http.HandlerFunc
		Bodies  map[string]string
	}

	fakeServers := make(map[string]*fakeServer)

	fakeServerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if server, ok := fakeServers[r.Host]; !ok {
			w.Write([]byte("no server"))
		} else {
			host := r.Host
			path := r.RequestURI
			if path == "/" {
				path = ""
			}
			if body, ok := server.Bodies["http://"+host+path]; ok {
				w.Write([]byte(body))
				return
			}
			w.Write([]byte("no body"))
		}
	})

	fakeRootServer := httptest.NewServer(fakeServerHandler)
	fakeRootServerURL, err := url.Parse(fakeRootServer.URL)
	if err != nil {
		t.Error(err)
		return
	}
	fakeServers[fakeRootServerURL.Host] = &fakeServer{
		Handler: fakeServerHandler,
		Bodies:  make(map[string]string),
	}

	props := &CrawlBotTestProps{
		RootServer: fakeRootServer.URL,
	}

	fmt.Println(fakeRootServerURL.Host)

	if err := jsxData("crawlbot_test.jsx", props, &testCases); err != nil {
		t.Error(err)
		return
	}

	for _, testCase := range testCases {
		t.Log(testCase.Description)

		for u, page := range testCase.Pages {
			pageURL, err := url.Parse(u)
			if err != nil {
				t.Error(err)
				return
			}

			// TODO: http mock
			if _, ok := fakeServers[pageURL.Host]; !ok {
				if page.OutsideNetwork {
					//server, ok := fakeServers[pageURL.Host]
					//
					if !ok {
						fakeServers[pageURL.Host] = &fakeServer{
							Handler: fakeServerHandler,
							Bodies:  make(map[string]string),
						}
					}

					// TODO: for some reason gock mock also another requests (request to runner)
					//gock.New(pageURL.Scheme + "//" + pageURL.Host).
					//	Get(pageURL.Path).
					//	Filter(func(request *http.Request) bool { // avoid issues with mocking outside requests
					//		return server != nil
					//	}).
					//	Map(func(r *http.Request) *http.Request {
					//		if body, ok := server.Bodies[r.URL.String()]; ok {
					//			r.Response.Body = ioutil.NopCloser(strings.NewReader(body))
					//		}
					//
					//		return r
					//	})

				} else {
					l, err := net.Listen("tcp", pageURL.Host)
					if err != nil {
						t.Error(err)
						return
					}

					ts := httptest.NewUnstartedServer(fakeServerHandler)
					ts.Listener.Close()
					ts.Listener = l
					ts.Start()

					fakeServers[pageURL.Host] = &fakeServer{
						Handler: fakeServerHandler,
						Bodies:  make(map[string]string),
					}
				}
			}

			pURL := pageURL.String()
			server := fakeServers[pageURL.Host]

			server.Bodies[pURL] = page.Body
		}

		{
			var followLinks []*metav1.StringFilter

			if testCase.FollowLinks != nil {
				followLinks = make([]*metav1.StringFilter, 0)
				for _, f := range testCase.FollowLinks {
					//re, err := regexp.Compile(f.Match)
					//if err != nil {
					//	t.Error(err)
					//	return
					//}

					followLinks = append(followLinks, &metav1.StringFilter{
						Is:    f.Is,
						Match: f.Match,
					})
				}
			}
			_, err := runner.Run(&metav1.RunnerUpCreate{
				ID:                 runID,
				URL:                testCase.StartURL,
				MaxDepth:           testCase.MaxDepth,
				ScrapeLinksPattern: testCase.ScrapeLinksPattern,
				FollowLinks:        followLinks,
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

		if linkerNodes != nil && (len(linkerNodes) != len(expectLinks)) {
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
