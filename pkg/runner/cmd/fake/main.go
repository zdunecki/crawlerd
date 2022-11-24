package main

import (
	"flag"

	v1 "github.com/zdunecki/crawlerd/pkg/runner/api/runnerv1"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":9998", "api address")
	flag.Parse()

	v1.New(addr, newFakeStorage(), v1.Config{}).ListenAndServe()
}
