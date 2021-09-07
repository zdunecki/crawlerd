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
	"regexp"
	"strings"
	"time"

	"crawlerd/api"
	"crawlerd/crawlerdpb"
	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
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

// TODO: routes
type v1 struct {
	store store.Repository

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

func Unpack(bucketName string, bucket *blob.Bucket, dstDir string, r io.Reader, jobRepo store.Job) error {
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
	// TODO: send bundle file to store
	// TODO: error handling
	// TODO: parse env i.e $TOKEN in config files

	// TODO: add source code files into real codespace id
	exampleCodeSpaceID := "jg3cekg3x"

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

			// TODO: configure output.js
			// TODO: check if not conflict with file existing in this fs level
			// TODO: send into another place? - separate from source code
			jobEntryPoint := path.Join(dstDir, jobID)
			outFilePath := path.Join(jobEntryPoint, "output.js")

			// TODO: build in-memory instead of fs
			result := esbuild.Build(esbuild.BuildOptions{
				EntryPoints: []string{path.Join(jobEntryPoint, entry)},
				Format:      esbuild.FormatCommonJS,
				Outfile:     outFilePath,
				Bundle:      true,
				Write:       false, // TODO: Write: false
				LogLevel:    esbuild.LogLevelInfo,
			})

			if len(result.Errors) > 0 {
				log.Error(result.Errors)
				continue
			}

			//jobID, _ := filepath.Split(fullJobPath)

			// TODO: when OutputFiles can be > 1 ?
			object := "bundles/jobs/" + jobID + "/output.js"
			if err := bucket.WriteAll(context.TODO(), "bundles/jobs/"+jobID+"/output.js", result.OutputFiles[0].Contents, nil); err != nil {
				log.Error(err)
				continue
			}

			jobRepo.PatchOneByID(context.TODO(), jobID, &metav1.JobPatch{
				JavaScriptBundleSrc: path.Join(bucketName, object),
			})
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
					if err := bucket.WriteAll(context.TODO(), path.Join("codespaces", exampleCodeSpaceID, filePath), fileB, nil); err != nil {
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
					//f.Close() TODO: don't close folder file

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
	if v.store == nil {
		return ErrNoStorage
	}

	if v.scheduler == nil {
		return ErrNoScheduler
	}

	// TODO: store abstraction
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

	bucketName := "crawlerd"
	bucket, err := s3blob.OpenBucket(context.TODO(), sess, bucketName, nil) // TODO: get bucket name from bucket?

	if err != nil {
		v.log.Error(err)
		return err
	}

	// TODO urls are now queues
	// TODO: auth
	// TODO: batch errors
	// TODO: ScrapeLinksPattern, FollowLinks on frontend side
	v1.Post("/v1/request-queue/batch", func(c api.Context) {
		var req []*metav1.RequestQueueCreateAPI
		rq := make([]*metav1.RequestQueueCreate, 0)

		rs := v.store.Runner()

		// TODO: body limitations?
		data, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			c.BadRequest()
			return
		}

		{
			linkNodes := make([]*metav1.LinkNodeCreate, 0)

			for _, r := range req {
				if err := r.Validate(); err != nil {
					v.log.Error(err)
					c.BadRequest().JSON(&APIError{
						Type:    ErrorTypeValidation,
						Message: err.Error(),
					})
					return
				}

				runner, err := rs.GetByID(c.RequestContext(), r.RunID)
				if err != nil {
					c.InternalError().JSON(&APIError{
						Type: ErrorTypeInternal,
					})
					return
				}

				addLinkNode := func() {
					linkNodes = append(linkNodes, &metav1.LinkNodeCreate{
						URL: metav1.NewLinkURL(r.URL),
					})
				}

				if runner.RunnerConfig.ScrapeLinksPattern != "" {
					if re, err := regexp.Compile(runner.RunnerConfig.ScrapeLinksPattern); err != nil {
						v.log.Error(err)
					} else {
						if re.Match([]byte(r.URL)) {
							addLinkNode()
						}
					}
				} else {
					addLinkNode()
				}

				addRequestQue := func() {
					rq = append(rq, &metav1.RequestQueueCreate{
						RunID:  r.RunID,
						URL:    r.URL,
						Depth:  r.Depth,
						Status: metav1.RequestQueueStatusQueued,
					})
				}

				if runner.RunnerConfig.FollowLinks == nil {
					addRequestQue()
				} else {
					shouldAddRq := false
					for _, filter := range runner.RunnerConfig.FollowLinks {
						if filter.Match != "" {
							re, _ := regexp.Compile(filter.Match)
							if re.MatchString(r.URL) {
								shouldAddRq = true
								break
							}
						} else if filter.Is != "" {
							if r.URL == filter.Is {
								shouldAddRq = true
								break
							}
						}
					}

					if shouldAddRq {
						addRequestQue()
					}
				}

			}

			_, err := v.store.Linker().InsertManyIfNotExists(c.RequestContext(), linkNodes)
			if err != nil {
				v.log.Error(err)
				c.InternalError().JSON(&APIError{
					Type: ErrorTypeInternal,
				})
				return
			}
		}

		ids, err := v.store.RequestQueue().InsertMany(c.RequestContext(), rq)
		if err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		c.Created().JSON(&ResponseRequestQueueCreate{
			IDs: ids,
		})
	})

	v1.Post("/v1/urls", func(ctx api.Context) {
		var req *RequestPostURL

		data, err := ioutil.ReadAll(ctx.Request().Body)
		if data != nil && len(data) >= DefaultMaxPOSTContentLength.Int() {
			ctx.RequestEntityTooLarge()
			return
		}

		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
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

		done, seq, err := v.store.URL().InsertOne(ctx.RequestContext(), req.URL, req.Interval)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
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

	v1.Patch("/v1/urls/{id}", func(ctx api.Context) {
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
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		if err := json.NewDecoder(bytes.NewReader(data)).Decode(&req); err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.store.URL().UpdateOneByID(ctx.RequestContext(), id, RequestPatchURL{
			URL:      req.URL,
			Interval: req.Interval,
		})
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
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

	v1.Delete("/v1/urls/{id}", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		done, err := v.store.URL().DeleteOneByID(ctx.RequestContext(), id)

		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
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

	v1.Get("/v1/urls", func(ctx api.Context) {
		urls, err := v.store.URL().FindAll(ctx.RequestContext())
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if urls == nil {
			urls = []metav1.URL{}
		}

		ctx.JSON(urls)
	})

	v1.Get("/v1/urls/{id}/history", func(ctx api.Context) {
		id, err := ctx.ParamInt("id")
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		history, err := v.store.History().FindByID(ctx.RequestContext(), id)
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		if history == nil {
			history = []metav1.History{}
		}

		ctx.JSON(history)
	})

	// TODO: bigcrawl endpoint
	// TODO: auth

	// BIG CRAWL ENDPOINTS

	v1.Post("/v1/cmd/apply", func(ctx api.Context) {
		//r, err := gzip.NewReader(ctx.Request().Body)
		wd, err := os.Getwd()
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		err = Unpack(bucketName, bucket, path.Join(wd, "OUTPUT"), ctx.Request().Body, v.store.Job())
		//err = Unpack(bucket, ctx.Request().Body)
		if err != nil {
			v.log.Error(err)
			ctx.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		ctx.JSON("ok")
	})

	v1.Post("/v1/jobs", func(c api.Context) {
		req := &metav1.JobCreate{}

		if err := c.Bind(req); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		if err := req.Validate(); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeValidation,
			})
			return
		}

		if id, err := v.store.Job().InsertOne(context.TODO(), req); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		} else {
			c.JSON(map[string]string{
				"id": id,
			})
		}
	})

	v1.Get("/v1/jobs", func(c api.Context) {
		if jobs, err := v.store.Job().FindAll(context.TODO()); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		} else {
			c.JSON(jobs)
		}
	})

	v1.Get("/v1/jobs/{id}", func(c api.Context) {
		id := c.Param("id")

		if job, err := v.store.Job().FindOneByID(context.TODO(), id); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		} else {
			c.JSON(job)
		}
	})

	v1.Patch("/v1/jobs/{id}", func(c api.Context) {
		req := &metav1.JobPatch{}

		if err := c.Bind(req); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		id := c.Param("id")

		if job, err := v.store.Job().FindOneByID(context.TODO(), id); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		} else {
			req.ApplyJob(job)

			if err := req.Validate(); err != nil {
				v.log.Error(err)
				c.BadRequest().JSON(&APIError{
					Type: ErrorTypeValidation,
				})
				return
			}
		}

		if err := v.store.Job().PatchOneByID(context.TODO(), id, req); err != nil {
			v.log.Error(err)
			c.InternalError().JSON(&APIError{
				Type: ErrorTypeInternal,
			})
			return
		}

		c.JSON("ok")
	})

	v1.Get("/v1/linker", func(ctx api.Context) {
		nodes, err := v.store.Linker().FindAll(ctx.RequestContext())
		if err != nil {
			v.log.Error(err)
			ctx.BadRequest()
			return
		}

		if nodes == nil {
			nodes = []*metav1.LinkNode{}
		}

		ctx.JSON(nodes)
	})

	v.log.Info("listening on: ", addr)
	return http.ListenAndServe(addr, v1.Handler())
}

func (v *v1) Store() store.Repository {
	return v.store
}

func (v *v1) schedulerRetry(f func() error) error {
	return backoff.Retry(f, v.schedulerBackoff)
}
