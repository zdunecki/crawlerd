package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: improvements - streaming etc.

// TODO: cli is part of bigcrawl, not crawlerd
func main() {
	var rootCmd = &cobra.Command{Use: "bc"}

	var cmdApplyFiles string

	httpClient := &http.Client{
		Timeout: time.Second * 150,
	}

	apiURL := os.Getenv("API_URL")

	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	var cmdApply = &cobra.Command{
		Use:   "apply",
		Short: "",
		Long:  "",
		//Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			fmt.Println(pwd)

			// TODO: json support
			viper.SetConfigType("yaml")
			viper.SetConfigName("bigcrawl")
			viper.AddConfigPath(pwd)
			if err := viper.ReadInConfig(); err != nil {
				panic(err)
			}

			// TODO: send zip to server
			// TODO: token authentication
			// TODO: mask tokens if passed as a value in zipped files
			token := viper.GetString("token")
			if !strings.HasPrefix(token, "$") {
				panic("token should be passed as environment variable")
			}
			token = os.Getenv(token[1:])

			fmt.Println("token:", token)

			applyRoot := path.Join(pwd, cmdApplyFiles)
			w := new(bytes.Buffer)
			if err := Tar(applyRoot, w); err != nil {
				panic(err)
			}

			resp, err := httpClient.Post(apiURL+"/v1/cmd/apply", "application/zip, application/octet-stream", w)
			if err != nil {
				panic(err)
			}

			fmt.Println(resp.StatusCode)
			// TODO: server should unzip and make executable structure
			// currently we need to know where is entry file
			//
			//folderConfig := make(map[string]*viper.Viper)
			//err = filepath.Walk(cmdApplyFiles, func(p string, info os.FileInfo, err error) error {
			//	// ignore config files
			//	if strings.Contains(p, "bigcrawl.yaml") {
			//		return nil
			//	}
			//
			//	if info.IsDir() {
			//		v := viper.New()
			//
			//		f, err := os.Open(path.Join(pwd, p))
			//		if err != nil {
			//			panic(err)
			//		}
			//
			//		if err := v.ReadConfig(f); err != nil {
			//			panic(err)
			//		}
			//
			//		folderConfig[p] = v
			//
			//		return nil
			//	}
			//
			//	if err != nil {
			//		return err
			//	}
			//
			//	pathDir, f := path.Split(p)
			//
			//	fmt.Println(pathDir, f, folderConfig[pathDir].Get("entry"))
			//	return nil
			//})
			//
			//if err != nil {
			//	panic(err)
			//}
		},
	}

	cmdApply.Flags().StringVarP(&cmdApplyFiles, "files", "f", ".", "files to apply")

	rootCmd.AddCommand(cmdApply)
	rootCmd.Execute()
}
