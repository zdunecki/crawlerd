package v1

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"crawlerd/api"
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func Untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	configFiles := make(map[string]*viper.Viper)
	content2 := make(map[string]map[string][]byte)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			for dir, _ := range content2 {
				findConfig := func() *viper.Viper {
					if _, ok := configFiles[dir]; ok {
						return configFiles[dir]
					} else {
						movementDir := dir

						for {
							if movementDir == "" || movementDir == "." {
								break
							}
							movementDir = filepath.Join(movementDir, "../")
							if _, ok := configFiles[movementDir]; ok {
								return configFiles[movementDir]
							}
						}
					}

					return nil
				}

				cfg := findConfig()

				if cfg == nil {
					continue
				}
				jobID := cfg.GetString("job_id")
				entry := cfg.GetString("entry")

				fmt.Println(jobID, entry)

				// TODO: insert files to API based on config
				//
				//for filePath, fileB := range fileMap {
				//
				//}
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
		target := filepath.Join(dst, header.Name)
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

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if err := createDirIfNotExists(target); err != nil {
				return err
			}

		// if it's a file create it
		case tar.TypeReg:
			//f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			dir, _ := path.Split(target)

			if err := createDirIfNotExists(dir); err != nil {
				return err
			}

			f, err := os.Create(target)
			if err != nil {
				return err
			}

			b, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}

			if _, ok := content2[targetDir]; ok {
				content2[targetDir][header.Name] = b
			} else {
				content2[targetDir] = make(map[string][]byte)
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
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

		err = Untar(path.Join(wd, "OUTPUT"), ctx.Request().Body)
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
