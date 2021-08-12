package main

import (
	"fmt"
	"math/rand"
	"time"

	"crawlerd/api"
	v1 "crawlerd/api/v1"
	"crawlerd/api/v1/client"
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

func testMongoDBAPI() (v1.V1, func(), error) {
	mgoPreset := mongopreset.Preset()
	mongoContainer, err := gnomock.Start(mgoPreset)

	done := func() {
		gnomock.Stop(mongoContainer)
	}

	if err != nil {
		done()
		return nil, func() {

		}, err
	}

	appAddr := ":6666"
	schedulerGRPCAddr := ":" + randomPort()
	mongoDBName := "test"
	mongoURI := fmt.Sprintf("mongodb://%s", mongoContainer.DefaultAddress())

	apiV1, err := v1.New(
		v1.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
		v1.WithGRPCSchedulerServer(schedulerGRPCAddr),
	)

	if err != nil {
		done()
		return nil, func() {

		}, err
	}

	go func() {
		if err := apiV1.Serve(appAddr, api.New(chi.NewRouter())); err != nil {
			done()
		}
	}()

	time.Sleep(time.Second * 1)

	c, err := client.NewWithOpts(client.WithHTTP("http://localhost" + appAddr))

	return c, done, nil
}
