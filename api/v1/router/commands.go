package router

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	esbuild "github.com/evanw/esbuild/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zdunecki/crawlerd/api"
	"github.com/zdunecki/crawlerd/api/v1/objects"
	metav1 "github.com/zdunecki/crawlerd/pkg/meta/metav1"
	"github.com/zdunecki/crawlerd/pkg/store"
	"gocloud.dev/blob"
)

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

// TODO: refactor
func unpack(bucketName string, bucket *blob.Bucket, dstDir string, r io.Reader, jobRepo store.Job) error {
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

func (r *router) commandsApply(c api.Context) {
	wd, err := os.Getwd()
	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	err = unpack(r.storageBucketName, r.storageBucket, path.Join(wd, "OUTPUT"), c.Request().Body, r.store.Job())
	//err = Unpack(bucket, ctx.Request().Body)
	if err != nil {
		r.log.Error(err)
		c.InternalError().JSON(&objects.APIError{
			Type: objects.ErrorTypeInternal,
		})
		return
	}

	c.JSON("ok")
}
