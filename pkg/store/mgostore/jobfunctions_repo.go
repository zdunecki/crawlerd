package mgostore

import (
	"context"
	"errors"
	"path"

	"crawlerd/pkg/store"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob/s3blob"
)

type jobFunctions struct {
	session *session.Session
	job     store.Job
}

// TODO: refactor
func NewJobFunctions(jobRepo store.Job) *jobFunctions {
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("7E17SM0N1X3C7VTNPVV4", "NY2DH3G5W0Zn8Pdkp7W+IiR3oyYxLcRRqF1MNYW+", ""),
		Endpoint:         aws.String("http://172.22.0.2:9000"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String("us-east-1"),
	})

	return &jobFunctions{session: sess, job: jobRepo}
}

func (jf *jobFunctions) GetByID(ctx context.Context, jobID string) (string, error) {
	j, err := jf.job.FindOneByID(ctx, jobID)
	if err != nil {
		return "", err
	}

	if j == nil {
		return "", errors.New("not found")
	}

	bucketName, objectName := path.Split(j.JavaScriptBundleSrc)
	bucket, err := s3blob.OpenBucket(context.TODO(), jf.session, bucketName, nil) // TODO: cache

	content, err := bucket.ReadAll(ctx, objectName)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
