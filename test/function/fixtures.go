package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"crawlerd/api"
	v1 "crawlerd/api/v1"
	"crawlerd/api/v1/client"
	runnerv1 "crawlerd/pkg/runner/api/v1"
	"crawlerd/pkg/runner/testkit"
	"crawlerd/pkg/store"
	"github.com/go-chi/chi/v5"
	"github.com/orlangure/gnomock"
	mongopreset "github.com/orlangure/gnomock/preset/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func randomPort() string {
	min := 9890
	max := 9900
	return fmt.Sprintf("%d", rand.Intn(max-min)+min)
}

func testMongoDBAPI() (v1.V1, store.Repository, func(), error) {
	mgoPreset := mongopreset.Preset()
	mongoContainer, err := gnomock.Start(mgoPreset)

	done := func() {
		gnomock.Stop(mongoContainer)
	}

	if err != nil {
		done()
		return nil, nil, func() {

		}, err
	}

	appAddr := ":8080"
	schedulerGRPCAddr := ":" + randomPort()
	mongoDBName := "test"
	mongoURI := fmt.Sprintf("mongodb://%s", mongoContainer.DefaultAddress())

	apiV1, err := v1.New(
		v1.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
		v1.WithGRPCSchedulerServer(schedulerGRPCAddr),
	)

	if err != nil {
		done()
		return nil, nil, func() {

		}, err
	}

	go func() {
		r := chi.NewRouter()
		// TODO: cors config
		r.Use(func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")

				if r.Method == "OPTIONS" {
					return
				}

				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(fn)
		})

		if err := apiV1.Serve(appAddr, api.New(r)); err != nil {
			done()
		}
	}()

	time.Sleep(time.Second * 1)

	c, err := client.NewWithOpts(client.WithHTTPAddr("http://localhost" + appAddr))

	return c, apiV1.Store(), done, nil
}

func testRunner(getFunction testkit.GetFn, handler http.HandlerFunc, store store.Repository) (runnerv1.V1, string, error) {
	runnerStore := testkit.NewTestStore(getFunction, store)

	addr := ":7777"

	go func() {
		runnerv1.New(addr, runnerStore, runnerv1.Config{
			APIURL: "http://localhost:8080/v1",
		}).ListenAndServe()
	}()

	c := runnerv1.NewHTTPClient(addr, &http.Client{
		Timeout: time.Second * 60,
	})

	s := httptest.NewServer(handler)

	time.Sleep(time.Second)

	return c, s.URL, nil
}
