package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"crawlerd/api"
	"crawlerd/api/v1"
	"crawlerd/api/v1/sdk"
	"crawlerd/pkg/runner/api/runnerv1"
	"crawlerd/pkg/store"
	storeOptions "crawlerd/pkg/store/options"
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

func testMongoDBAPI() (v1.V1, store.Repository, *storeOptions.RepositoryOption, func(), error) {
	mgoPreset := mongopreset.Preset()
	mongoContainer, err := gnomock.Start(mgoPreset)

	done := func() {
		gnomock.Stop(mongoContainer)
	}

	if err != nil {
		done()
		return nil, nil, nil, func() {

		}, err
	}

	appAddr := ":8080"
	schedulerGRPCAddr := ":" + randomPort()
	mongoDBName := "test"
	mongoURI := fmt.Sprintf("mongodb://%s", mongoContainer.DefaultAddress())

	storeOptions := storeOptions.
		Client().
		WithMongoDB(mongoDBName, options.Client().ApplyURI(mongoURI))

	apiV1, err := v1.New(
		v1.WithStore(storeOptions),
		v1.WithGRPCSchedulerServer(schedulerGRPCAddr),
	)

	if err != nil {
		done()
		return nil, nil, nil, func() {

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

	c, err := sdk.NewWithOpts(sdk.WithHTTPAddr("http://localhost" + appAddr))

	return c, apiV1.Store(), storeOptions, done, nil
}

func testRunner(store store.Repository) (runnerv1.V1, error) {
	addr := ":7777"

	go func() {
		runnerv1.New(addr, store, runnerv1.Config{
			APIURL: "http://localhost:8080/v1",
		}).ListenAndServe()
	}()

	c := runnerv1.NewHTTPClient(addr, &http.Client{
		Timeout: time.Second * 60,
	})

	time.Sleep(time.Second)

	return c, nil
}
