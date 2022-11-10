package worker

import (
	"bytes"
	"github.com/oxffaa/gopher-parse-sitemap"
	"io"
	"net/http"
)

type ExtracterAPIResponses interface {
	*ExtracterBlogSitemapResponse | *ExtracterBlogResponse
}

type ExtracterAPI[R ExtracterAPIResponses] interface {
	Extract(io.Reader) (R, error)
}

// blog sitemap extracter
type ExtracterBlogSitemapResponse struct {
	Error error
	Links []string // TODO: links stream / cursor
}

func NewExtracterBlogSitemap() ExtracterAPI[*ExtracterBlogSitemapResponse] {
	return &ExtracterBlogSitemap{}
}

type ExtracterBlogSitemap struct {
}

// TODO: streaming
func (ebs *ExtracterBlogSitemap) Extract(reader io.Reader) (*ExtracterBlogSitemapResponse, error) {
	// TODO: use Parse and ParseIndex because it's separate method in this lib - find better solution

	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)

	hasIndex := false

	links := make([]string, 0)

	if err := sitemap.ParseIndex(tee, func(entry sitemap.IndexEntry) error {
		hasIndex = true
		entry.GetLocation()
		resp, err := http.Get(entry.GetLocation())
		if err != nil {
			return nil
		}

		sitemap.Parse(resp.Body, func(entry sitemap.Entry) error {
			links = append(links, entry.GetLocation())
			return nil
		})

		return nil
	}); err != nil {
		return &ExtracterBlogSitemapResponse{
			Error: err,
		}, err
	}

	if !hasIndex {
		if err := sitemap.Parse(bytes.NewReader(buf.Bytes()), func(entry sitemap.Entry) error {
			links = append(links, entry.GetLocation())
			return nil
		}); err != nil {
			return &ExtracterBlogSitemapResponse{
				Error: err,
			}, err
		}
	}

	return &ExtracterBlogSitemapResponse{
		Links: links,
	}, nil
}

// blog extracter
type ExtracterBlogSpec struct {
	IgnoreSiteMap bool
}

type ExtracterBlogResponse struct {
	Error error
}

// TODO: options like regexp etc.
func NewExtracterBlog(spec *ExtracterBlogSpec) ExtracterAPI[*ExtracterBlogResponse] {
	return &ExtracterBlog{
		spec: spec,
	}
}

type ExtracterBlog struct {
	spec *ExtracterBlogSpec
}

func (eb *ExtracterBlog) Extract(reader io.Reader) (*ExtracterBlogResponse, error) {
	return &ExtracterBlogResponse{}, nil
}
