package worker

import (
	"bytes"
	"fmt"
	"github.com/oxffaa/gopher-parse-sitemap"
	"io"
	"net/http"
)

type ExtracterAPI interface {
	Extract(io.Reader) (interface{}, error)
}

// blog sitemap extracter
type ExtracterBlogSitemapResponse struct {
	Links []string // TODO: links stream / cursor
}

type ExtracterBlogSitemap struct {
	Blog *ExtracterBlog
}

func (ebs *ExtracterBlogSitemap) Extract(reader io.Reader) (interface{}, error) {
	// TODO: use Parse and ParseIndex because it's separate method in this lib - find better solution

	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)

	hasIndex := false

	if err := sitemap.ParseIndex(tee, func(entry sitemap.IndexEntry) error {
		hasIndex = true
		entry.GetLocation()
		resp, err := http.Get(entry.GetLocation())
		if err != nil {
			return nil
		}

		sitemap.Parse(resp.Body, func(entry sitemap.Entry) error {
			fmt.Println(entry.GetLocation(), "each")
			return nil
		})

		fmt.Println(entry.GetLocation(), "index")
		return nil
	}); err != nil {
		return nil, err
	}

	if !hasIndex {
		if err := sitemap.Parse(bytes.NewReader(buf.Bytes()), func(entry sitemap.Entry) error {
			fmt.Println(entry.GetLocation(), "each2")
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return &ExtracterBlogSitemapResponse{}, nil
}

// blog extracter
type ExtracterBlog struct {
	Category string
}

func (eb *ExtracterBlog) Extract(reader io.Reader) (interface{}, error) {
	return &ExtracterBlogSitemapResponse{}, nil
}
