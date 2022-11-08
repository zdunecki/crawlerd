package worker

import "io"

type ExtracterAPI interface {
	ExtractBytes([]byte) error

	Extract(reader io.Reader) error
}

type ExtracterBlog struct {
}

type Extracter struct {
}
