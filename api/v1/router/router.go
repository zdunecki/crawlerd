package router

import (
	"context"

	"crawlerd/api"
	"crawlerd/api/v1/objects"
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/store"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

type router struct {
	store store.Repository

	scheduler crawlerdpb.SchedulerClient

	schedulerRetry func(func() error) error

	log *log.Entry

	storageBucketName string
	storageBucket     *blob.Bucket
}

func New(apiv1 api.API, store store.Repository, scheduler crawlerdpb.SchedulerClient, schedulerRetry func(func() error) error, log *log.Entry) error {
	r := &router{
		store:          store,
		scheduler:      scheduler,
		schedulerRetry: schedulerRetry,
		log:            log,
	}

	// TODO: storage abstraction
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("7E17SM0N1X3C7VTNPVV4", "NY2DH3G5W0Zn8Pdkp7W+IiR3oyYxLcRRqF1MNYW+", ""),
		Endpoint:         aws.String("http://172.22.0.2:9000"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	})
	if err != nil {
		log.Error(err)
	} else {
		bucketName := "crawlerd"
		bucket, err := s3blob.OpenBucket(context.TODO(), sess, bucketName, nil) // TODO: get bucket name from bucket?

		if err != nil {
			log.Error(err)
		}

		r.storageBucketName = bucketName
		r.storageBucket = bucket
	}

	v1Path := "/v1"
	requestQueuePath := v1Path + "/request-queue"
	urlsPath := v1Path + "/urls"
	linkerPath := v1Path + "/linker"
	jobsPath := v1Path + "/jobs"
	commandsPath := v1Path + "/cmd"

	// request queue
	apiv1.Post(requestQueuePath+"/batch", r.requestQueueBatchCreate)

	// seed
	apiv1.Get("/seed", r.seedList)
	apiv1.Post("/seed", r.seedAppend)
	apiv1.Delete("/seed/{id}", r.seedDelete)

	// urls
	apiv1.Get(urlsPath, r.urlsGetAll)
	apiv1.Get(urlsPath+"/{id}/history", r.urlsHistoryGet)
	apiv1.Post(urlsPath, r.urlsCreate, api.WithMaxBytes(objects.DefaultMaxPOSTContentLength))
	apiv1.Patch(urlsPath+"/{id}", r.urlsPatch, api.WithMaxBytes(objects.DefaultMaxPOSTContentLength))
	apiv1.Delete(urlsPath+"/{id}", r.urlsDelete)

	// TODO: bigcrawl endpoint
	// TODO: auth

	// BIG CRAWL ENDPOINTS
	apiv1.Post(commandsPath+"/apply", r.commandsApply)

	// jobs
	apiv1.Get(jobsPath, r.jobsGetAll)
	apiv1.Get(jobsPath+"/{id}", r.jobsGetByID)
	apiv1.Post(jobsPath, r.jobsCreate)
	apiv1.Patch(jobsPath+"/{id}", r.jobsPath)

	// linker
	apiv1.Get(linkerPath, r.linkerGetAll)

	return nil
}
