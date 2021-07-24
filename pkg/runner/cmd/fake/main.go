package main

import (
	"flag"

	v1 "crawlerd/pkg/runner/api/v1"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":9998", "api address")
	flag.Parse()

	v1.New(addr, newFakeStorage()).ListenAndServe()
}
