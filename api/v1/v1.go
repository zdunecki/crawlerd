package v1

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"crawlerd/api"
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cenkalti/backoff/v3"
	esbuild "github.com/evanw/esbuild/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const (
	IntervalMinValue = 5
)

type V1URL interface {
	Create(*RequestPostURL) (*ResponsePostURL, error)
	Patch(id string, data *RequestPatchURL) (*ResponsePostURL, error)
	Delete(id string) error
	All() ([]*objects.URL, error)
	History(urlID string) ([]*objects.History, error)
}

type V1 interface {
	URL() V1URL
}

type v1 struct {
	storage   storage.Storage
	scheduler crawlerdpb.SchedulerClient

	schedulerBackoff *backoff.ExponentialBackOff

	log *log.Entry
}

func createDirIfNotExists(target string) error {
	if _, err := os.Stat(target); err != nil {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	return nil
}

// i.e {"path-to-file-1": "content", "path-to-file-2": "content2"}`

type VirtualFiles map[string][]byte

// i.e `"folder": {"path-to-file-1": "content", "path-to-file-2": "content2"}`

type VirtualFilePath map[string]VirtualFiles

func Unpack(bucket *blob.Bucket, dstDir string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	configFiles := make(map[string]*viper.Viper)

	content2 := make(VirtualFilePath)

	// TODO: not in defer
	// TODO: delete dstDir after bundle
	// TODO: send bundle file to storage
	// TODO: error handling
	// TODO: parse env i.e $TOKEN in config files

	defer func() {
		for _, cfg := range configFiles {
			jobID := cfg.GetString("job_id")
			entry := cfg.GetString("entry")

			if jobID == "" { // omit root
				continue
			}

			if entry == "" {
				entry = "index.js" // TODO: index.ts
			}

			// TODO: build in-memory instead of fs
			result := esbuild.Build(esbuild.BuildOptions{
				EntryPoints: []string{path.Join(dstDir, jobID, entry)},
				Outfile:     path.Join(dstDir, jobID, "output.js"),
				Bundle:      true,
				Write:       true,
				LogLevel:    esbuild.LogLevelInfo,
			})

			if len(result.Errors) > 0 {
				log.Error(result.Errors)
			}
		}

	}()

	//filesPerJob := make(map[string]*BuildPath)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			for dir, fileMap := range content2 {
				findConfig := func() (*viper.Viper, string) {
					if _, ok := configFiles[dir]; ok {
						return configFiles[dir], dir
					} else {
						movementDir := dir

						for {
							if movementDir == "" || movementDir == "." {
								break
							}
							movementDir = filepath.Join(movementDir, "../") + "/"
							if _, ok := configFiles[movementDir]; ok {
								return configFiles[movementDir], movementDir
							}
						}
					}

					return nil, ""
				}

				cfg, _ := findConfig()

				if cfg == nil {
					continue
				}
				jobID := cfg.GetString("job_id")
				//entry := cfg.GetString("entry")

				// TODO: insert files to API based on config
				// TODO: bundle files per job
				for filePath, fileB := range fileMap {
					//filesPerJob[jobID].Paths[dir][filePath] = fileB
					if err := bucket.WriteAll(context.TODO(), filePath, fileB, nil); err != nil {
						log.Error(err)
					}

					_, closestConfigPath := findConfig()

					// root out is in closest config level
					targetFile := filepath.Join(dstDir, jobID, strings.ReplaceAll(filePath, closestConfigPath, ""))
					targetDir, _ := path.Split(targetFile)
					if err := createDirIfNotExists(targetDir); err != nil {
						log.Error(err)
						continue
					}

					f, err := os.Create(targetFile)
					if err != nil {
						log.Error(err)
						continue
					}
					f.Close()

					if _, err := io.Copy(f, bytes.NewReader(fileB)); err != nil {
						log.Error(err)
						continue
					}

				}
			}
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		//target := filepath.Join(dst, header.Name)
		targetDir, targetFile := path.Split(header.Name)

		// TODO: support json etc.
		if targetFile == "bigcrawl.yaml" {
			b, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(bytes.NewReader(b)); err != nil {
				panic(err)
			}

			configFiles[targetDir] = v
		}

		// check the file type
		switch header.Typeflag {

		case tar.TypeDir:

		case tar.TypeReg:
			b, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}

			if _, ok := content2[targetDir]; !ok {
				content2[targetDir] = make(map[string][]byte)
			}
			content2[targetDir][header.Name] = b
		}
	}
}

func New(opts ...Option) (*v1, error) {
	v := &v1{
		log: log.WithFields(map[string]interface{}{
			"service": "api",
		}),
	}

	for _, o := range opts {
		if err := o(v); err != nil {
			return nil, err
		}
	}

	{
		bo := backoff.NewExponentialBackOff()
		bo.MaxInterval = time.Second * 2
		bo.MaxElapsedTime = time.Second * 15

		v.schedulerBackoff = bo
	}

	return v, nil
}

func (v *v1) Serve(addr string, v1 api.API) error {
	if v.storage == nil {
		return ErrNoStorage
	}

	if v.scheduler == nil {
		return ErrNoScheduler
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
		v.log.Error(err)
		return err
	}

	bucket, err := s3blob.OpenBucket(context.TODO(), sess, "crawlerd", nil)
	if err != nil {
		v.log.Error(err)
		return err
	}

	v1.Post("/api/urls", func(ctx api.Context) {
		var req *RequestPostURL

		data, err := ioutil.ReadAll(ctx.Request().Body)
		if data != nil && len(data) >= DefaultMaxPOSTContentLength.Int() {
			ctx.RequestEntityTooLarge()
			return
		}

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if req.Interval < IntervalMinValue {
			ctx.BadRequest()
			return
		}

		done, seq, err := v.storage.URL().InsertOne(ctx.RequestContext(), req.URL, req.Interval)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.BadRequest()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.AddURL(ctx.RequestContext(), &crawlerdpb.RequestURL{
				Id:       int64(seq),
				Url:      req.URL,
				Interval: int64(req.Interval),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.Created().JSON(&ResponsePostURL{
			ID: seq,
		})
	}, api.WithMaxBytes(DefaultMaxPOSTContentLength))

	v1.Patch("/api/urls/{id}", func(ctx api.Context) {
		var req RequestPatchURL

		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		data, err := ioutil.ReadAll(ctx.Request().Body)
		if data != nil && len(data) >= DefaultMaxPOSTContentLength.Int() {
			ctx.RequestEntityTooLarge()
			return
		}

		if err != nil {
			log.Error(err)
			ctx.InternalError()
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.storage.URL().UpdateOneByID(ctx.RequestContext(), id, RequestPatchURL{
			URL:      req.URL,
			Interval: req.Interval,
		})
		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.NotFound()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.UpdateURL(ctx.RequestContext(), &crawlerdpb.RequestURL{
				Id:       int64(id),
				Url:      *req.URL,
				Interval: int64(*req.Interval),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.JSON(&ResponsePostURL{
			ID: id,
		})
	}, api.WithMaxBytes(DefaultMaxPOSTContentLength))

	v1.Delete("/api/urls/{id}", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.storage.URL().DeleteOneByID(ctx.RequestContext(), id)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if !done {
			ctx.NotFound()
			return
		}

		if err := v.schedulerRetry(func() error {
			_, e := v.scheduler.DeleteURL(ctx.RequestContext(), &crawlerdpb.RequestDeleteURL{
				Id: int64(id),
			})
			return e
		}); err != nil {
			v.log.Error(err)
		}

		ctx.NoContent()
	})

	v1.Get("/api/urls", func(ctx api.Context) {
		urls, err := v.storage.URL().FindAll(ctx.RequestContext())
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if urls == nil {
			urls = []objects.URL{}
		}

		ctx.JSON(urls)
	})

	v1.Get("/api/urls/{id}/history", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		history, err := v.storage.History().FindByID(ctx.RequestContext(), id)
		if err != nil {
			v.log.Error(err)
			ctx.InternalError()
			return
		}

		if history == nil {
			history = []objects.History{}
		}

		ctx.JSON(history)
	})

	// TODO: bigcrawl endpoints
	// TODO: auth
	v1.Post("/v1/cmd/apply", func(ctx api.Context) {
		//r, err := gzip.NewReader(ctx.Request().Body)
		wd, err := os.Getwd()
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON("something went wrong")
			return
		}

		err = Unpack(bucket, path.Join(wd, "OUTPUT"), ctx.Request().Body)
		//err = Unpack(bucket, ctx.Request().Body)
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON("something went wrong")
			return
		}

		ctx.JSON("abc")
	})

	v.log.Info("listening on: ", addr)
	return http.ListenAndServe(addr, v1.Handler())
}

func (v *v1) schedulerRetry(f func() error) error {
	return backoff.Retry(f, v.schedulerBackoff)
}
